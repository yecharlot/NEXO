package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Core struct {
	BaseURL, Token string
	Client         *http.Client
}

func NewCore(base, token string) *Core {
	return &Core{BaseURL: strings.TrimRight(base, "/"), Token: token, Client: &http.Client{Timeout: 4 * time.Second}}
}
func (c *Core) configured() bool { return c != nil && c.BaseURL != "" }
func (c *Core) request(ctx context.Context, method, path string, body any) (map[string]any, error) {
	if !c.configured() {
		return nil, fmt.Errorf("PrismaTEC-Core URL not configured")
	}
	var b bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&b).Encode(body); err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, &b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var out map[string]any
	if err = json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("AIP %s: HTTP %d: %v", path, res.StatusCode, out)
	}
	return out, nil
}
func (c *Core) Info(ctx context.Context) (map[string]any, error) {
	return c.request(ctx, http.MethodGet, "/aip/v1/info", nil)
}
func (c *Core) CreateOrganism(ctx context.Context, name string) (string, error) {
	out, err := c.request(ctx, http.MethodPost, "/aip/v1/commands", map[string]any{"aip": "v1", "action": "create", "params": map[string]any{"name": name, "capabilities": []string{"memory.read", "memory.write"}}})
	if err != nil {
		return "", err
	}
	org, ok := out["organism"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("AIP create response did not contain organism")
	}
	id, ok := org["id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("AIP create response did not contain organism id")
	}
	_, err = c.request(ctx, http.MethodPost, "/aip/v1/commands", map[string]any{"aip": "v1", "action": "start", "params": map[string]any{"id": id}})
	if err != nil {
		return "", err
	}
	return id, nil
}
func (c *Core) SetMemory(ctx context.Context, organismID, key, value string) error {
	_, err := c.request(ctx, http.MethodPost, "/aip/v1/commands", map[string]any{"aip": "v1", "action": "memory.set", "params": map[string]any{"id": organismID, "key": key, "value": value}})
	return err
}

type MiniNode struct {
	BaseURL string
	Client  *http.Client
}

func NewMiniNode(base string) *MiniNode {
	return &MiniNode{BaseURL: strings.TrimRight(base, "/"), Client: &http.Client{Timeout: 20 * time.Second}}
}
func (m *MiniNode) Call(ctx context.Context, capability string, payload any) (map[string]any, error) {
	if m == nil || m.BaseURL == "" {
		return nil, fmt.Errorf("MiniNode not configured")
	}
	paths := map[string]string{"mind": "/api/mind/tick", "zyrion": "/api/zyrion", "lispai": "/api/lispai"}
	path, ok := paths[capability]
	if !ok {
		return nil, fmt.Errorf("unknown cognitive capability")
	}
	b, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.BaseURL+path, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := m.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var out map[string]any
	if err = json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("MiniNode %s: HTTP %d", capability, res.StatusCode)
	}
	return out, nil
}
