package search

import (
	"sort"
	"strings"
	"unicode"

	"github.com/yecharlot/NEXO/internal/domain"
)

type Result struct {
	AssetID         string  `json:"asset_id"`
	AssetTitle      string  `json:"asset_title"`
	Start           float64 `json:"start_seconds"`
	End             float64 `json:"end_seconds"`
	Text            string  `json:"text"`
	Score           float64 `json:"score"`
	EvidenceID      string  `json:"evidence_id"`
	TimestampSource string  `json:"timestamp_source,omitempty"`
}

func tokens(s string) []string {
	s = strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n").Replace(strings.ToLower(s))
	var out []string
	var b strings.Builder
	flush := func() {
		if b.Len() > 1 {
			out = append(out, strings.ToLower(b.String()))
		}
		b.Reset()
	}
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()
	return out
}
func Search(assets map[string]domain.Asset, projectID, query string, limit int) []Result {
	q := map[string]bool{}
	for _, t := range tokens(query) {
		q[t] = true
	}
	if len(q) == 0 {
		return nil
	}
	var out []Result
	for _, a := range assets {
		if projectID != "" && a.ProjectID != projectID {
			continue
		}
		for _, s := range a.Segments {
			ts := tokens(s.Text)
			if len(ts) == 0 {
				continue
			}
			freq := map[string]int{}
			for _, t := range ts {
				freq[t]++
			}
			score := 0.0
			for t := range q {
				if n := freq[t]; n > 0 {
					score += 1 + float64(n-1)*0.25
				}
			}
			if score > 0 {
				score = score / float64(len(q))
				out = append(out, Result{AssetID: a.ID, AssetTitle: a.Title, Start: s.Start, End: s.End, Text: s.Text, Score: score, EvidenceID: s.ID, TimestampSource: s.TimestampSource})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score == out[j].Score {
			return out[i].Start < out[j].Start
		}
		return out[i].Score > out[j].Score
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}
