package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
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

func TestStructuredKnowledgeRequiresProjectEvidence(t *testing.T) {
	_, ts := testServer(t)
	res := postJSON(t, ts.URL, "/api/projects", map[string]string{"name": "Documental"})
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("project status %d", res.StatusCode)
	}
	var project struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(res.Body).Decode(&project)
	res.Body.Close()

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, _ := mw.CreateFormFile("media", "documentary.mp4")
	_, _ = part.Write([]byte("fake media bytes"))
	_ = mw.Close()
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/projects/"+project.ID+"/assets", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("upload status %d", res.StatusCode)
	}
	var asset struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(res.Body).Decode(&asset)
	res.Body.Close()
	res = postJSON(t, ts.URL, "/api/assets/"+asset.ID+"/transcript", map[string]string{"text": "00:00:10-00:00:20 | La investigadora explica el origen del proyecto."})
	if res.StatusCode != http.StatusOK {
		t.Fatalf("transcript status %d", res.StatusCode)
	}
	var transcript struct {
		Segments []struct {
			ID string `json:"id"`
		} `json:"segments"`
	}
	_ = json.NewDecoder(res.Body).Decode(&transcript)
	res.Body.Close()
	if len(transcript.Segments) != 1 {
		t.Fatalf("expected one evidence segment, got %d", len(transcript.Segments))
	}
	evidenceID := transcript.Segments[0].ID
	base := "/api/projects/" + project.ID + "/knowledge"

	res = postJSON(t, ts.URL, base+"/entities", map[string]any{"name": "Investigadora", "type": "person", "evidence_ids": []string{evidenceID}, "confidence": 1.0})
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("entity status %d: %s", res.StatusCode, readBody(t, res))
	}
	var entity struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(res.Body).Decode(&entity)
	res.Body.Close()

	res = postJSON(t, ts.URL, base+"/claims", map[string]any{"text": "La investigadora explica el origen del proyecto.", "entity_ids": []string{entity.ID}, "evidence_ids": []string{evidenceID}, "confidence": 0.9})
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("claim status %d: %s", res.StatusCode, readBody(t, res))
	}
	var claim struct {
		ZyrionState int `json:"zyrion_state"`
	}
	_ = json.NewDecoder(res.Body).Decode(&claim)
	res.Body.Close()
	if claim.ZyrionState != 2 {
		t.Fatalf("new claim must default to indeterminate (2), got %d", claim.ZyrionState)
	}

	res = postJSON(t, ts.URL, base+"/relations", map[string]any{"subject_entity_id": entity.ID, "object_entity_id": entity.ID, "predicate": "relates_to", "evidence_ids": []string{evidenceID}, "confidence": 1.0})
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("self relation should be rejected, got %d", res.StatusCode)
	}
	res = postJSON(t, ts.URL, base+"/entities", map[string]any{"name": "Proyecto", "type": "concept", "evidence_ids": []string{evidenceID}, "confidence": 0.8})
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("second entity status %d: %s", res.StatusCode, readBody(t, res))
	}
	var secondEntity struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(res.Body).Decode(&secondEntity)
	res.Body.Close()
	res = postJSON(t, ts.URL, base+"/relations", map[string]any{"subject_entity_id": entity.ID, "object_entity_id": secondEntity.ID, "predicate": "explains_origin_of", "evidence_ids": []string{evidenceID}, "confidence": 0.8})
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("valid relation status %d: %s", res.StatusCode, readBody(t, res))
	}
	res.Body.Close()

	res = postJSON(t, ts.URL, base+"/claims", map[string]any{"text": "Unsupported claim", "evidence_ids": []string{"missing-evidence"}, "confidence": 1.0})
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("unknown evidence should be rejected, got %d", res.StatusCode)
	}

	res, err = http.Get(ts.URL + base)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("knowledge snapshot status %d", res.StatusCode)
	}
	var graph struct {
		Entities  []any `json:"entities"`
		Claims    []any `json:"claims"`
		Relations []any `json:"relations"`
	}
	if err = json.NewDecoder(res.Body).Decode(&graph); err != nil {
		t.Fatal(err)
	}
	if len(graph.Entities) != 2 || len(graph.Claims) != 1 || len(graph.Relations) != 1 {
		t.Fatalf("unexpected knowledge snapshot: %#v", graph)
	}
}

func readBody(t *testing.T, r *http.Response) string {
	t.Helper()
	defer r.Body.Close()
	b, _ := io.ReadAll(r.Body)
	return string(b)
}


func TestSearchNoMatchesReturnsEmptyArray(t *testing.T) {
	_, ts := testServer(t)
	res := postJSON(t, ts.URL, "/api/projects", map[string]string{"name": "Search empty"})
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("project status %d", res.StatusCode)
	}
	var project struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(res.Body).Decode(&project)
	res.Body.Close()

	res = postJSON(t, ts.URL, "/api/search", map[string]any{
		"project_id": project.ID,
		"query":      "not-present",
		"limit":      10,
	})
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("search status %d", res.StatusCode)
	}
	var out struct {
		Count   int               `json:"count"`
		Results []json.RawMessage `json:"results"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.Results == nil {
		t.Fatal("results must be an empty JSON array, not null")
	}
	if out.Count != 0 || len(out.Results) != 0 {
		t.Fatalf("expected zero results, got count=%d results=%d", out.Count, len(out.Results))
	}
}
