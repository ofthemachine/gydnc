package storage

import (
	"fmt"

	"gydnc/model"
	"gydnc/storage/inmem"
	"gydnc/storage/localfs"
)

// BackendRegistry stores registered backend instances by name
var BackendRegistry = make(map[string]ReadOnlyBackend)

// NewBackendFromConfig creates a new backend based on the provided configuration.
// configDir is the directory of the main gydnc config file, used to resolve relative paths in backend configs.
// Returns the backend interface and any error encountered during initialization.
func NewBackendFromConfig(name string, cfg *model.StorageConfig, configDir string) (ReadOnlyBackend, error) {
	if cfg == nil {
		return nil, fmt.Errorf("cannot create backend '%s' from nil config", name)
	}

	var backend ReadOnlyBackend
	var err error

	switch cfg.Type {
	case "localfs":
		if cfg.LocalFS == nil {
			return nil, fmt.Errorf("localfs config is required for backend '%s' (type 'localfs')", name)
		}
		// Pass configDir to localfs.NewStore
		store, storeErr := localfs.NewStore(*cfg.LocalFS, configDir)
		if storeErr != nil {
			return nil, fmt.Errorf("failed to create localfs backend '%s': %w", name, storeErr)
		}
		// Init now mainly sets the name, path resolution is in NewStore.
		if initErr := store.Init(map[string]interface{}{"name": name}); initErr != nil {
			return nil, fmt.Errorf("failed to initialize localfs backend '%s': %w", name, initErr)
		}
		backend = store

	case "inmem":
		// InMemStore might not need configDir, but the pattern should be consistent if it ever did.
		// For now, NewStore doesn't take it.
		store := inmem.NewStore(name)
		// If inmem.Store had an Init that could fail:
		// if err := store.Init(map[string]interface{}{"name": name}); err != nil {
		// 	 return nil, fmt.Errorf("failed to initialize inmem backend '%s': %w", name, err)
		// }
		backend = store

	default:
		return nil, fmt.Errorf("unsupported backend type '%s' for backend '%s'", cfg.Type, name)
	}

	if err != nil { // This check is somewhat redundant now as errors are returned directly above
		return nil, err
	}

	// Register the backend (optional, depends if registry is actively used elsewhere dynamically)
	// If AppContext.GetBackend relies on this registry, it's important.
	// If AppContext directly calls NewBackendFromConfig each time, it's less critical but can be a cache.
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
// configDir is the directory of the main configuration file, needed for resolving relative paths.
// Returns a map of backend names to initialization errors.
func InitializeBackends(cfg *model.Config, configDir string) map[string]error {
	errors := make(map[string]error)
	if cfg == nil {
		errors["_global"] = fmt.Errorf("cannot initialize backends from nil main config")
		return errors
	}

	for name, backendCfg := range cfg.StorageBackends {
		// Pass configDir down to NewBackendFromConfig
		_, err := NewBackendFromConfig(name, backendCfg, configDir)
		if err != nil {
			errors[name] = err
		}
	}

	return errors
}
