package httpapi

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/yecharlot/NEXO/internal/domain"
	"github.com/yecharlot/NEXO/internal/integration"
	"github.com/yecharlot/NEXO/internal/search"
	"github.com/yecharlot/NEXO/internal/store"
	"github.com/yecharlot/NEXO/internal/transcription"
)

type Server struct {
	db                                         *store.Store
	dataDir, webDir, token, paymentURL, proKey string
	core                                       *integration.Core
	mini                                       *integration.MiniNode
	transcriber                                *transcription.Client
	llm                                        *integration.LLM
	mediaSecret                                []byte
	mux                                        *http.ServeMux
}

func New(db *store.Store, dataDir, webDir, token, paymentURL, proKey, coreURL, coreToken, miniURL string) *Server {
	secret := make([]byte, 32)
	if token != "" {
		secret = []byte(token)
	} else if _, err := rand.Read(secret); err != nil {
		panic(err)
	}
	s := &Server{mediaSecret: secret, db: db, dataDir: dataDir, webDir: webDir, token: token, paymentURL: paymentURL, proKey: proKey, core: integration.NewCore(coreURL, coreToken), mini: integration.NewMiniNode(miniURL), transcriber: transcription.New(os.Getenv("NEXO_TRANSCRIPTION_URL"), os.Getenv("NEXO_TRANSCRIPTION_KEY"), os.Getenv("NEXO_TRANSCRIPTION_MODEL")), llm: integration.NewLLM(os.Getenv("NEXO_LLM_URL"), os.Getenv("NEXO_LLM_API_KEY"), os.Getenv("NEXO_LLM_MODEL")), mux: http.NewServeMux()}
	s.routes()
	return s
}
func (s *Server) Handler() http.Handler { return s.recoverer(s.cors(s.auth(s.mux))) }
func (s *Server) routes() {
	s.mux.HandleFunc("/api/health", s.health)
	s.mux.HandleFunc("/api/config", s.config)
	s.mux.HandleFunc("/api/projects", s.projects)
	s.mux.HandleFunc("/api/projects/", s.project)
	s.mux.HandleFunc("/api/assets/", s.asset)
	s.mux.HandleFunc("/api/search", s.search)
	s.mux.HandleFunc("/api/ask", s.ask)
	s.mux.HandleFunc("/api/export.csv", s.exportCSV)
	s.mux.HandleFunc("/api/license/activate", s.activate)
	s.mux.HandleFunc("/api/cognition/", s.cognition)
	s.mux.HandleFunc("/media/", s.signedMedia)
	s.mux.Handle("/", http.FileServer(http.Dir(s.webDir)))
}
func (s *Server) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.token != "" && strings.HasPrefix(r.URL.Path, "/api/") && r.URL.Path != "/api/health" {
			if r.Header.Get("Authorization") != "Bearer "+s.token {
				jsonErr(w, 401, "unauthorized")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func (s *Server) recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				jsonErr(w, 500, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
func jsonOut(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func jsonErr(w http.ResponseWriter, status int, msg string) {
	jsonOut(w, status, map[string]string{"error": msg})
}
func readJSON(r *http.Request, v any) error {
	d := json.NewDecoder(io.LimitReader(r.Body, 2<<20))
	d.DisallowUnknownFields()
	return d.Decode(v)
}
func id(prefix string) string { return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano()) }
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	jsonOut(w, 200, map[string]any{"ok": true, "product": "NEXO", "version": "0.1.0"})
}
func (s *Server) config(w http.ResponseWriter, r *http.Request) {
	coreConfigured := os.Getenv("NEXO_PRISMATEC_URL") != ""
	miniConfigured := os.Getenv("NEXO_MININODE_URL") != ""
	jsonOut(w, 200, map[string]any{"pro": s.db.IsPro(), "payment_url": s.paymentURL, "max_upload_mb": 512, "prismatec_configured": coreConfigured, "mininode_configured": miniConfigured, "transcription_configured": s.transcriber.Configured(), "llm_configured": s.llm.Configured()})
}
func (s *Server) projects(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		db := s.db.Snapshot()
		out := make([]domain.Project, 0, len(db.Projects))
		for _, p := range db.Projects {
			out = append(out, p)
		}
		sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
		jsonOut(w, 200, map[string]any{"projects": out})
	case http.MethodPost:
		var in struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		if err := readJSON(r, &in); err != nil || strings.TrimSpace(in.Name) == "" {
			jsonErr(w, 400, "name is required")
			return
		}
		p := domain.Project{ID: id("prj"), Name: strings.TrimSpace(in.Name), Description: in.Description, CreatedAt: time.Now().UTC()}
		if err := s.db.PutProject(p); err != nil {
			jsonErr(w, 500, err.Error())
			return
		}
		if os.Getenv("NEXO_PRISMATEC_URL") != "" {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			if oid, e := s.core.CreateOrganism(ctx, "nexo-"+p.ID); e == nil {
				p.CoreOrganismID = oid
				p.CoreStatus = "connected"
				_ = s.db.PutProject(p)
			} else {
				p.CoreStatus = "unavailable: " + e.Error()
				_ = s.db.PutProject(p)
			}
		} else {
			p.CoreStatus = "not_configured"
			_ = s.db.PutProject(p)
		}
		jsonOut(w, 201, p)
	default:
		jsonErr(w, 405, "method not allowed")
	}
}
func (s *Server) project(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/projects/"), "/"), "/")
	if len(parts) == 2 && parts[1] == "assets" {
		s.upload(w, r, parts[0])
		return
	}
	if len(parts) != 1 || parts[0] == "" {
		jsonErr(w, 404, "not found")
		return
	}
	db := s.db.Snapshot()
	p, ok := db.Projects[parts[0]]
	if !ok {
		jsonErr(w, 404, "project not found")
		return
	}
	assets := []domain.Asset{}
	for _, a := range db.Assets {
		if a.ProjectID == p.ID {
			a.MediaPath = ""
			assets = append(assets, a)
		}
	}
	sort.Slice(assets, func(i, j int) bool { return assets[i].CreatedAt.After(assets[j].CreatedAt) })
	jsonOut(w, 200, map[string]any{"project": p, "assets": assets})
}
func (s *Server) asset(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/assets/"), "/"), "/")
	if len(parts) < 1 || parts[0] == "" {
		jsonErr(w, 404, "not found")
		return
	}
	a, ok := s.db.GetAsset(parts[0])
	if !ok {
		jsonErr(w, 404, "asset not found")
		return
	}
	if len(parts) == 1 && r.Method == http.MethodGet {
		a.MediaPath = ""
		jsonOut(w, 200, a)
		return
	}
	if len(parts) == 2 && parts[1] == "transcript" && r.Method == http.MethodPost {
		s.setTranscript(w, r, a)
		return
	}
	if len(parts) == 2 && parts[1] == "captions" && r.Method == http.MethodPost {
		s.setCaptions(w, r, a)
		return
	}
	if len(parts) == 2 && parts[1] == "media-url" && r.Method == http.MethodGet {
		exp := time.Now().Add(5 * time.Minute).Unix()
		sig := s.mediaSignature(a.ID, exp)
		jsonOut(w, 200, map[string]string{"url": fmt.Sprintf("/media/%s?exp=%d&sig=%s", a.ID, exp, sig)})
		return
	}
	if len(parts) == 2 && parts[1] == "media" && r.Method == http.MethodGet {
		if a.MediaPath == "" {
			jsonErr(w, 404, "media file not found")
			return
		}
		http.ServeFile(w, r, a.MediaPath)
		return
	}
	jsonErr(w, 404, "route not found")
}
func (s *Server) upload(w http.ResponseWriter, r *http.Request, projectID string) {
	if r.Method != http.MethodPost {
		jsonErr(w, 405, "method not allowed")
		return
	}
	db := s.db.Snapshot()
	if _, ok := db.Projects[projectID]; !ok {
		jsonErr(w, 404, "project not found")
		return
	}
	if !s.db.IsPro() {
		count := 0
		for _, a := range db.Assets {
			if a.ProjectID == projectID {
				count++
			}
		}
		if count >= 3 {
			jsonErr(w, 402, "free plan includes 3 media assets; activate Pro to continue")
			return
		}
	}
	r.Body = http.MaxBytesReader(w, r.Body, 520<<20)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		jsonErr(w, 400, "invalid multipart upload or file exceeds 512 MB")
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}
	f, h, err := r.FormFile("media")
	if err != nil {
		jsonErr(w, 400, "media file is required")
		return
	}
	defer f.Close()
	ext := strings.ToLower(filepath.Ext(h.Filename))
	if !allowedExt(ext) {
		jsonErr(w, 415, "unsupported media type; use mp4, mov, webm, mp3, wav, m4a or mpeg")
		return
	}
	assetID := id("med")
	dir := filepath.Join(s.dataDir, "media", assetID)
	if err = os.MkdirAll(dir, 0700); err != nil {
		jsonErr(w, 500, "cannot create media storage")
		return
	}
	dst := filepath.Join(dir, "source"+ext)
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		jsonErr(w, 500, "cannot create media file")
		return
	}
	n, copyErr := io.Copy(out, io.LimitReader(f, 512<<20+1))
	closeErr := out.Close()
	if copyErr != nil || closeErr != nil || n > 512<<20 {
		_ = os.Remove(dst)
		jsonErr(w, 400, "upload failed or file exceeds 512 MB")
		return
	}
	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		title = strings.TrimSuffix(filepath.Base(h.Filename), ext)
	}
	a := domain.Asset{ID: assetID, ProjectID: projectID, Title: title, Filename: filepath.Base(h.Filename), ContentType: h.Header.Get("Content-Type"), Size: n, MediaPath: dst, Segments: []domain.Segment{}, CreatedAt: time.Now().UTC(), AnalysisStatus: "awaiting_transcript"}
	if err = s.db.PutAsset(a); err != nil {
		jsonErr(w, 500, "cannot save media metadata")
		return
	}
	if s.transcriber.Configured() {
		go s.transcribe(a)
	}
	if p, ok := db.Projects[projectID]; ok && p.CoreOrganismID != "" {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		_ = s.core.SetMemory(ctx, p.CoreOrganismID, "asset."+a.ID, a.Title)
		cancel()
	}
	a.MediaPath = ""
	jsonOut(w, 201, a)
}
func allowedExt(e string) bool {
	switch e {
	case ".mp4", ".mov", ".webm", ".mkv", ".mp3", ".wav", ".m4a", ".mpeg", ".mpg", ".ogg":
		return true
	}
	return false
}
func (s *Server) setTranscript(w http.ResponseWriter, r *http.Request, a domain.Asset) {
	var in struct {
		Text string `json:"text"`
	}
	if err := readJSON(r, &in); err != nil || strings.TrimSpace(in.Text) == "" {
		jsonErr(w, 400, "transcript text is required")
		return
	}
	a.Transcript = in.Text
	a.Segments = splitTranscript(in.Text)
	a.AnalysisStatus = "indexed"
	if err := s.db.PutAsset(a); err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	a.MediaPath = ""
	jsonOut(w, 200, a)
}
func splitTranscript(text string) []domain.Segment {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	out := []domain.Segment{}
	cursor := 0.0
	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}
		start, end, body, ok := parseTimedLine(t)
		if ok {
			t = body
		} else {
			start = cursor
			end = start + max(5, float64(len([]rune(t)))/15)
		}
		out = append(out, domain.Segment{ID: id("ev"), Start: start, End: end, Text: t, Kind: "transcript", TimestampSource: map[bool]string{true: "provided", false: "estimated"}[ok], Confidence: map[bool]float64{true: 1, false: 0.3}[ok]})
		cursor = end
	}
	return out
}
func parseTimedLine(line string) (float64, float64, string, bool) {
	parts := strings.SplitN(line, "|", 2)
	if len(parts) != 2 {
		return 0, 0, "", false
	}
	times := strings.Split(strings.TrimSpace(parts[0]), "-")
	if len(times) != 2 {
		return 0, 0, "", false
	}
	a, e1 := parseTime(times[0])
	b, e2 := parseTime(times[1])
	if e1 != nil || e2 != nil || b < a {
		return 0, 0, "", false
	}
	return a, b, strings.TrimSpace(parts[1]), true
}
func parseTime(v string) (float64, error) {
	v = strings.TrimSpace(v)
	p := strings.Split(v, ":")
	if len(p) < 2 || len(p) > 3 {
		return 0, errors.New("invalid time")
	}
	sec, _ := strconv.ParseFloat(p[len(p)-1], 64)
	min, _ := strconv.ParseFloat(p[len(p)-2], 64)
	var hr float64
	if len(p) == 3 {
		hr, _ = strconv.ParseFloat(p[0], 64)
	}
	return hr*3600 + min*60 + sec, nil
}
func (s *Server) setCaptions(w http.ResponseWriter, r *http.Request, a domain.Asset) {
	b, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 5<<20))
	if err != nil {
		jsonErr(w, 400, "captions too large")
		return
	}
	a.Transcript = string(b)
	a.Segments = parseSRT(string(b))
	if len(a.Segments) == 0 {
		a.Segments = splitTranscript(string(b))
	}
	a.AnalysisStatus = "indexed"
	if err = s.db.PutAsset(a); err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	a.MediaPath = ""
	jsonOut(w, 200, a)
}
func parseSRT(src string) []domain.Segment {
	blocks := strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n\n")
	out := []domain.Segment{}
	for _, b := range blocks {
		lines := strings.Split(strings.TrimSpace(b), "\n")
		idx := -1
		for i, l := range lines {
			if strings.Contains(l, "-->") {
				idx = i
				break
			}
		}
		if idx < 0 {
			continue
		}
		ts := strings.Split(lines[idx], "-->")
		if len(ts) != 2 {
			continue
		}
		a, e1 := parseSRTTime(ts[0])
		z, e2 := parseSRTTime(ts[1])
		if e1 != nil || e2 != nil {
			continue
		}
		txt := strings.TrimSpace(strings.Join(lines[idx+1:], " "))
		if txt != "" {
			out = append(out, domain.Segment{ID: id("ev"), Start: a, End: z, Text: txt, Kind: "caption", TimestampSource: "subtitle", Confidence: 1})
		}
	}
	return out
}
func parseSRTTime(v string) (float64, error) {
	v = strings.ReplaceAll(strings.TrimSpace(v), ",", ".")
	return parseTime(v)
}
func (s *Server) search(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonErr(w, 405, "method not allowed")
		return
	}
	var in struct {
		ProjectID string `json:"project_id"`
		Query     string `json:"query"`
		Limit     int    `json:"limit"`
	}
	if err := readJSON(r, &in); err != nil || strings.TrimSpace(in.Query) == "" {
		jsonErr(w, 400, "query is required")
		return
	}
	if in.Limit <= 0 || in.Limit > 100 {
		in.Limit = 25
	}
	db := s.db.Snapshot()
	res := search.Search(db.Assets, in.ProjectID, in.Query, in.Limit)
	jsonOut(w, 200, map[string]any{"query": in.Query, "count": len(res), "results": res, "mode": "evidence-retrieval"})
}
func (s *Server) exportCSV(w http.ResponseWriter, r *http.Request) {
	if !s.db.IsPro() {
		jsonErr(w, 402, "CSV export is a Pro feature")
		return
	}
	projectID := r.URL.Query().Get("project_id")
	db := s.db.Snapshot()
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=nexo-evidence.csv")
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"asset_id", "asset_title", "evidence_id", "start_seconds", "end_seconds", "timestamp_source", "text"})
	for _, a := range db.Assets {
		if projectID != "" && a.ProjectID != projectID {
			continue
		}
		for _, seg := range a.Segments {
			_ = cw.Write([]string{a.ID, a.Title, seg.ID, fmt.Sprintf("%.3f", seg.Start), fmt.Sprintf("%.3f", seg.End), seg.TimestampSource, seg.Text})
		}
	}
	cw.Flush()
}
func (s *Server) activate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonErr(w, 405, "method not allowed")
		return
	}
	if s.proKey == "" {
		jsonErr(w, 503, "Pro activation is not configured")
		return
	}
	var in struct {
		Key string `json:"key"`
	}
	if err := readJSON(r, &in); err != nil || in.Key != s.proKey {
		jsonErr(w, 403, "invalid activation key")
		return
	}
	if err := s.db.ActivatePro(); err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	jsonOut(w, 200, map[string]any{"ok": true, "plan": "pro"})
}

