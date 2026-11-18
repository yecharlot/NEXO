package httpapi

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/yecharlot/NEXO/internal/store"
)

func testServer(t *testing.T) (*Server, *httptest.Server) {
	t.Helper()
	dir := t.TempDir()
	db, err := store.Open(filepath.Join(dir, "db.json"))
	if err != nil {
		t.Fatal(err)
	}
	web := filepath.Join(dir, "web")
	if err = os.MkdirAll(web, 0700); err != nil {
		t.Fatal(err)
	}
	s := New(db, dir, web, "", "", "test-pro-key", "", "", "")
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)
	return s, ts
}
func postJSON(t *testing.T, base, path string, v any) *http.Response {
	t.Helper()
	b, _ := json.Marshal(v)
	r, err := http.Post(base+path, "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	return r
}
func TestProjectUploadTranscriptAndSearch(t *testing.T) {
	_, ts := testServer(t)
	res := postJSON(t, ts.URL, "/api/projects", map[string]string{"name": "Research"})
	if res.StatusCode != 201 {
		t.Fatalf("create project status %d", res.StatusCode)
	}
	var p struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(res.Body).Decode(&p)
	res.Body.Close()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, _ := mw.CreateFormFile("media", "lecture.mp4")
	_, _ = part.Write([]byte("fake media bytes"))
	_ = mw.WriteField("title", "Lecture")
	_ = mw.Close()
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/projects/"+p.ID+"/assets", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 201 {
		t.Fatalf("upload status %d", res.StatusCode)
	}
	var a struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(res.Body).Decode(&a)
	res.Body.Close()
	mediaURL, err := http.Get(ts.URL + "/api/assets/" + a.ID + "/media-url")
	if err != nil {
		t.Fatal(err)
	}
	var media struct {
		URL string `json:"url"`
	}
	_ = json.NewDecoder(mediaURL.Body).Decode(&media)
	mediaURL.Body.Close()
	if mediaURL.StatusCode != 200 || media.URL == "" {
		t.Fatalf("media URL unavailable: %d", mediaURL.StatusCode)
	}
	mediaRes, err := http.Get(ts.URL + media.URL)
	if err != nil {
		t.Fatal(err)
	}
	mediaRes.Body.Close()
	if mediaRes.StatusCode != 200 {
		t.Fatalf("signed media route status %d", mediaRes.StatusCode)
	}
	res = postJSON(t, ts.URL, "/api/assets/"+a.ID+"/transcript", map[string]string{"text": "00:00:12-00:00:18 | Solar energy storage improves grid stability."})
	if res.StatusCode != 200 {
		t.Fatalf("transcript status %d", res.StatusCode)
	}
	res.Body.Close()
	res = postJSON(t, ts.URL, "/api/search", map[string]any{"project_id": p.ID, "query": "solar storage", "limit": 10})
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("search status %d", res.StatusCode)
	}
	var out struct {
		Count   int `json:"count"`
		Results []struct {
			Start float64 `json:"start_seconds"`
		} `json:"results"`
	}
	_ = json.NewDecoder(res.Body).Decode(&out)
	if out.Count != 1 || len(out.Results) != 1 || out.Results[0].Start != 12 {
		t.Fatalf("unexpected search results: %#v", out)
	}
}
func TestFreePlanAndProExportGate(t *testing.T) {
	s, _ := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/export.csv", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)
	if rr.Code != 402 {
		t.Fatalf("want 402, got %d", rr.Code)
	}
	rr = httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/license/activate", bytes.NewBufferString(`{"key":"test-pro-key"}`)))
	if rr.Code != 200 {
		t.Fatalf("activate status %d: %s", rr.Code, rr.Body.String())
	}
	rr = httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/export.csv", nil))
	if rr.Code != 200 {
		t.Fatalf("pro export status %d", rr.Code)
	}
}
