package transcription

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Segment struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}
type Result struct {
	Text     string    `json:"text"`
	Segments []Segment `json:"segments"`
}
type Client struct {
	URL, Key, Model string
	HTTP            *http.Client
	FFmpeg          string
}

func New(url, key, model string) *Client {
	if url == "" {
		url = "https://api.openai.com/v1/audio/transcriptions"
	}
	if model == "" {
		model = "whisper-1"
	}
	ff, _ := exec.LookPath("ffmpeg")
	return &Client{URL: url, Key: key, Model: model, HTTP: &http.Client{Timeout: 20 * time.Minute}, FFmpeg: ff}
}
func (c *Client) Configured() bool { return c != nil && c.Key != "" && c.FFmpeg != "" }
func (c *Client) Transcribe(ctx context.Context, source string) (Result, error) {
	if !c.Configured() {
		return Result{}, fmt.Errorf("automatic transcription requires NEXO_TRANSCRIPTION_KEY and ffmpeg")
	}
	tmp, err := os.MkdirTemp("", "nexo-audio-")
	if err != nil {
		return Result{}, err
	}
	defer os.RemoveAll(tmp)
	pattern := filepath.Join(tmp, "chunk_%03d.mp3")
	cmd := exec.CommandContext(ctx, c.FFmpeg, "-hide_banner", "-loglevel", "error", "-y", "-i", source, "-vn", "-f", "segment", "-segment_time", "600", "-reset_timestamps", "1", "-c:a", "libmp3lame", "-b:a", "48k", "-ac", "1", "-ar", "16000", pattern)
	if out, e := cmd.CombinedOutput(); e != nil {
		return Result{}, fmt.Errorf("audio extraction: %w: %s", e, strings.TrimSpace(string(out)))
	}
	files, err := filepath.Glob(filepath.Join(tmp, "chunk_*.mp3"))
	if err != nil {
		return Result{}, err
	}
	if len(files) == 0 {
		return Result{}, fmt.Errorf("ffmpeg produced no audio chunks")
	}
	var combined Result
	for i, path := range files {
		part, err := c.transcribeChunk(ctx, path)
		if err != nil {
			return Result{}, fmt.Errorf("transcription chunk %d: %w", i+1, err)
		}
		if part.Text != "" {
			if combined.Text != "" {
				combined.Text += " "
			}
			combined.Text += part.Text
		}
		for _, seg := range part.Segments {
			seg.Start += float64(i * 600)
			seg.End += float64(i * 600)
			combined.Segments = append(combined.Segments, seg)
		}
	}
	return combined, nil
}

func (c *Client) transcribeChunk(ctx context.Context, path string) (Result, error) {
	f, err := os.Open(path)
	if err != nil {
		return Result{}, err
	}
	defer f.Close()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	_ = mw.WriteField("model", c.Model)
	_ = mw.WriteField("response_format", "verbose_json")
	_ = mw.WriteField("timestamp_granularities[]", "segment")
	part, err := mw.CreateFormFile("file", filepath.Base(path))
	if err != nil {
		return Result{}, err
	}
	if _, err = io.Copy(part, f); err != nil {
		return Result{}, err
	}
	if err = mw.Close(); err != nil {
		return Result{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL, &body)
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Key)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	res, err := c.HTTP.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(res.Body, 20<<20))
	if err != nil {
		return Result{}, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return Result{}, fmt.Errorf("transcription provider HTTP %d: %s", res.StatusCode, string(raw))
	}
	var result Result
	if err = json.Unmarshal(raw, &result); err != nil {
		return Result{}, err
	}
	return result, nil
}
