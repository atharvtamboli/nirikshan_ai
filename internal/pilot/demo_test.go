package pilot

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDemoLifecycle(t *testing.T) {
	t.Setenv("NIRIKSHAN_DEMO", "1")
	dir := t.TempDir()
	setup := func() *http.ServeMux {
		mux := http.NewServeMux()
		registerDemo(mux, dir)
		return mux
	}
	mux := setup()
	send := func(e demoEvent, expected int) {
		t.Helper()
		b, _ := json.Marshal(e)
		r := httptest.NewRequest("POST", "/api/demo", bytes.NewReader(b))
		r.RemoteAddr = "127.0.0.1:12345"
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		if w.Code != expected {
			t.Fatalf("received %d, expected %d: %s", w.Code, expected, w.Body)
		}
	}
	send(demoEvent{ID: "bad", Project: "demo-bridge", Kind: "payment", Title: "Extra payment", Ref: "seed-8", Amount: 1}, 422)
	report := demoEvent{ID: "report-1", Project: "demo-bridge", Kind: "report", Title: "Approach needs repair"}
	send(report, 201)
	send(report, 200)
	send(demoEvent{ID: "response-1", Project: "demo-bridge", Kind: "response", Title: "Inspection scheduled", Ref: report.ID}, 201)
	send(demoEvent{ID: "resolve-1", Project: "demo-bridge", Kind: "resolve", Title: "Repair recorded", Detail: "Approach repaired", Ref: report.ID}, 201)
	send(demoEvent{ID: "reopen-1", Project: "demo-bridge", Kind: "reopen", Title: "Issue persists", Detail: "New damage observed", Ref: report.ID}, 201)
	var photo bytes.Buffer
	if err := png.Encode(&photo, image.NewRGBA(image.Rect(0, 0, 4, 4))); err != nil {
		t.Fatal(err)
	}
	photoEvent := demoEvent{ID: "photo-1", Project: "demo-bridge", Kind: "report", Title: "Photo observation", Photo: base64.StdEncoding.EncodeToString(photo.Bytes()), Location: true, Latitude: 24.123456, Longitude: 73.123456, Accuracy: 20}
	send(photoEvent, 201)
	send(photoEvent, 200)
	photoEvent.ID = "photo-2"
	send(photoEvent, 201)
	request := httptest.NewRequest("GET", "/api/demo", nil)
	request.RemoteAddr = "127.0.0.1:12345"
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	var history []demoEvent
	if err := json.Unmarshal(response.Body.Bytes(), &history); err != nil {
		t.Fatal(err)
	}
	last := history[len(history)-1]
	if !last.Duplicate || last.Photo != "" || last.PhotoHash == "" || last.Latitude != 24.123 {
		t.Fatalf("unexpected evidence metadata: %+v", last)
	}
	request = httptest.NewRequest("GET", "/api/demo/photo/"+last.PhotoHash, nil)
	request.RemoteAddr = "127.0.0.1:12345"
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != 200 || response.Header().Get("Content-Type") != "image/jpeg" {
		t.Fatalf("photo failed: %d", response.Code)
	}
	mux = setup()
	send(report, 200)
	send(demoEvent{ID: "sanction-1", Project: "demo-road", Kind: "sanction", Title: "Road project", Amount: 100000}, 201)
	send(demoEvent{ID: "award-1", Project: "demo-road", Kind: "award", Title: "Premature award", Amount: 90000}, 422)
	send(demoEvent{ID: "tender-1", Project: "demo-road", Kind: "tender", Title: "Tender", Amount: 95000}, 201)
	send(demoEvent{ID: "bid-1", Project: "demo-road", Kind: "bid", Title: "Quotation", Amount: 90000}, 201)
	send(demoEvent{ID: "award-2", Project: "demo-road", Kind: "award", Title: "Award without bid", Amount: 90000}, 422)
	send(demoEvent{ID: "award-3", Project: "demo-road", Kind: "award", Title: "Award", Amount: 90000, Ref: "bid-1"}, 201)
	send(demoEvent{ID: "variation-1", Project: "demo-bridge", Kind: "variation", Title: "Additional work", Detail: "Approved increase", Ref: "seed-item", Unit: "m3", QuantityMilli: 1000, RatePaise: 600000, Amount: 600000}, 201)
	send(demoEvent{ID: "bill-invalid", Project: "demo-bridge", Kind: "bill", Title: "Rate mismatch", Ref: "seed-item", Unit: "m3", QuantityMilli: 1000, RatePaise: 500000, Amount: 500000}, 422)
	test := demoEvent{ID: "quality-1", Project: "demo-bridge", Kind: "quality_test", Title: "Concrete cylinder result", Detail: "Illustrative lab sheet Q-01", Ref: "seed-item", Sample: "Q-01", TestUnit: "MPa", TestRule: "minimum", TestLimit: 30000, TestValue: 28000}
	send(test, 201)
	test.ID = "quality-2"
	test.Ref = "missing"
	send(test, 422)
	send(demoEvent{ID: "bill-2", Project: "demo-bridge", Kind: "bill", Title: "Structured bill", Invoice: "INV-02", Unit: "m3", QuantityMilli: 10000, RatePaise: 600000, Amount: 6000000}, 201)
	certificate := demoEvent{ID: "cert-2", Project: "demo-bridge", Kind: "certificate", Title: "Measurement", Ref: "bill-2", Unit: "m3", QuantityMilli: 8000, RatePaise: 600000, Amount: 4800000}
	send(certificate, 201)
	certificate.ID = "cert-3"
	send(certificate, 422)
}
