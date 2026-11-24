package pilot

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type demoEvent struct {
	ID            string  `json:"id"`
	Project       string  `json:"project"`
	Kind          string  `json:"kind"`
	Title         string  `json:"title"`
	Detail        string  `json:"detail"`
	Amount        int64   `json:"amountPaise"`
	Ref           string  `json:"ref"`
	Time          string  `json:"time"`
	Photo         string  `json:"photo,omitempty"`
	PhotoHash     string  `json:"photoHash,omitempty"`
	Location      bool    `json:"location,omitempty"`
	Latitude      float64 `json:"latitude,omitempty"`
	Longitude     float64 `json:"longitude,omitempty"`
	Accuracy      float64 `json:"accuracy,omitempty"`
	Capture       string  `json:"capture,omitempty"`
	Duplicate     bool    `json:"duplicate,omitempty"`
	Invoice       string  `json:"invoice,omitempty"`
	Unit          string  `json:"unit,omitempty"`
	QuantityMilli int64   `json:"quantityMilli,omitempty"`
	RatePaise     int64   `json:"ratePaise,omitempty"`
	State         string  `json:"state,omitempty"`
	District      string  `json:"district,omitempty"`
	Category      string  `json:"category,omitempty"`
	StartPoint    string  `json:"startPoint,omitempty"`
	EndPoint      string  `json:"endPoint,omitempty"`
	StartDate     string  `json:"startDate,omitempty"`
	EndDate       string  `json:"endDate,omitempty"`
	Progress      int     `json:"progress,omitempty"`
	Source        string  `json:"source,omitempty"`
	TestUnit      string  `json:"testUnit,omitempty"`
	TestLimit     int64   `json:"testLimitMilli,omitempty"`
	TestValue     int64   `json:"testValueMilli,omitempty"`
	TestRule      string  `json:"testRule,omitempty"`
	Sample        string  `json:"sample,omitempty"`
}

var stages = []string{"sanction", "tender", "award", "work_order", "construction", "inspection", "completion", "maintenance"}

