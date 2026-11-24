package pilot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestReviewLifecycle(t *testing.T) {
	dir := t.TempDir()
	store, err := openStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	app := &App{store: store, token: "review", ids: map[string]bool{"one": true}, limits: map[string]window{}}
	h := app.routes(t.TempDir())
	call := func(method, path, data, auth string, want int) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, bytes.NewBufferString(data))
		req.Header.Set("Content-Type", "application/json")
		if auth != "" {
			req.Header.Set("Authorization", "Bearer "+auth)
		}
		res := httptest.NewRecorder()
		h.ServeHTTP(res, req)
		if res.Code != want {
			t.Fatalf("%s %s returned %d, wanted %d: %s", method, path, res.Code, want, res.Body.String())
		}
		return res
	}
	call("GET", "/api/review", "", "", 401)
	call("POST", "/api/evidence", `{"kind":"invalid"}`, "", 422)
	call("POST", "/api/evidence", `{"project":"one","kind":"bill","title":"Unexpected amount","url":"https://example.com","date":"2025-01-01","amountPaise":-1}`, "", 422)
	submission := call("POST", "/api/evidence", `{"project":"one","kind":"bill","title":"Bill 1","url":"https://example.com","date":"2025-01-01","amountPaise":100}`, "", 202)
	var receipt map[string]string
	json.Unmarshal(submission.Body.Bytes(), &receipt)
	call("GET", "/api/receipts/unknown", "", "", 404)
	if receipt["id"] == "" || !bytes.Contains(call("GET", "/api/receipts/"+receipt["id"], "", "", 200).Body.Bytes(), []byte(`"pending"`)) {
		t.Fatal("receipt did not identify the pending submission")
	}
	if body := call("GET", "/api/evidence?project=one", "", "", 200).Body.String(); body != "[]\n" {
		t.Fatal(body)
	}
	if len(store.list("", "pending")) != 1 {
		t.Fatal("missing pending record")
	}
	decision := fmt.Sprintf(`{"id":"%s","decision":"approve","reason":"Checked source"}`, receipt["id"])
	call("POST", "/api/review", decision, "review", 200)
	if !bytes.Contains(call("GET", "/api/receipts/"+receipt["id"], "", "", 200).Body.Bytes(), []byte(`"approve"`)) {
		t.Fatal("receipt did not update after review")
	}
	call("POST", "/api/review", decision, "review", 409)
	recovered, err := openStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(recovered.list("one", "approve")) != 1 || len(recovered.list("", "pending")) != 0 {
		t.Fatal("review did not persist")
	}
	path := filepath.Join(dir, "000000000001.event")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, bytes.Replace(b, []byte("Bill 1"), []byte("Bill 9"), 1), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = openStore(dir); err == nil {
		t.Fatal("accepted altered event")
	}
}
func TestConcurrentAppends(t *testing.T) {
	dir := t.TempDir()
	s, err := openStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			err := s.append(Event{Action: "pending", ID: id, Evidence: &Evidence{ID: id}})
			if err != nil {
				t.Error(err)
			}
		}(fmt.Sprint(i))
	}
	wg.Wait()
	recovered, err := openStore(dir)
	if err != nil || len(recovered.events) != 16 {
		t.Fatal("concurrent event loss", err)
	}
}
func TestRateLimit(t *testing.T) {
	a := App{limits: map[string]window{}}
	for i := 0; i < 12; i++ {
		if !a.allow("local") {
			t.Fatal("early limit")
		}
	}
	if a.allow("local") {
		t.Fatal("limit exceeded")
	}
}
