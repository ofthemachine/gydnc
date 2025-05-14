package cmd

import (
	"fmt"

	"gydnc/config"
	"gydnc/storage"
	"gydnc/storage/localfs"
)

var activeBackend storage.Backend

// InitActiveBackend initializes the storage backend based on the global configuration.
// It should be called after the configuration has been loaded.
func InitActiveBackend() error {
	cfg := config.Get()

	backendName := cfg.DefaultBackend
	if backendName == "" {
		// If no default backend is specified, we might try to find *any* localfs backend
		// or default to a conventional name if appropriate for `gydnc init` scenarios.
		// For MVP, let's assume `gydnc init` will set a DefaultBackend.
		// If not, commands needing a backend will fail if activeBackend is nil.
		fmt.Println("Notice: No DefaultBackend specified in configuration. Some commands may not function.")
		activeBackend = nil
		return nil
	}

	storageCfg, ok := cfg.StorageBackends[backendName]
	if !ok || storageCfg == nil {
		fmt.Printf("Notice: Configuration for default backend '%s' not found. Some commands may not function.\n", backendName)
		activeBackend = nil
		return nil
	}

	if storageCfg.Type != "localfs" {
		return fmt.Errorf("default backend '%s' is of type '%s', but only 'localfs' is supported in MVP", backendName, storageCfg.Type)
	}

	if storageCfg.LocalFS == nil {
		return fmt.Errorf("localfs configuration for backend '%s' is missing", backendName)
	}

	localStore, err := localfs.NewStore(*storageCfg.LocalFS) // storageCfg.LocalFS is already a pointer, dereference for NewStore
	if err != nil {
		return fmt.Errorf("failed to create new localfs store for backend '%s': %w", backendName, err)
	}

	if err := localStore.Init(nil); err != nil {
		return fmt.Errorf("failed to initialize localfs store for backend '%s' at %s: %w", backendName, storageCfg.LocalFS.Path, err)
	}

	activeBackend = localStore
	// fmt.Printf("Active backend set to: '%s' (type: %s) at %s\n", backendName, activeBackend.GetName(), storageCfg.LocalFS.Path)
	return nil
}

// GetActiveBackend returns the currently active and initialized storage backend.
// Commands should check if the returned backend is nil.
func GetActiveBackend() storage.Backend {
	return activeBackend
}