func demoSeed() []demoEvent {
	events := []demoEvent{
		{ID: "seed-1", Project: "demo-bridge", Kind: "sanction", Title: "Aravalli demonstration bridge", Detail: "Fictional route: Village A to Village B · 180 metres · Demo public works division", Amount: 10000000000, Time: "2026-11-01T09:00:00Z"},
		{ID: "seed-2", Project: "demo-bridge", Kind: "tender", Title: "Open tender", Detail: "180 metre bridge and approach road. Illustrative estimate.", Amount: 9800000000, Time: "2026-11-03T09:00:00Z"},
		{ID: "seed-3", Project: "demo-bridge", Kind: "bid", Title: "Illustrative contractor A", Detail: "Quotation for entire contract", Amount: 9200000000, Time: "2026-11-05T09:00:00Z"},
		{ID: "seed-4", Project: "demo-bridge", Kind: "award", Title: "Contract awarded", Detail: "Illustrative contractor A · 240 days", Amount: 9200000000, Time: "2026-11-08T09:00:00Z"},
		{ID: "seed-5", Project: "demo-bridge", Kind: "work_order", Title: "Work order issued", Detail: "Start 10 November · planned completion 8 July", Time: "2026-11-10T09:00:00Z"},
		{ID: "seed-6", Project: "demo-bridge", Kind: "construction", Title: "Pier foundations underway", Detail: "Pier 2 concrete work in progress", Time: "2026-11-15T09:00:00Z"},
		{ID: "seed-7", Project: "demo-bridge", Kind: "delivery", Title: "Concrete delivery", Detail: "120 m³ · delivery note DEMO-01", Time: "2026-11-16T09:00:00Z"},
		{ID: "seed-8", Project: "demo-bridge", Kind: "bill", Title: "Pier 2 concrete bill", Detail: "120 m³ × ₹6,000", Amount: 72000000, Time: "2026-11-17T09:00:00Z"},
		{ID: "seed-9", Project: "demo-bridge", Kind: "certificate", Title: "Certified Pier 2 measurement", Detail: "100 m³ × ₹6,000", Amount: 60000000, Ref: "seed-8", Time: "2026-11-18T09:00:00Z"},
		{ID: "seed-10", Project: "demo-bridge", Kind: "payment", Title: "Certified payment recorded", Detail: "Illustrative payment voucher", Amount: 60000000, Ref: "seed-8", Time: "2026-11-19T09:00:00Z"},
	}
	events[0].State = "Rajasthan"
	events[0].District = "Udaipur"
	events[0].Category = "Bridge"
	events[0].StartPoint = "Demo Village A"
	events[0].EndPoint = "Demo Village B"
	events[0].StartDate = "2026-11-10"
	events[0].EndDate = "2027-07-08"
	events[3].Ref = "seed-3"
	events[7].Invoice = "DEMO-BILL-01"
	events[7].Ref = "seed-item"
	events[7].Unit = "m3"
	events[7].QuantityMilli = 120000
	events[7].RatePaise = 600000
	events[8].Unit = "m3"
	events[8].QuantityMilli = 100000
	events[8].RatePaise = 600000
	item := demoEvent{ID: "seed-item", Project: "demo-bridge", Kind: "work_item", Title: "Pier 2 concrete", Detail: "Illustrative bill of quantities", Unit: "m3", QuantityMilli: 120000, RatePaise: 600000, Amount: 72000000, Time: "2026-11-12T09:00:00Z"}
	events = append(append(append([]demoEvent{}, events[:5]...), item), events[5:]...)
	events = append(events,
		demoEvent{ID: "seed-progress", Project: "demo-bridge", Kind: "progress", Title: "Foundation milestone update", Detail: "Fictional contractor-reported progress; approach road still pending.", Progress: 35, Time: "2026-11-20T09:00:00Z"},
		demoEvent{ID: "seed-quality", Project: "demo-bridge", Kind: "quality_test", Title: "Pier 2 concrete strength", Detail: "Illustrative laboratory sheet LAB-DEMO-17; follow-up measurement requested.", Ref: "seed-item", Sample: "LAB-DEMO-17", TestUnit: "MPa", TestRule: "minimum", TestLimit: 30000, TestValue: 28000, Time: "2026-11-21T09:00:00Z"},
		demoEvent{ID: "seed-report", Project: "demo-bridge", Kind: "report", Title: "Unfinished approach road", Detail: "Fictional resident observation: uneven approach surface needs inspection.", Time: "2026-11-22T09:00:00Z"},
	)
	// A completed fictional school project makes the entire lifecycle available at first launch.
	school := []demoEvent{
		{ID: "school-sanction", Kind: "sanction", Title: "Aravalli school renewal", Detail: "Fictional renovation of a rural school building.", Amount: 2500000000, State: "Rajasthan", District: "Udaipur", Category: "School", StartPoint: "School entrance", EndPoint: "Classroom wing", StartDate: "2026-01-12", EndDate: "2026-09-30"},
		{ID: "school-tender", Kind: "tender", Title: "School renewal tender", Amount: 2450000000},
		{ID: "school-bid", Kind: "bid", Title: "Illustrative contractor B", Amount: 2400000000},
		{ID: "school-award", Kind: "award", Title: "Contract award", Amount: 2400000000, Ref: "school-bid"},
		{ID: "school-order", Kind: "work_order", Title: "Work order and classroom schedule"},
		{ID: "school-start", Kind: "construction", Title: "Classroom renovation started"},
		{ID: "school-item", Kind: "work_item", Title: "Renovation package", Unit: "package", QuantityMilli: 1000, RatePaise: 1000000000, Amount: 1000000000},
		{ID: "school-bill", Kind: "bill", Title: "Renovation package bill", Ref: "school-item", Invoice: "SCHOOL-DEMO-01", Unit: "package", QuantityMilli: 1000, RatePaise: 1000000000, Amount: 1000000000},
		{ID: "school-cert", Kind: "certificate", Title: "Accepted package measurement", Ref: "school-bill", Unit: "package", QuantityMilli: 1000, RatePaise: 1000000000, Amount: 1000000000},
		{ID: "school-payment", Kind: "payment", Title: "Package payment recorded", Ref: "school-bill", Amount: 1000000000},
		{ID: "school-progress", Kind: "progress", Title: "All scheduled physical works reported complete", Progress: 100},
		{ID: "school-inspect", Kind: "inspection", Title: "Completion inspection recorded", Detail: "Illustrative inspection record"},
		{ID: "school-complete", Kind: "completion", Title: "School handed over"},
		{ID: "school-maintain", Kind: "maintenance", Title: "Maintenance period started", Detail: "Illustrative maintenance record; remaining contract payment documents unlisted."},
	}
	for i, event := range school {
		event.Project = "demo-school"
		event.Time = time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC).AddDate(0, 0, i*20).Format(time.RFC3339)
		events = append(events, event)
	}
	return events
}

