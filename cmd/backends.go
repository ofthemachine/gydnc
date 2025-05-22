package cmd

import (
	"fmt"
	"log/slog"      // Standard library slog
	"path/filepath" // Import filepath

	"gydnc/model"
	"gydnc/service"
	"gydnc/storage"
	"gydnc/storage/localfs"
)

var activeBackend storage.Backend
var activeBackendName string // Store the name of the active backend
var cfgService *service.ConfigService

// InitActiveBackend initializes the storage backend based on the global configuration.
// It should be called after the configuration has been loaded.
func InitActiveBackend() error {
	// Initialize the app context and config service
	if appContext == nil || appContext.Config == nil {
		// This should be called after initialization in root.go
		return fmt.Errorf("app context or config is not initialized")
	}

	cfg := appContext.Config
	cfgService = service.NewConfigService(appContext)

	// Get configured path from current app context
	configPath, err := cfgService.GetEffectiveConfigPath(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	slog.Debug("[InitActiveBackend] Loaded config path", "cfgPath", configPath)

	backendN := cfg.DefaultBackend
	if backendN == "" {
		// If no default backend is specified, we might try to find *any* localfs backend
		// or default to a conventional name if appropriate for `gydnc init` scenarios.
		// For MVP, let's assume `gydnc init` will set a DefaultBackend.
		// If not, commands needing a backend will fail if activeBackend is nil.
		fmt.Println("Notice: No DefaultBackend specified in configuration. Some commands may not function.")
		activeBackend = nil
		activeBackendName = ""
		return nil
	}

	storageCfg, ok := cfg.StorageBackends[backendN]
	if !ok || storageCfg == nil {
		fmt.Printf("Notice: Configuration for default backend '%s' not found. Some commands may not function.\n", backendN)
		activeBackend = nil
		activeBackendName = ""
		return nil
	}

	if storageCfg.Type != "localfs" {
		activeBackend = nil // Ensure it's nil if not usable
		activeBackendName = ""
		return fmt.Errorf("backend '%s' has an unsupported type '%s' for the create command", backendN, storageCfg.Type)
	}

	if storageCfg.LocalFS == nil {
		activeBackend = nil // Ensure it's nil
		activeBackendName = ""
		return fmt.Errorf("localfs configuration for backend '%s' is missing", backendN)
	}

	// Resolve the LocalFS path: if relative, it's relative to the config file's directory.
	resolvedPath := storageCfg.LocalFS.Path
	slog.Debug("[InitActiveBackend] Initial resolvedPath from storageCfg.LocalFS.Path", "resolvedPath", resolvedPath)

	if !filepath.IsAbs(resolvedPath) && configPath != "" {
		configFileDir := filepath.Dir(configPath)
		resolvedPath = filepath.Join(configFileDir, resolvedPath)
		slog.Debug("[InitActiveBackend] Path resolved relative to config file dir", "configFileDir", configFileDir, "newResolvedPath", resolvedPath)
	}
	// Make it absolute for the store, as the store expects an absolute base path.
	absResolvedPath, err := filepath.Abs(resolvedPath)
	if err != nil {
		activeBackend = nil
		activeBackendName = ""
		return fmt.Errorf("failed to get absolute path for resolved localfs path '%s': %w", resolvedPath, err)
	}
	slog.Debug("[InitActiveBackend] Final absolute resolved path for store", "absResolvedPath", absResolvedPath)

	// Create a new LocalFSConfig with the resolved absolute path for the store
	// This is crucial because NewStore and the Store itself expect/work with an absolute basePath
	storeSpecificConfig := model.LocalFSConfig{Path: absResolvedPath}

	localStore, err := localfs.NewStore(storeSpecificConfig)
	if err != nil {
		activeBackend = nil
		activeBackendName = ""
		return fmt.Errorf("failed to create new localfs store for backend '%s' (resolved path: %s): %w", backendN, absResolvedPath, err)
	}

	if err := localStore.Init(nil); err != nil {
		activeBackend = nil
		activeBackendName = ""
		// Use the original configured path for error reporting if it's more user-friendly
		userFacingPath := storageCfg.LocalFS.Path
		if absResolvedPath != userFacingPath { // if path was resolved from relative to absolute
			userFacingPath = fmt.Sprintf("%s (resolved to %s)", storageCfg.LocalFS.Path, absResolvedPath)
		}
		return fmt.Errorf("failed to initialize localfs store for backend '%s' at %s: %w", backendN, userFacingPath, err)
	}

	activeBackend = localStore
	activeBackendName = backendN // Correctly store the name on success
	return nil
}

// GetActiveBackend returns the currently active and initialized storage backend and its name.
// Commands should check if the returned backend is nil.
func GetActiveBackend() (storage.Backend, string) { // Return name as well
	return activeBackend, activeBackendName
}

// InitializeBackendFromConfig initializes a backend from the provided configuration
func InitializeBackendFromConfig(backendName string, backendConfig *model.StorageConfig) (storage.Backend, error) {
	if backendConfig == nil {
		return nil, fmt.Errorf("backend configuration is nil")
	}

	if backendConfig.Type != "localfs" {
		return nil, fmt.Errorf("backend '%s' has an unsupported type '%s'", backendName, backendConfig.Type)
	}

	if backendConfig.LocalFS == nil {
		return nil, fmt.Errorf("localfs configuration for backend '%s' is missing", backendName)
	}

	// Get config path from service
	configPath, err := cfgService.GetEffectiveConfigPath(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	// Resolve the path if it's relative to the config file's directory
	resolvedPath := backendConfig.LocalFS.Path

	if !filepath.IsAbs(resolvedPath) && configPath != "" {
		configFileDir := filepath.Dir(configPath)
		resolvedPath = filepath.Join(configFileDir, resolvedPath)
	}

	// Create a new LocalFSConfig with the resolved path
	storeSpecificConfig := model.LocalFSConfig{Path: resolvedPath}

	localStore, err := localfs.NewStore(storeSpecificConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create new localfs store for backend '%s': %w", backendName, err)
	}

	if err := localStore.Init(nil); err != nil {
		return nil, fmt.Errorf("failed to initialize localfs store for backend '%s': %w", backendName, err)
	}

	// Set the name of the store
	localStore.SetName(backendName)

	return localStore, nil
}
