package pilot

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var workspaceID = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)
var workspaceActions = map[string]string{
	"sanction": "government", "tender": "government", "award": "government", "work_order": "government",
	"work_item": "government", "variation": "government", "payment": "government", "completion": "government",
	"maintenance": "government", "response": "government", "acknowledge": "government", "resolve": "government",
	"bid": "contractor", "construction": "contractor", "delivery": "contractor", "progress": "contractor", "bill": "contractor",
	"certificate": "engineer", "inspection": "engineer", "quality_test": "engineer",
	"report": "citizen", "reopen": "citizen", "news": "citizen",
}

type workspaceAccount struct {
	Name, Role string
	Projects   []string
	Salt, Hash []byte
}

func workspaceHash(password string, salt []byte) []byte {
	hash, err := pbkdf2.Key(sha256.New, password, salt, 600000, 32)
	if err != nil {
		panic(err)
	}
	return hash
}

func provisionWorkspaceAccount(dir string, args []string, input io.Reader) error {
	if len(args) < 2 || !workspaceID.MatchString(args[0]) {
		return fmt.Errorf("usage: account-add USER government|contractor|engineer|citizen [PROJECT_ID ...]")
	}
	a := workspaceAccount{Name: args[0], Role: args[1], Projects: args[2:]}
	switch a.Role {
	case "government":
		if len(a.Projects) != 0 {
			return fmt.Errorf("government accounts cover the workspace")
		}
	case "contractor", "engineer", "citizen":
		if len(a.Projects) == 0 {
			return fmt.Errorf("assign at least one project identifier")
		}
	default:
		return fmt.Errorf("unknown account role")
	}
	for _, p := range a.Projects {
		if !workspaceID.MatchString(p) {
			return fmt.Errorf("invalid project identifier %q", p)
		}
	}
	data, err := io.ReadAll(io.LimitReader(input, 259))
	if err != nil {
		return err
	}
	password := strings.TrimSuffix(strings.TrimSuffix(string(data), "\n"), "\r")
	if len(password) < 12 || len(password) > 256 || strings.ContainsAny(password, "\r\n") {
		return fmt.Errorf("supply a 12–256 byte password on stdin")
	}
	a.Salt = make([]byte, 32)
	if _, err = rand.Read(a.Salt); err != nil {
		return err
	}
	a.Hash = workspaceHash(password, a.Salt)
	data, err = json.Marshal(a)
	if err != nil {
		return err
	}
	accounts := filepath.Join(dir, "workspace", "accounts")
	if err = os.MkdirAll(accounts, 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(accounts, "account-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	_, err = f.Write(data)
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Link(f.Name(), filepath.Join(accounts, "user-"+a.Name+".json"))
}

type workspaceSession struct {
	account workspaceAccount
	expires time.Time
}
type workspaceAuth struct {
	dir       string
	secure    bool
	mu        sync.Mutex
	sessions  map[string]workspaceSession
	attempts  []time.Time
	hashSlots chan struct{}
}

func registerWorkspace(mux *http.ServeMux, data string, secure bool) error {
	dir := filepath.Join(data, "workspace")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	auth := &workspaceAuth{dir: dir, secure: secure, sessions: make(map[string]workspaceSession), hashSlots: make(chan struct{}, 4)}
	mux.HandleFunc("POST /api/workspace/session", auth.login)
	mux.HandleFunc("GET /api/workspace/session", func(w http.ResponseWriter, r *http.Request) {
		if a, ok := auth.require(w, r); ok {
			reply(w, 200, map[string]string{"name": a.Name, "role": a.Role})
		}
	})
	mux.HandleFunc("DELETE /api/workspace/session", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := auth.require(w, r); !ok {
			return
		}
		cookie, _ := r.Cookie(workspaceCookie)
		auth.mu.Lock()
		delete(auth.sessions, cookie.Value)
		auth.mu.Unlock()
		auth.cookie(w, "", -1)
		w.WriteHeader(http.StatusNoContent)
	})
	registerLifecycle(mux, dir, auth)
	return nil
}

const workspaceCookie = "nirikshan_workspace"
const workspaceSessionLife = 8 * time.Hour

func (auth *workspaceAuth) cookie(w http.ResponseWriter, value string, age int) {
	http.SetCookie(w, &http.Cookie{Name: workspaceCookie, Value: value, Path: "/api/workspace", HttpOnly: true,
		Secure: auth.secure, SameSite: http.SameSiteStrictMode, MaxAge: age})
}

func workspaceWrite(w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("X-Nirikshan-Workspace") != "1" {
		fail(w, 403, "Workspace request header required")
		return false
	}
	return true
}

func (auth *workspaceAuth) login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if !workspaceWrite(w, r) {
		return
	}
	var login struct{ Name, Password string }
	if !decode(w, r, &login) {
		return
	}
	if !workspaceID.MatchString(login.Name) || len(login.Password) > 256 {
		fail(w, 401, "Invalid username or password")
		return
	}
	now := time.Now()
	auth.mu.Lock()
	first := 0
	for first < len(auth.attempts) && now.Sub(auth.attempts[first]) > time.Minute {
		first++
	}
	auth.attempts = auth.attempts[first:]
	limited := len(auth.attempts) >= 30
	if !limited {
		auth.attempts = append(auth.attempts, now)
	}
	auth.mu.Unlock()
	select {
	case auth.hashSlots <- struct{}{}:
		defer func() { <-auth.hashSlots }()
	default:
		limited = true
	}
	if limited {
		fail(w, 429, "Too many sign-in attempts; try again in a minute")
		return
	}
	raw, err := os.ReadFile(filepath.Join(auth.dir, "accounts", "user-"+login.Name+".json"))
	var a workspaceAccount
	exists := err == nil && json.Unmarshal(raw, &a) == nil && a.Name == login.Name && len(a.Salt) == 32 && len(a.Hash) == 32
	if !exists {
		a.Salt, a.Hash = make([]byte, 32), make([]byte, 32)
	}
	if subtle.ConstantTimeCompare(workspaceHash(login.Password, a.Salt), a.Hash) != 1 || !exists {
		fail(w, 401, "Invalid username or password")
		return
	}
	var token [32]byte
	if _, err = rand.Read(token[:]); err != nil {
		fail(w, 503, "Could not start a session")
		return
	}
	id := hex.EncodeToString(token[:])
	auth.mu.Lock()
	for key, session := range auth.sessions {
		if !session.expires.After(now) {
			delete(auth.sessions, key)
		}
	}
	if previous, err := r.Cookie(workspaceCookie); err == nil {
		delete(auth.sessions, previous.Value)
	}
	auth.sessions[id] = workspaceSession{a, now.Add(workspaceSessionLife)}
	auth.mu.Unlock()
	auth.cookie(w, id, int(workspaceSessionLife.Seconds()))
	reply(w, 200, map[string]string{"name": a.Name, "role": a.Role})
}

