package search

import (
	"github.com/yecharlot/NEXO/internal/domain"
	"testing"
)

func TestSearchReturnsTimestampedEvidence(t *testing.T) {
	assets := map[string]domain.Asset{"a": {ID: "a", Title: "Lecture", ProjectID: "p", Segments: []domain.Segment{{ID: "ev1", Start: 12, End: 18, Text: "Solar energy storage improves grid stability"}}}}
	got := Search(assets, "p", "solar storage", 10)
	if len(got) != 1 {
		t.Fatalf("want 1 result, got %d", len(got))
	}
	if got[0].Start != 12 || got[0].EvidenceID != "ev1" {
		t.Fatalf("timestamp/evidence lost: %#v", got[0])
	}
}
func TestSearchScopesToProject(t *testing.T) {
	assets := map[string]domain.Asset{"a": {ID: "a", ProjectID: "p1", Segments: []domain.Segment{{ID: "1", Text: "climate adaptation"}}}, "b": {ID: "b", ProjectID: "p2", Segments: []domain.Segment{{ID: "2", Text: "climate adaptation"}}}}
	got := Search(assets, "p1", "climate", 10)
	if len(got) != 1 || got[0].AssetID != "a" {
		t.Fatalf("unexpected project scope: %#v", got)
	}
}

func TestSearchNormalizesSpanishAccents(t *testing.T) {
	assets := map[string]domain.Asset{"a": {ID: "a", Segments: []domain.Segment{{ID: "ev", Text: "La energía solar"}}}}
	got := Search(assets, "", "energia", 10)
	if len(got) != 1 {
		t.Fatalf("accent-insensitive search failed: %#v", got)
	}
}
