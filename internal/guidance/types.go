package guidance

// GuidanceManifestItem represents a single guidance entity as listed by 'gydnc list'.
// It provides key metadata for quick assessment and filtering.
type GuidanceManifestItem struct {
	Alias          string                 `json:"alias"`                     // Human-readable alias (e.g., from filename)
	SourceBackend  string                 `json:"source_backend"`            // Name of the backend this item came from
	CID            string                 `json:"cid"`                       // Content ID of the current version of the file
	Title          string                 `json:"title,omitempty"`           // From 'title' field in frontmatter
	Description    string                 `json:"description,omitempty"`     // From 'description' field in frontmatter
	Tags           []string               `json:"tags,omitempty"`            // From 'tags' field in frontmatter
	CustomMetadata map[string]interface{} `json:"custom_metadata,omitempty"` // All other frontmatter fields
	Body           string                 `json:"body,omitempty"`            // The body content of the guidance, after frontmatter
}
