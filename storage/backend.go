package storage

// Backend defines the interface for guidance storage backends.
// Each method is responsible for handling interactions with a specific storage mechanism (e.g., local filesystem, Git repository).
// Errors returned should be descriptive and allow the caller to understand the nature of the failure.
type Backend interface {
	// Init initializes the backend with the given configuration.
	// The config map can contain backend-specific settings.
	// It should ensure that the backend is ready for operations (e.g., directory exists, connection established).
	Init(config map[string]interface{}) error

	// Read retrieves the raw content and metadata of a guidance entity by its alias.
	// It returns the raw bytes of the entity (e.g., content of a .g6e file),
	// a map of metadata (e.g., parsed frontmatter, last modified date, author),
	// and an error if the entity cannot be read.
	// For .g6e files, the raw content includes both frontmatter and body.
	// The metadata map allows flexibility in what backends can provide.
	Read(alias string) (content []byte, metadata map[string]interface{}, err error)

	// Write creates or updates a guidance entity with the given alias and data.
	// commitMsgDetails provides context for version control systems (e.g., author, message, operation type like create/update).
	// It should handle creating necessary structures (like subdirectories for aliases with '/').
	// Implementations should be idempotent where possible.
	Write(alias string, data []byte, commitMsgDetails map[string]string) error

	// List returns a list of available guidance aliases, potentially filtered by a query string.
	// The filterQuery syntax is backend-dependent but should allow for common use cases like tag filtering.
	// It returns a slice of aliases (e.g., "my-rule", "directory/my-other-rule") and an error.
	List(filterQuery string) (aliases []string, err error)

	// GetName returns a unique name for the backend implementation (e.g., "localfs", "git").
	GetName() string

	// Future methods could include:
	// Delete(alias string) error
	// Exists(alias string) (bool, error)
	// GetTags(alias string) ([]string, error)
	// GetAllTags() ([]string, error)
}
