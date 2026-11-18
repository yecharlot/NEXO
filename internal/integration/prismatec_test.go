package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCoreCreateOrganismUsesAIPV1(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/aip/v1/commands" || r.Method != http.MethodPost {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if calls == 1 {
			if body["action"] != "create" {
				t.Fatalf("first action=%v", body["action"])
			}
			_, _ = w.Write([]byte(`{"ok":true,"organism":{"id":"nexo-organism-1"}}`))
			return
		}
		if body["action"] != "start" {
			t.Fatalf("second action=%v", body["action"])
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	client := NewCore(server.URL, "")
	id, err := client.CreateOrganism(context.Background(), "nexo-project")
	if err != nil {
		t.Fatal(err)
	}
	if id != "nexo-organism-1" || calls != 2 {
		t.Fatalf("id=%q calls=%d", id, calls)
	}
}
