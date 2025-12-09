package types

// GuidanceListItem represents a guidance entity in list operations
type GuidanceListItem struct {
	Alias string   `json:"alias" jsonschema:"the unique identifier for the guidance entity"`
	Title string   `json:"title" jsonschema:"the title of the guidance entity"`
	Tags  []string `json:"tags" jsonschema:"tags associated with the guidance entity"`
}

// GuidanceGetItem represents a guidance entity in get operations
type GuidanceGetItem struct {
	Title       string   `json:"title" jsonschema:"the title of the guidance entity"`
	Description string   `json:"description,omitempty" jsonschema:"the description of the guidance entity"`
	Tags        []string `json:"tags" jsonschema:"tags associated with the guidance entity"`
	Body        string   `json:"body" jsonschema:"the full body content of the guidance entity"`
}

// GuidanceWriteOutput represents the output of write operations
type GuidanceWriteOutput struct {
	Operation string `json:"operation" jsonschema:"the operation that was performed"`
	Alias     string `json:"alias" jsonschema:"the alias of the entity that was written"`
	Backend   string `json:"backend" jsonschema:"the backend where the entity was written"`
	Success   bool   `json:"success" jsonschema:"whether the operation succeeded"`
	Message   string `json:"message,omitempty" jsonschema:"optional message about the operation"`
}
