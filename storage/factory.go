package storage

import (
	"fmt"

	"gydnc/config"
	"gydnc/storage/inmem"
	"gydnc/storage/localfs"
)

// BackendRegistry stores registered backend instances by name
var BackendRegistry = make(map[string]ReadOnlyBackend)

// NewBackendFromConfig creates a new backend based on the provided configuration.
// Returns the backend interface and any error encountered during initialization.
func NewBackendFromConfig(name string, cfg *config.StorageConfig) (ReadOnlyBackend, error) {
	if cfg == nil {
		return nil, fmt.Errorf("cannot create backend from nil config")
	}

	var backend ReadOnlyBackend

	switch cfg.Type {
	case "localfs":
		if cfg.LocalFS == nil {
			return nil, fmt.Errorf("localfs config is required for backend type 'localfs'")
		}
		store, err := localfs.NewStore(*cfg.LocalFS)
		if err != nil {
			return nil, fmt.Errorf("failed to create localfs backend: %w", err)
		}
		err = store.Init(map[string]interface{}{"name": name})
		if err != nil {
			return nil, fmt.Errorf("failed to initialize localfs backend: %w", err)
		}
		backend = store

	case "inmem":
		store := inmem.NewStore(name)
		backend = store

	default:
		return nil, fmt.Errorf("unsupported backend type: %s", cfg.Type)
	}

	// Register the backend
	BackendRegistry[name] = backend

	return backend, nil
}

// GetBackend retrieves a backend from the registry by name.
// Returns nil if the backend is not found.
func GetBackend(name string) ReadOnlyBackend {
	return BackendRegistry[name]
}

// ClearRegistry clears all registered backends.
// This is primarily useful for testing.
func ClearRegistry() {
	BackendRegistry = make(map[string]ReadOnlyBackend)
}

// InitializeBackends initializes all backends defined in the configuration.
// Returns a map of backend names to initialization errors.
func InitializeBackends(cfg *config.Config) map[string]error {
	errors := make(map[string]error)

	for name, backendCfg := range cfg.StorageBackends {
		_, err := NewBackendFromConfig(name, backendCfg)
		if err != nil {
			errors[name] = err
		}
	}

	return errors
}
