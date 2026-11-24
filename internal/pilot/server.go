package pilot

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Evidence struct {
	ID        string `json:"id"`
	Project   string `json:"project"`
	Kind      string `json:"kind"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Amount    *int64 `json:"amountPaise,omitempty"`
	Reference string `json:"reference"`
	Date      string `json:"date"`
}
type Event struct {
	Action   string    `json:"action"`
	ID       string    `json:"id"`
	Evidence *Evidence `json:"evidence,omitempty"`
	Reason   string    `json:"reason,omitempty"`
	Time     string    `json:"time"`
	Previous string    `json:"previous"`
	Hash     string    `json:"hash,omitempty"`
}
type Store struct {
	mu     sync.Mutex
	dir    string
	events []Event
	state  map[string]string
}

func digest(e Event) string {
	e.Hash = ""
	b, _ := json.Marshal(e)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
func (s *Store) apply(e Event) {
	s.events = append(s.events, e)
	s.state[e.ID] = e.Action
}
func (s *Store) head() string {
	if len(s.events) == 0 {
		return ""
	}
	return s.events[len(s.events)-1].Hash
}
func (s *Store) valid(e Event) bool {
	if e.ID == "" || e.Previous != s.head() || digest(e) != e.Hash {
		return false
	}
	if e.Action == "pending" {
		return e.Evidence != nil && e.Evidence.ID == e.ID && s.state[e.ID] == ""
	}
	return (e.Action == "approve" || e.Action == "reject") && s.state[e.ID] == "pending" && e.Evidence == nil
}
func openStore(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	s := &Store{dir: dir, state: map[string]string{}}
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".event" {
			continue
		}
		f := filepath.Join(dir, file.Name())
		var e Event
		b, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		if file.Name() != fmt.Sprintf("%012d.event", len(s.events)+1) || json.Unmarshal(b, &e) != nil || !s.valid(e) {
			return nil, fmt.Errorf("invalid event at %s", f)
		}
		s.apply(e)
	}
	return s, nil
}

var conflict = errors.New("record has already changed")

func (s *Store) append(e Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	e.Time = time.Now().UTC().Format(time.RFC3339Nano)
	e.Previous = s.head()
	e.Hash = digest(e)
	if !s.valid(e) {
		return conflict
	}
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(s.dir, "pending-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(b); err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	// A hard link publishes the synced file atomically and refuses an existing sequence number.
	target := filepath.Join(s.dir, fmt.Sprintf("%012d.event", len(s.events)+1))
	if err = os.Link(f.Name(), target); err != nil {
		return err
	}
	s.apply(e)
	return nil
}
func (s *Store) list(project, status string) []Evidence {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []Evidence{}
	for i := len(s.events) - 1; i >= 0; i-- {
		e := s.events[i]
		if e.Evidence != nil && s.state[e.ID] == status && (project == "" || e.Evidence.Project == project) {
			out = append(out, *e.Evidence)
		}
	}
	return out
}

type window struct {
	start time.Time
	count int
}
type App struct {
	store    *Store
	token    string
	projects []byte
	ids      map[string]bool
	mu       sync.Mutex
	limits   map[string]window
}

func (a *App) allow(key string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now()
	for k, v := range a.limits {
		if now.Sub(v.start) >= time.Minute {
			delete(a.limits, k)
		}
	}
	w := a.limits[key]
	if w.start.IsZero() {
		if len(a.limits) >= 4096 {
			return false
		}
		w.start = now
	}
	if w.count >= 12 {
		return false
	}
	w.count++
	a.limits[key] = w
	return true
}
func reply(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
func fail(w http.ResponseWriter, status int, msg string) {
	reply(w, status, map[string]string{"error": msg})
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	if strings.TrimSpace(strings.Split(r.Header.Get("Content-Type"), ";")[0]) != "application/json" {
		fail(w, 415, "Expected application/json")
		return false
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16384))
	dec.DisallowUnknownFields()
	if dec.Decode(v) != nil || dec.Decode(&struct{}{}) != io.EOF {
		fail(w, 400, "Invalid JSON or oversized request")
		return false
	}
	return true
}
func validEvidence(e Evidence) bool {
	validKind := map[string]bool{"bill": true, "payment": true, "inspection": true, "news": true, "observation": true, "award": true}
	if !validKind[e.Kind] || len(strings.TrimSpace(e.Title)) < 2 || len(e.Title) > 1600 || len(e.Reference) > 1024 || len(e.URL) > 2048 {
		return false
	}
	if _, err := time.Parse("2006-01-02", e.Date); err != nil {
		return false
	}
	if e.Amount != nil && (*e.Amount < 0 || *e.Amount > 9007199254740991 || (e.Kind != "bill" && e.Kind != "payment" && e.Kind != "award")) {
		return false
	}
	if e.URL == "" {
		return e.Kind == "observation"
	}
	u, err := url.ParseRequestURI(e.URL)
	return err == nil && u.Scheme == "https" && u.Hostname() != "" && u.User == nil
}
func (a *App) routes(web string) http.Handler {
	mux := http.NewServeMux()
	registerDemo(mux, a.store.dir)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { reply(w, 200, map[string]bool{"ok": true}) })
	mux.HandleFunc("GET /api/projects", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(a.projects)
	})
	mux.HandleFunc("GET /api/evidence", func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Query().Get("project")
		if p != "" && !a.ids[p] {
			fail(w, 404, "Unknown project")
			return
		}
		reply(w, 200, a.store.list(p, "approve"))
	})
	mux.HandleFunc("POST /api/evidence", func(w http.ResponseWriter, r *http.Request) {
		if a.token == "" {
			fail(w, 503, "Review queue is offline")
			return
		}
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		if !a.allow(ip) {
			fail(w, 429, "Please try again in one minute")
			return
		}
		var e Evidence
		if !decode(w, r, &e) {
			return
		}
		if e.ID != "" || !a.ids[e.Project] || !validEvidence(e) {
			fail(w, 422, "Check project, type, title, HTTPS source, amount and date")
			return
		}
		var id [16]byte
		if _, err := rand.Read(id[:]); err != nil {
			fail(w, 500, "Receipt unavailable")
			return
		}
		e.ID = hex.EncodeToString(id[:])
		if err := a.store.append(Event{Action: "pending", ID: e.ID, Evidence: &e}); err != nil {
			fail(w, 503, "Could not save submission")
			return
		}
		reply(w, 202, map[string]string{"id": e.ID, "status": "pending"})
	})
	mux.HandleFunc("GET /api/receipts/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		id := r.PathValue("id")
		a.store.mu.Lock()
		status := a.store.state[id]
		a.store.mu.Unlock()
		if status == "" {
			fail(w, 404, "Receipt unavailable")
			return
		}
		reply(w, 200, map[string]string{"id": id, "status": status})
	})
	mux.HandleFunc("/api/review", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if a.token == "" || subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte("Bearer "+a.token)) != 1 {
			fail(w, 401, "Reviewer authentication required")
			return
		}
		if r.Method == "GET" {
			reply(w, 200, a.store.list("", "pending"))
			return
		}
		if r.Method != "POST" {
			w.Header().Set("Allow", "GET, POST")
			fail(w, 405, "Method unavailable")
			return
		}
		var d struct{ ID, Decision, Reason string }
		if !decode(w, r, &d) {
			return
		}
		if (d.Decision != "approve" && d.Decision != "reject") || strings.TrimSpace(d.Reason) == "" || len(d.Reason) > 2048 {
			fail(w, 422, "A decision and reason are required")
			return
		}
		err := a.store.append(Event{Action: d.Decision, ID: d.ID, Reason: d.Reason})
		if errors.Is(err, conflict) {
			fail(w, 409, "No pending submission matches this ID")
			return
		}
		if err != nil {
			fail(w, 503, "Could not save review")
			return
		}
		reply(w, 200, map[string]string{"id": d.ID, "decision": d.Decision})
	})
	mux.Handle("/api/", http.NotFoundHandler())
	mux.Handle("/", http.FileServer(http.Dir(web)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		mux.ServeHTTP(w, r)
	})
}
func env(k, v string) string {
	if s := os.Getenv(k); s != "" {
		return s
	}
	return v
}
func Run() {
	web := env("WEB_DIR", "web")
	data, err := os.ReadFile(filepath.Join(web, "projects.json"))
	if err != nil {
		log.Fatal(err)
	}
	var projects []struct{ ID string }
	if err = json.Unmarshal(data, &projects); err != nil || len(projects) == 0 {
		log.Fatal("invalid project catalog")
	}
	s, err := openStore(env("DATA_DIR", "data"))
	if err != nil {
		log.Fatal(err)
	}
	a := &App{store: s, projects: data, token: os.Getenv("REVIEW_TOKEN"), ids: map[string]bool{}, limits: map[string]window{}}
	for _, p := range projects {
		if p.ID == "" || a.ids[p.ID] {
			log.Fatal("missing or duplicate project ID")
		}
		a.ids[p.ID] = true
	}
	srv := &http.Server{Addr: ":" + env("PORT", "8080"), Handler: a.routes(web), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	log.Printf("Nirikshan listening on %s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
