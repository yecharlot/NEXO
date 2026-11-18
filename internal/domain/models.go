package domain

import "time"

type Project struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description,omitempty"`
	CoreOrganismID string    `json:"core_organism_id,omitempty"`
	CoreStatus     string    `json:"core_status,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type Segment struct {
	ID              string   `json:"id"`
	Start           float64  `json:"start_seconds"`
	End             float64  `json:"end_seconds"`
	Text            string   `json:"text"`
	Kind            string   `json:"kind,omitempty"`
	TimestampSource string   `json:"timestamp_source,omitempty"`
	Entities        []string `json:"entities,omitempty"`
	Confidence      float64  `json:"confidence,omitempty"`
}

type Asset struct {
	ID             string    `json:"id"`
	ProjectID      string    `json:"project_id"`
	Title          string    `json:"title"`
	Filename       string    `json:"filename"`
	ContentType    string    `json:"content_type"`
	Size           int64     `json:"size_bytes"`
	MediaPath      string    `json:"media_path,omitempty"`
	Transcript     string    `json:"transcript,omitempty"`
	Segments       []Segment `json:"segments"`
	CreatedAt      time.Time `json:"created_at"`
	AnalysisStatus string    `json:"analysis_status"`
}

type License struct {
	Plan        string    `json:"plan"`
	ActivatedAt time.Time `json:"activated_at,omitempty"`
}

type Database struct {
	Projects  map[string]Project  `json:"projects"`
	Assets    map[string]Asset    `json:"assets"`
	Entities  map[string]Entity   `json:"entities"`
	Claims    map[string]Claim    `json:"claims"`
	Relations map[string]Relation `json:"relations"`
	License   License             `json:"license"`
}