func validateDemo(events []demoEvent, e demoEvent) error {
	if strings.TrimSpace(e.ID) == "" || len(e.ID) > 100 || strings.TrimSpace(e.Project) == "" || len(e.Project) > 100 || strings.TrimSpace(e.Title) == "" || len(e.Title) > 160 || len(e.Detail) > 4000 || e.Amount < 0 || e.Amount > 100000000000000 {
		return fmt.Errorf("check the title, project, details and amount")
	}
	if len(e.Invoice) > 120 || len(e.Unit) > 32 || e.QuantityMilli < 0 || e.QuantityMilli > 1000000000 || e.RatePaise < 0 || e.RatePaise > 1000000000 {
		return fmt.Errorf("invalid bill or measurement fields")
	}
	if e.Progress < 0 || e.Progress > 100 || (e.Progress != 0 && e.Kind != "progress") {
		return fmt.Errorf("progress must be a percentage between 0 and 100")
	}
	if e.Source != "" {
		u, err := url.Parse(e.Source)
		if len(e.Source) > 2048 || err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil {
			return fmt.Errorf("source must be an HTTPS document link")
		}
	}
	if e.Kind == "quality_test" {
		if e.TestUnit == "" || len(e.TestUnit) > 32 || strings.TrimSpace(e.Sample) == "" || len(e.Sample) > 120 || (e.TestRule != "minimum" && e.TestRule != "maximum") || e.TestLimit < 0 || e.TestLimit > 1000000000 || e.TestValue < 0 || e.TestValue > 1000000000 || (e.Source == "" && strings.TrimSpace(e.Detail) == "") {
			return fmt.Errorf("enter sample, test unit, limit, measured result and supporting test details")
		}
	} else if e.TestUnit != "" || e.TestLimit != 0 || e.TestValue != 0 || e.TestRule != "" || e.Sample != "" {
		return fmt.Errorf("quality fields belong to a test result")
	}
	if e.State != "" || e.District != "" || e.Category != "" || e.StartPoint != "" || e.EndPoint != "" || e.StartDate != "" || e.EndDate != "" {
		if e.Kind != "sanction" || len(e.State) > 100 || len(e.District) > 100 || len(e.Category) > 80 || len(e.StartPoint) > 160 || len(e.EndPoint) > 160 {
			return fmt.Errorf("invalid project location or scope")
		}
		if e.StartDate != "" {
			if _, err := time.Parse("2006-01-02", e.StartDate); err != nil {
				return fmt.Errorf("invalid start date")
			}
		}
		if e.EndDate != "" {
			if _, err := time.Parse("2006-01-02", e.EndDate); err != nil || (e.StartDate != "" && e.EndDate < e.StartDate) {
				return fmt.Errorf("invalid planned end date")
			}
		}
	}
	if e.QuantityMilli > 0 || e.RatePaise > 0 || e.Unit != "" || e.Invoice != "" {
		if e.Kind != "bill" && e.Kind != "certificate" && e.Kind != "delivery" && e.Kind != "work_item" && e.Kind != "variation" {
			return fmt.Errorf("quantity fields belong to bills, deliveries or certificates")
		}
		if strings.TrimSpace(e.Unit) == "" || e.QuantityMilli <= 0 {
			return fmt.Errorf("enter quantity and unit")
		}
		if e.Kind != "delivery" && e.RatePaise <= 0 {
			return fmt.Errorf("enter a contract rate")
		}
	}
	current := -1
	var bill, certified, paid int64
	var billEvent demoEvent
	var referenced demoEvent
	var measured int64
	var report string
	for _, previous := range events {
		if previous.Project != e.Project {
			continue
		}
		if previous.ID == e.Ref {
			referenced = previous
		}
		for i, stage := range stages {
			if previous.Kind == stage {
				current = i
			}
		}
		if previous.ID == e.Ref && previous.Kind == "bill" {
			bill = previous.Amount
			billEvent = previous
		}
		if previous.Ref == e.Ref && previous.Kind == "certificate" {
			certified += previous.Amount
			measured += previous.QuantityMilli
		}
		if previous.Ref == e.Ref && previous.Kind == "payment" {
			paid += previous.Amount
		}
		if previous.ID == e.Ref && previous.Kind == "report" {
			report = "submitted"
		}
		if previous.Ref == e.Ref && report != "" {
			switch previous.Kind {
			case "response", "acknowledge":
				report = "acknowledged"
			case "resolve":
				report = "resolved"
			case "reopen":
				report = "submitted"
			}
		}
	}
	for i, stage := range stages {
		if e.Kind == stage {
			if i != current+1 {
				return fmt.Errorf("complete the previous stage first")
			}
			if i < 3 && e.Amount <= 0 {
				return fmt.Errorf("enter the sanctioned, tender or contract amount")
			}
			if e.Kind == "award" && (referenced.Kind != "bid" || referenced.Amount != e.Amount) {
				return fmt.Errorf("choose a recorded quotation matching the award amount")
			}
			return nil
		}
	}
	switch e.Kind {
	case "quality_test":
		if current >= 4 && referenced.Kind == "work_item" {
			return nil
		}
	case "bid":
		if current == 1 && e.Amount > 0 {
			return nil
		}
	case "delivery", "progress":
		if current >= 4 && current < 6 {
			return nil
		}
	case "work_item":
		if current >= 3 && current < 7 && e.QuantityMilli > 0 && e.RatePaise > 0 && e.Amount == (e.QuantityMilli*e.RatePaise+500)/1000 {
			return nil
		}
	case "variation":
		if current >= 3 && current < 7 && referenced.Kind == "work_item" && e.Unit == referenced.Unit && e.RatePaise == referenced.RatePaise && e.QuantityMilli > 0 && e.Amount == (e.QuantityMilli*e.RatePaise+500)/1000 && strings.TrimSpace(e.Detail) != "" {
			return nil
		}
	case "bill":
		if current >= 4 && current < 7 && e.Amount > 0 {
			if e.Ref != "" && (referenced.Kind != "work_item" || e.Unit != referenced.Unit || e.RatePaise != referenced.RatePaise || e.QuantityMilli <= 0) {
				return fmt.Errorf("bill must match the selected contract item unit and rate")
			}
			return nil
		}
	case "certificate":
		if bill > 0 && e.Amount > 0 && e.Amount <= bill-certified {
			if billEvent.QuantityMilli > 0 && (e.Unit != billEvent.Unit || e.RatePaise != billEvent.RatePaise || e.QuantityMilli <= 0 || e.QuantityMilli > billEvent.QuantityMilli-measured || e.Amount != (e.QuantityMilli*e.RatePaise+500)/1000) {
				return fmt.Errorf("certification must match the bill unit, rate, remaining quantity and measured amount")
			}
			return nil
		}
	case "payment":
		if bill > 0 && e.Amount > 0 && e.Amount <= certified-paid {
			return nil
		}
	case "report":
		if current >= 0 {
			return nil
		}
	case "news":
		if current >= 0 && e.Source != "" {
			return nil
		}
	case "response":
		if report == "submitted" || report == "acknowledged" {
			return nil
		}
	case "acknowledge":
		if report == "submitted" {
			return nil
		}
	case "resolve":
		if report == "acknowledged" && strings.TrimSpace(e.Detail) != "" {
			return nil
		}
	case "reopen":
		if report == "resolved" && strings.TrimSpace(e.Detail) != "" {
			return nil
		}
	}
	return fmt.Errorf("event requires an eligible stage, reference and amount")
}

