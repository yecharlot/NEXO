package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiniNodeRoutesMindZyrionAndLispAI(t *testing.T) {
	seen := map[string]bool{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen[r.URL.Path] = true
		var payload map[string]any
		_ = json.NewDecoder(r.Body).Decode(&payload)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	m := NewMiniNode(server.URL)
	for _, cap := range []string{"mind", "zyrion", "lispai"} {
		if _, err := m.Call(context.Background(), cap, map[string]any{}); err != nil {
			t.Fatal(err)
		}
	}
	if !seen["/api/mind/tick"] || !seen["/api/zyrion"] || !seen["/api/lispai"] {
		t.Fatalf("missing MiniNode route: %#v", seen)
	}
}
