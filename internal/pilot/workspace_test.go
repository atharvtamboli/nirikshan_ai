package pilot

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWorkspaceProjectLifecycle(t *testing.T) {
	dir := t.TempDir()
	password := "a test-only secret"
	for _, args := range [][]string{
		{"department", "government"}, {"builder", "contractor", "bridge"}, {"other-bidder", "contractor", "bridge"},
		{"outside", "engineer", "elsewhere"}, {"inspector", "engineer", "bridge"},
		{"resident", "citizen", "bridge"}, {"resident-two", "citizen", "bridge"},
	} {
		if err := provisionWorkspaceAccount(dir, args, strings.NewReader(password)); err != nil {
			t.Fatal(err)
		}
	}
	create := func() *http.ServeMux {
		mux := http.NewServeMux()
		if err := registerWorkspace(mux, dir, true); err != nil {
			t.Fatal(err)
		}
		return mux
	}
	mux := create()
	call := func(method, path string, body any, cookie *http.Cookie, header bool, want int) *httptest.ResponseRecorder {
		t.Helper()
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		request := httptest.NewRequest(method, path, bytes.NewReader(raw))
		request.Header.Set("Content-Type", "application/json")
		if header {
			request.Header.Set("X-Nirikshan-Workspace", "1")
		}
		if cookie != nil {
			request.AddCookie(cookie)
		}
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		if response.Code != want {
			t.Fatalf("%s %s: got %d want %d: %s", method, path, response.Code, want, response.Body.String())
		}
		return response
	}
	session, projects := "/api/workspace/session", "/api/workspace/projects"
	login := func(name string) *http.Cookie {
		t.Helper()
		r := call("POST", session, map[string]string{"name": name, "password": password}, nil, true, 200)
		cookie := r.Result().Cookies()[0]
		if !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteStrictMode {
			t.Fatalf("unexpected session cookie: %v", cookie)
		}
		return cookie
	}
	event := func(kind, id string, amount int64, ref string) demoEvent {
		return demoEvent{Project: "bridge", Kind: kind, ID: id, Title: id, Amount: amount, Ref: ref}
	}
	record := func(cookie *http.Cookie, e demoEvent, want int) *httptest.ResponseRecorder {
		t.Helper()
		return call("POST", projects, e, cookie, true, want)
	}
	call("GET", projects, nil, nil, true, 401)
	call("POST", session, map[string]string{"name": "department", "password": password}, nil, false, 403)
	call("POST", session, map[string]string{"name": "department", "password": "wrong"}, nil, true, 401)
	gov, builder, other, external, engineer, citizen, observer := login("department"), login("builder"), login("other-bidder"), login("outside"), login("inspector"), login("resident"), login("resident-two")
	sanction := event("sanction", "sanction", 100000, "")
	record(builder, sanction, 403)
	record(gov, sanction, 201)
	record(gov, sanction, 200)
	record(gov, event("tender", "notice", 90000, ""), 201)
	record(gov, event("bid", "gov-bid", 90000, ""), 403)
	record(builder, event("bid", "builder-bid", 85000, ""), 201)
	record(other, event("bid", "other-bid", 88000, ""), 201)
	body := call("GET", projects, nil, builder, false, 200).Body.String()
	if strings.Contains(body, "other-bid") || !strings.Contains(body, "builder-bid") {
		t.Fatalf("bid isolation failed: %s", body)
	}
	call("GET", projects+"?snapshot=bridge", nil, external, false, 404)
	call("GET", projects+"?audit=bridge", nil, external, false, 404)
	record(gov, event("award", "award", 85000, "builder-bid"), 201)
	record(gov, event("work_order", "order", 0, ""), 201)
	record(other, event("construction", "start", 0, ""), 403)
	record(builder, event("construction", "start", 0, ""), 201)
	bill := event("bill", "bill", 15000, "")
	bill.Title = "=formula"
	record(builder, bill, 201)
	record(external, event("certificate", "cert", 12000, "bill"), 403)
	record(engineer, event("certificate", "cert", 12000, "bill"), 201)
	record(gov, event("payment", "excess", 12001, "bill"), 422)
	record(gov, event("payment", "payment", 12000, "bill"), 201)
	record(engineer, event("inspection", "inspection", 0, ""), 201)
	record(gov, event("completion", "completion", 0, ""), 201)
	record(gov, event("maintenance", "maintenance", 0, ""), 201)
	report := event("report", "report-one", 0, "")
	record(citizen, report, 201)
	body = call("GET", projects, nil, observer, false, 200).Body.String()
	if strings.Contains(body, "report-one") || strings.Contains(body, `"kind":"bill"`) {
		t.Fatalf("citizen visibility failed: %s", body)
	}
	record(gov, event("acknowledge", "ack", 0, "report-one"), 201)
	resolution := event("resolve", "resolution", 0, "report-one")
	resolution.Detail = "Issue resolved"
	record(gov, resolution, 201)
	reopen := event("reopen", "reopen", 0, "report-one")
	reopen.Detail = "Further review requested"
	record(observer, reopen, 403)
	record(citizen, reopen, 201)
	snapshot := call("GET", projects+"?snapshot=bridge", nil, gov, false, 200).Body
	if !strings.Contains(snapshot.String(), "nirikshan-workspace-v1") {
		t.Fatal(snapshot.String())
	}
	if _, err := VerifyEvidence(snapshot, ""); err != nil {
		t.Fatal(err)
	}
	audit := call("GET", projects+"?audit=bridge", nil, gov, false, 200)
	rows, err := csv.NewReader(strings.NewReader(strings.TrimPrefix(audit.Body.String(), "\uFEFF"))).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, row := range rows {
		if row[1] == "bill" && row[3] == "builder" && row[5] == "'=formula" {
			found = true
		}
	}
	if !found {
		t.Fatal("audit did not attribute and escape the submitted bill")
	}
	call("DELETE", session, nil, gov, false, 403)
	call("DELETE", session, nil, gov, true, 204)
	call("GET", projects, nil, gov, false, 401)
	mux = create()
	gov = login("department")
	r := call("GET", projects, nil, gov, false, 200)
	content, _ := io.ReadAll(r.Result().Body)
	if !strings.Contains(string(content), `"kind":"maintenance"`) {
		t.Fatal("project history did not survive restart")
	}
}
