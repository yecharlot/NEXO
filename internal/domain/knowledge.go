package domain

import "time"

// Entity is a named concept or participant explicitly grounded in source evidence.
type Entity struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Description string    `json:"description,omitempty"`
	EvidenceIDs []string  `json:"evidence_ids"`
	Source      string    `json:"source"`
	Confidence  float64   `json:"confidence"`
	CreatedAt   time.Time `json:"created_at"`
}

// Claim is a statement linked to one or more exact evidence segments.
// ZyrionState follows the ecosystem contract: 1=True, 0=False, 2=Indeterminate.
// New claims default to 2; NEXO does not infer truth automatically in this phase.
type Claim struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	Text        string    `json:"text"`
	Kind        string    `json:"kind"`
	ZyrionState int       `json:"zyrion_state"`
	EntityIDs   []string  `json:"entity_ids,omitempty"`
	EvidenceIDs []string  `json:"evidence_ids"`
	Source      string    `json:"source"`
	Confidence  float64   `json:"confidence"`
	CreatedAt   time.Time `json:"created_at"`
}

// Relation connects two entities without discarding the evidence that supports it.
type Relation struct {
	ID              string    `json:"id"`
	ProjectID       string    `json:"project_id"`
	SubjectEntityID string    `json:"subject_entity_id"`
	Predicate       string    `json:"predicate"`
	ObjectEntityID  string    `json:"object_entity_id"`
	EvidenceIDs     []string  `json:"evidence_ids"`
	Source          string    `json:"source"`
	Confidence      float64   `json:"confidence"`
	CreatedAt       time.Time `json:"created_at"`
}