func (auth *workspaceAuth) require(w http.ResponseWriter, r *http.Request) (workspaceAccount, bool) {
	w.Header().Set("Cache-Control", "no-store")
	if r.Method != http.MethodGet && r.Method != http.MethodHead && !workspaceWrite(w, r) {
		return workspaceAccount{}, false
	}
	if cookie, err := r.Cookie(workspaceCookie); err == nil {
		auth.mu.Lock()
		session, ok := auth.sessions[cookie.Value]
		auth.mu.Unlock()
		if ok && time.Now().Before(session.expires) {
			return session.account, true
		}
	}
	fail(w, 401, "Sign into the department workspace")
	return workspaceAccount{}, false
}

func (a workspaceAccount) canAccess(project string) bool {
	if a.Role == "government" {
		return true
	}
	for _, p := range a.Projects {
		if p == project {
			return true
		}
	}
	return false
}

func (a workspaceAccount) visible(events []demoEvent) []demoEvent {
	own := make(map[string]bool)
	for _, e := range events {
		if e.Kind == "report" && e.Actor == a.Name {
			own[e.ID] = true
		}
	}
	rows := make([]demoEvent, 0, len(events))
	for _, e := range events {
		if !a.canAccess(e.Project) || (e.Kind == "bid" && a.Role != "government" && e.Actor != a.Name) {
			continue
		}
		if a.Role != "government" {
			switch e.Kind {
			case "report", "news":
				if e.Actor != a.Name {
					continue
				}
			case "acknowledge", "response", "resolve", "reopen":
				if !own[e.Ref] {
					continue
				}
			}
		}
		if a.Role == "citizen" {
			switch e.Kind {
			case "sanction", "tender", "award", "work_order", "construction", "inspection", "completion", "maintenance", "report", "news", "acknowledge", "response", "resolve", "reopen":
			default:
				continue
			}
		}
		rows = append(rows, e)
	}
	return rows
}