func (s *Server) transcribe(a domain.Asset) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	current, ok := s.db.GetAsset(a.ID)
	if !ok {
		return
	}
	current.AnalysisStatus = "transcribing"
	_ = s.db.PutAsset(current)
	result, err := s.transcriber.Transcribe(ctx, current.MediaPath)
	current, ok = s.db.GetAsset(a.ID)
	if !ok {
		return
	}
	if err != nil {
		current.AnalysisStatus = "transcription_failed"
		_ = s.db.PutAsset(current)
		return
	}
	current.Transcript = result.Text
	current.Segments = nil
	for _, seg := range result.Segments {
		if strings.TrimSpace(seg.Text) == "" {
			continue
		}
		current.Segments = append(current.Segments, domain.Segment{ID: id("ev"), Start: seg.Start, End: seg.End, Text: strings.TrimSpace(seg.Text), Kind: "transcript", TimestampSource: "automatic_transcription", Confidence: 1})
	}
	if len(current.Segments) == 0 {
		current.Segments = splitTranscript(result.Text)
	}
	current.AnalysisStatus = "indexed"
	_ = s.db.PutAsset(current)
}

func (s *Server) cognition(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonErr(w, 405, "method not allowed")
		return
	}
	capability := strings.TrimPrefix(r.URL.Path, "/api/cognition/")
	var payload map[string]any
	if err := readJSON(r, &payload); err != nil {
		jsonErr(w, 400, "invalid JSON payload")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()
	out, err := s.mini.Call(ctx, capability, payload)
	if err != nil {
		jsonErr(w, 503, err.Error())
		return
	}
	jsonOut(w, 200, out)
}

