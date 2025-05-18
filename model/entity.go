package model

// Entity represents a single guidance entity as listed or retrieved from a backend.
// It provides key metadata for quick assessment, filtering, and internal operations.
type Entity struct {
	Alias          string                 `json:"alias"`                     // Human-readable alias (e.g., from filename)
	SourceBackend  string                 `json:"source_backend"`            // Name of the backend this item came from
	Title          string                 `json:"title,omitempty"`           // From 'title' field in frontmatter
	Description    string                 `json:"description,omitempty"`     // From 'description' field in frontmatter
	Tags           []string               `json:"tags,omitempty"`            // From 'tags' field in frontmatter
	CustomMetadata map[string]interface{} `json:"custom_metadata,omitempty"` // All other frontmatter fields
	Body           string                 `json:"body,omitempty"`            // The body content of the guidance, after frontmatter

	CID string `json:"-"` // Internal content ID, not surfaced in CLI output
}
