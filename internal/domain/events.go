package domain
import "time"
// Event conserva un acontecimiento temporal, su evidencia y la revisión humana.
type Event struct {
 ID string `json:"id"`
 ProjectID string `json:"project_id"`
 AssetID string `json:"asset_id"`
 Title string `json:"title"`
 Description string `json:"description,omitempty"`
 StartSeconds float64 `json:"start_seconds"`
 EndSeconds float64 `json:"end_seconds"`
 TimestampSource string `json:"timestamp_source,omitempty"`
 EvidenceIDs []string `json:"evidence_ids"`
 EntityIDs []string `json:"entity_ids,omitempty"`
 ClaimIDs []string `json:"claim_ids,omitempty"`
 Status string `json:"status"`
 Origin string `json:"origin"`
 Extractor string `json:"extractor,omitempty"`
 ExtractorVersion string `json:"extractor_version,omitempty"`
 CreatedAt time.Time `json:"created_at"`
 ReviewedAt time.Time `json:"reviewed_at,omitempty"`
 ReviewedBy string `json:"reviewed_by,omitempty"`
 ReviewNote string `json:"review_note,omitempty"`
}