func (s *Server) ask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonErr(w, 405, "method not allowed")
		return
	}
	if !s.db.IsPro() {
		jsonErr(w, 402, "evidence-grounded AI answers are a Pro feature")
		return
	}
	var in struct {
		ProjectID string `json:"project_id"`
		Question  string `json:"question"`
	}
	if err := readJSON(r, &in); err != nil || strings.TrimSpace(in.Question) == "" {
		jsonErr(w, 400, "question is required")
		return
	}
	if !s.llm.Configured() {
		jsonErr(w, 503, "configure NEXO_LLM_API_KEY and NEXO_LLM_MODEL to enable grounded answers")
		return
	}
	db := s.db.Snapshot()
	matches := search.Search(db.Assets, in.ProjectID, in.Question, 20)
	if len(matches) == 0 {
		jsonOut(w, 200, map[string]any{"answer": "No encontré evidencias coincidentes en las transcripciones indexadas. Prueba con términos más concretos o añade subtítulos.", "evidence": []any{}, "mode": "no-evidence"})
		return
	}
	evidence := make([]map[string]any, 0, len(matches))
	for _, m := range matches {
		evidence = append(evidence, map[string]any{"EvidenceID": m.EvidenceID, "asset_id": m.AssetID, "source": m.AssetTitle, "start_seconds": m.Start, "end_seconds": m.End, "timestamp_source": m.TimestampSource, "text": m.Text})
	}
	ctx, cancel := context.WithTimeout(r.Context(), 95*time.Second)
	defer cancel()
	answer, err := s.llm.Answer(ctx, in.Question, evidence)
	if err != nil {
		jsonErr(w, 502, err.Error())
		return
	}
	jsonOut(w, 200, map[string]any{"answer": answer, "evidence": evidence, "mode": "grounded-answer", "warning": "La respuesta es una síntesis generada; verifica las referencias originales antes de publicar."})
}

func (s *Server) mediaSignature(assetID string, exp int64) string {
	mac := hmac.New(sha256.New, s.mediaSecret)
	_, _ = mac.Write([]byte(fmt.Sprintf("%s:%d", assetID, exp)))
	return hex.EncodeToString(mac.Sum(nil))
}
func (s *Server) signedMedia(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/media/")
	exp, err := strconv.ParseInt(r.URL.Query().Get("exp"), 10, 64)
	sig := r.URL.Query().Get("sig")
	if id == "" || err != nil || exp < time.Now().Unix() || exp > time.Now().Add(10*time.Minute).Unix() || !hmac.Equal([]byte(sig), []byte(s.mediaSignature(id, exp))) {
		http.Error(w, "invalid or expired media link", http.StatusForbidden)
		return
	}
	a, ok := s.db.GetAsset(id)
	if !ok || a.MediaPath == "" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, a.MediaPath)
}
