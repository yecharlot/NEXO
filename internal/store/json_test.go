package store

import (
	"github.com/yecharlot/NEXO/internal/domain"
	"os"
	"path/filepath"
	"testing"
)

func TestAssetMediaPathSurvivesRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db.json")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	a := domain.Asset{ID: "asset-1", ProjectID: "p", Title: "source", MediaPath: "/var/lib/nexo/media/source.mp4"}
	if err = s.PutAsset(a); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := reopened.GetAsset(a.ID)
	if !ok || got.MediaPath != a.MediaPath {
		t.Fatalf("media path lost after restart: %#v", got)
	}
}

func TestOpenMigratesLegacyDatabaseMaps(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.json")
	if err := os.WriteFile(path, []byte(`{"projects":{},"assets":{},"license":{"plan":"free"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	snapshot := s.Snapshot()
	if snapshot.Entities == nil || snapshot.Claims == nil || snapshot.Relations == nil {
		t.Fatalf("legacy database maps were not initialized: %#v", snapshot)
	}
}