func demoPhoto(raw string) ([]byte, string, error) {
	if raw == "" {
		return nil, "", nil
	}
	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil || len(data) > 400000 {
		return nil, "", fmt.Errorf("photo must be a JPEG or PNG under 400 KB")
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || config.Width < 1 || config.Height < 1 || config.Width > 1600 || config.Height > 1600 {
		return nil, "", fmt.Errorf("photo dimensions must be at most 1600 by 1600")
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", fmt.Errorf("invalid photo")
	}
	var clean bytes.Buffer
	if err = jpeg.Encode(&clean, img, &jpeg.Options{Quality: 80}); err != nil {
		return nil, "", err
	}
	sum := sha256.Sum256(clean.Bytes())
	return clean.Bytes(), hex.EncodeToString(sum[:]), nil
}

func registerDemo(mux *http.ServeMux, dir string) {
	var mu sync.Mutex
	path := filepath.Join(dir, "demo-lifecycle.json")
	photoDir := filepath.Join(dir, "demo-photos")
	allowed := func(r *http.Request) bool {
		host, _, _ := net.SplitHostPort(r.RemoteAddr)
		ip := net.ParseIP(host)
		return os.Getenv("NIRIKSHAN_DEMO") == "1" && ip != nil && ip.IsLoopback()
	}
	mux.HandleFunc("GET /api/demo/photo/{hash}", func(w http.ResponseWriter, r *http.Request) {
		if !allowed(r) {
			fail(w, 403, "Local demo only")
			return
		}
		hash := r.PathValue("hash")
		if _, err := hex.DecodeString(hash); len(hash) != 64 || err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Cache-Control", "no-store")
		http.ServeFile(w, r, filepath.Join(photoDir, hash+".jpg"))
	})
	mux.HandleFunc("/api/demo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		// Unauthenticated demonstration writes are restricted to the local machine.
		if !allowed(r) {
			fail(w, 403, "Start the local server with NIRIKSHAN_DEMO=1")
			return
		}
		mu.Lock()
		defer mu.Unlock()
		events := demoSeed()
		b, err := os.ReadFile(path)
		if err == nil {
			if json.Unmarshal(b, &events) != nil {
				fail(w, 500, "Demo history is unreadable")
				return
			}
			// Introduce new fictional sample projects without rewriting an existing project's history.
			existingProjects := map[string]bool{}
			for _, event := range events {
				existingProjects[event.Project] = true
			}
			for _, event := range demoSeed() {
				if !existingProjects[event.Project] {
					events = append(events, event)
				}
			}
		} else if !os.IsNotExist(err) {
			fail(w, 500, "Demo history is unavailable")
			return
		}
		switch r.Method {
		case http.MethodGet:
			if project := r.URL.Query().Get("snapshot"); project != "" {
				snapshot, err := signDemoSnapshot(dir, project, events)
				if err != nil {
					fail(w, 503, "Could not sign project snapshot")
					return
				}
				reply(w, 200, snapshot)
				return
			}
			reply(w, 200, events)
		case http.MethodPost:
			var e demoEvent
			if strings.TrimSpace(strings.Split(r.Header.Get("Content-Type"), ";")[0]) != "application/json" {
				fail(w, 415, "Expected application/json")
				return
			}
			dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 600000))
			dec.DisallowUnknownFields()
			if dec.Decode(&e) != nil || dec.Decode(&struct{}{}) != io.EOF {
				fail(w, 400, "Invalid or oversized report")
				return
			}
			if e.PhotoHash != "" || e.Duplicate || (e.Location && (e.Latitude < -90 || e.Latitude > 90 || e.Longitude < -180 || e.Longitude > 180 || e.Accuracy < 0 || e.Accuracy > 100000)) {
				fail(w, 422, "Invalid evidence metadata")
				return
			}
			if (e.Photo != "" || e.Location || e.Capture != "") && e.Kind != "report" && e.Kind != "resolve" {
				fail(w, 422, "Evidence belongs to reports or resolutions")
				return
			}
			photo, hash, err := demoPhoto(e.Photo)
			if err != nil {
				fail(w, 422, err.Error())
				return
			}
			e.Photo = ""
			e.PhotoHash = hash
			if !e.Location {
				e.Latitude = 0
				e.Longitude = 0
				e.Accuracy = 0
			}
			if e.Location {
				e.Latitude = math.Round(e.Latitude*1000) / 1000
				e.Longitude = math.Round(e.Longitude*1000) / 1000
			}
			if hash == "" {
				e.Capture = ""
			} else if e.Capture != "camera_requested" {
				e.Capture = "file"
			}
			for _, previous := range events {
				if previous.ID == e.ID {
					e.Time = previous.Time
					e.Duplicate = previous.Duplicate
					if e != previous {
						fail(w, 409, "Receipt already used by another event")
						return
					}
					reply(w, 200, previous)
					return
				}
			}
			if err := validateDemo(events, e); err != nil {
				fail(w, 422, err.Error())
				return
			}
			if hash != "" {
				for _, previous := range events {
					if previous.PhotoHash == hash {
						e.Duplicate = true
						break
					}
				}
				if err := os.MkdirAll(photoDir, 0700); err != nil {
					fail(w, 503, "Could not save photo")
					return
				}
				photoPath := filepath.Join(photoDir, hash+".jpg")
				if _, err := os.Stat(photoPath); os.IsNotExist(err) {
					f, err := os.CreateTemp(photoDir, "image-*")
					if err != nil {
						fail(w, 503, "Could not save photo")
						return
					}
					defer os.Remove(f.Name())
					_, err = f.Write(photo)
					if err == nil {
						err = f.Sync()
					}
					closeErr := f.Close()
					if err == nil {
						err = closeErr
					}
					if err == nil {
						err = os.Rename(f.Name(), photoPath)
					}
					if err != nil {
						fail(w, 503, "Could not save photo")
						return
					}
				} else if err != nil {
					fail(w, 503, "Could not access photo")
					return
				}
			}
			e.Time = time.Now().UTC().Format(time.RFC3339Nano)
			events = append(events, e)
			b, err = json.Marshal(events)
			if err != nil {
				fail(w, 500, "Could not encode demo history")
				return
			}
			f, err := os.CreateTemp(dir, "demo-*")
			if err != nil {
				fail(w, 503, "Could not save demo history")
				return
			}
			defer os.Remove(f.Name())
			_, err = f.Write(b)
			if err == nil {
				err = f.Sync()
			}
			closeErr := f.Close()
			if err == nil {
				err = closeErr
			}
			if err == nil {
				err = os.Rename(f.Name(), path)
			}
			if err != nil {
				fail(w, 503, "Could not save demo history")
				return
			}
			reply(w, 201, e)
		default:
			w.Header().Set("Allow", "GET, POST")
			fail(w, 405, "Method unavailable")
		}
	})
}
