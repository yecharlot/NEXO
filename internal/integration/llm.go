package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type LLM struct {
	BaseURL, Key, Model string
	Client              *http.Client
}

func NewLLM(base, key, model string) *LLM {
	if base == "" {
		base = "https://api.openai.com/v1/chat/completions"
	}
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &LLM{BaseURL: base, Key: key, Model: model, Client: &http.Client{Timeout: 90 * time.Second}}
}
func (l *LLM) Configured() bool { return l != nil && l.Key != "" }
func (l *LLM) Answer(ctx context.Context, question string, evidence []map[string]any) (string, error) {
	if !l.Configured() {
		return "", fmt.Errorf("LLM not configured")
	}
	contextJSON, _ := json.Marshal(evidence)
	payload := map[string]any{"model": l.Model, "temperature": 0.1, "messages": []map[string]string{{"role": "system", "content": "Eres NEXO, un asistente de investigación audiovisual. Responde exclusivamente con la evidencia proporcionada. No inventes hechos ni citas. Distingue observación, inferencia e incertidumbre. Toda afirmación factual debe citar uno o más EvidenceID exactamente como [EvidenceID]. Si la evidencia no basta, dilo claramente. Responde en español."}, {"role": "user", "content": "Pregunta: " + question + "\n\nEvidencias recuperadas (JSON):\n" + string(contextJSON)}}}
	b, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, l.BaseURL, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+l.Key)
	req.Header.Set("Content-Type", "application/json")
	res, err := l.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(res.Body, 4<<20))
	if err != nil {
		return "", err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", fmt.Errorf("LLM provider HTTP %d: %s", res.StatusCode, strings.TrimSpace(string(raw)))
	}
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err = json.Unmarshal(raw, &out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("LLM returned no choices")
	}
	return strings.TrimSpace(out.Choices[0].Message.Content), nil
}
