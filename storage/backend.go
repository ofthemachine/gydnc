package storage

// ReadOnlyBackend defines the minimal interface for read-only backends.
type ReadOnlyBackend interface {
	// Read retrieves the raw content and metadata of a guidance entity by its alias.
	Read(alias string) (content []byte, metadata map[string]interface{}, err error)
	// List retrieves a list of guidance entity IDs (aliases) based on a prefix.
	List(prefix string) ([]string, error)
	// Stat retrieves metadata about a guidance entity by its alias.
	Stat(id string) (map[string]interface{}, error)
	// GetName returns a unique name for the backend implementation (e.g., "localfs", "git").
	GetName() string
	// IsWritable returns true if this backend supports write operations.
	IsWritable() bool
	// Capabilities returns a map of capability names to boolean values.
	// This allows for feature detection and future extensibility.
	Capabilities() map[string]bool
}

// Backend defines the interface for writable guidance storage backends.
type Backend interface {
	ReadOnlyBackend
	// Init initializes the backend with the given configuration.
	Init(config map[string]interface{}) error
	// Write creates or updates a guidance entity with the given alias and data.
	Write(alias string, data []byte, commitMsgDetails map[string]string) error
	// Delete removes a guidance entity by its alias.
	Delete(alias string) error
	// IsWritable returns true if this backend supports write operations.
	IsWritable() bool
	// Future: Capabilities() map[string]bool
}