func (a workspaceAccount) canRecord(events []demoEvent, e demoEvent) bool {
	if !a.canAccess(e.Project) || workspaceActions[e.Kind] != a.Role {
		return false
	}
	if a.Role == "contractor" && e.Kind != "bid" {
		for _, award := range events {
			if award.Project == e.Project && award.Kind == "award" {
				for _, bid := range events {
					if bid.ID == award.Ref && bid.Project == e.Project {
						return bid.Actor == a.Name
					}
				}
			}
		}
		return false
	}
	if e.Kind == "reopen" {
		for _, report := range events {
			if report.ID == e.Ref && report.Project == e.Project && report.Kind == "report" {
				return report.Actor == a.Name
			}
		}
		return false
	}
	return true
}

func auditCSV(w http.ResponseWriter, project string, events []demoEvent) {
	rows := make([]demoEvent, 0)
	for _, e := range events {
		if e.Project == project {
			rows = append(rows, e)
		}
	}
	if len(rows) == 0 {
		fail(w, 404, "Project unavailable")
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="nirikshan-`+project+`-audit.csv"`)
	writer := csv.NewWriter(w)
	writer.UseCRLF = true
	_, _ = io.WriteString(w, "\uFEFF")
	_ = writer.Write([]string{"Project", "Record ID", "Recorded UTC", "Account", "Action", "Title", "Details", "Amount (paise)", "Related record", "Invoice", "Unit", "Quantity (thousandths)", "Rate (paise)", "Source", "Photo SHA-256", "Progress (%)", "Sample", "Measurement unit", "Rule", "Limit (thousandths)", "Measured (thousandths)"})
	for _, e := range rows {
		_ = writer.Write([]string{e.Project, sheetText(e.ID), e.Time, e.Actor, e.Kind, sheetText(e.Title), sheetText(e.Detail),
			strconv.FormatInt(e.Amount, 10), sheetText(e.Ref), sheetText(e.Invoice), sheetText(e.Unit), strconv.FormatInt(e.QuantityMilli, 10),
			strconv.FormatInt(e.RatePaise, 10), sheetText(e.Source), e.PhotoHash, strconv.Itoa(e.Progress), sheetText(e.Sample),
			sheetText(e.TestUnit), e.TestRule, strconv.FormatInt(e.TestLimit, 10), strconv.FormatInt(e.TestValue, 10)})
	}
	writer.Flush()
}

func sheetText(value string) string {
	trimmed := strings.TrimLeft(value, " \r\n\t")
	if (trimmed != "" && strings.ContainsRune("=+-@", rune(trimmed[0]))) || strings.HasPrefix(value, "\t") || strings.HasPrefix(value, "\r") || strings.HasPrefix(value, "\n") {
		return "'" + value
	}
	return value
}
