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
	// cfgService = service.NewConfigService(appContext) // cfgService is already a global, ensure it's initialized if needed or pass appContext

	// Use appContext.ConfigPath directly
	configFilePath := appContext.ConfigPath
	if configFilePath == "" {
		// Fallback or error if ConfigPath is not set in AppContext
		// This might happen if initConfig in root.go didn't set it.
		// For now, let's try to get it via cfgService as a fallback, though ideally it should be set.
		var err error
		if cfgService == nil { // Ensure cfgService is initialized
			if appContext == nil { // Should not happen if first check passed
				return fmt.Errorf("appContext is nil, cannot initialize cfgService")
			}
			cfgService = service.NewConfigService(appContext)
		}
		configFilePath, err = cfgService.GetEffectiveConfigPath(cfgFile) // cfgFile is a global from root cmd
		if err != nil {
			return fmt.Errorf("failed to get effective config path: %w", err)
		}
		if configFilePath == "" {
			// If still empty, it implies no config file was found or specified,
			// which might be okay for some commands, but not for initializing a path-relative backend.
			slog.Warn("Config file path is empty; relative backend paths may not resolve correctly.")
			// Proceed, but NewStore might fail if the backend path is relative.
		}
	}

	slog.Debug("[InitActiveBackend] Using config file path for resolving relative backend paths", "configFilePath", configFilePath)
	configFileDir := ""
	if configFilePath != "" {
		configFileDir = filepath.Dir(configFilePath)
	}

	backendN := cfg.DefaultBackend
	if backendN == "" {
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
		activeBackend = nil
		activeBackendName = ""
		return fmt.Errorf("backend '%s' has an unsupported type '%s' for the create command", backendN, storageCfg.Type)
	}

	if storageCfg.LocalFS == nil {
		activeBackend = nil
		activeBackendName = ""
		return fmt.Errorf("localfs configuration for backend '%s' is missing", backendN)
	}

	// LocalFS path is now resolved inside localfs.NewStore using configFileDir
	storeSpecificConfig := *storageCfg.LocalFS

	// Pass configFileDir to localfs.NewStore
	localStore, err := localfs.NewStore(storeSpecificConfig, configFileDir)
	if err != nil {
		activeBackend = nil
		activeBackendName = ""
		return fmt.Errorf("failed to create new localfs store for backend '%s' (config path: %s, backend path: %s): %w", backendN, configFilePath, storeSpecificConfig.Path, err)
	}

	if err := localStore.Init(map[string]interface{}{"name": backendN}); err != nil { // Corrected quotes
		activeBackend = nil
		activeBackendName = ""
		return fmt.Errorf("failed to initialize localfs store for backend '%s' at %s: %w", backendN, storeSpecificConfig.Path, err)
	}

	activeBackend = localStore
	activeBackendName = backendN
	return nil
}

// GetActiveBackend returns the currently active and initialized storage backend and its name.
func GetActiveBackend() (storage.Backend, string) {
	return activeBackend, activeBackendName
}

// InitializeBackendFromConfig initializes a backend from the provided configuration
func InitializeBackendFromConfig(backendName string, backendConfig *model.StorageConfig) (storage.Backend, error) {
	if backendConfig == nil {
		return nil, fmt.Errorf("backend configuration for '%s' is nil", backendName)
	}

	if backendConfig.Type != "localfs" {
		return nil, fmt.Errorf("backend '%s' has an unsupported type '%s'", backendName, backendConfig.Type)
	}

	if backendConfig.LocalFS == nil {
		return nil, fmt.Errorf("localfs configuration for backend '%s' is missing", backendName)
	}

	// Use appContext.ConfigPath directly
	configFilePath := appContext.ConfigPath
	if configFilePath == "" {
		var err error
		if cfgService == nil {
			if appContext == nil {
				return nil, fmt.Errorf("appContext is nil, cannot initialize cfgService for InitializeBackendFromConfig")
			}
			cfgService = service.NewConfigService(appContext)
		}
		configFilePath, err = cfgService.GetEffectiveConfigPath(cfgFile) // cfgFile is a global from root cmd
		if err != nil {
			return nil, fmt.Errorf("failed to get effective config path for backend '%s': %w", backendName, err)
		}
		if configFilePath == "" {
			slog.Warn("Config file path is empty for InitializeBackendFromConfig; relative backend paths may not resolve correctly.", "backendName", backendName)
		}
	}

	configFileDir := ""
	if configFilePath != "" {
		configFileDir = filepath.Dir(configFilePath)
	}

	storeSpecificConfig := *backendConfig.LocalFS

	// Pass configFileDir to localfs.NewStore
	localStore, err := localfs.NewStore(storeSpecificConfig, configFileDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create new localfs store for backend '%s' (config path: %s, backend path: %s): %w", backendName, configFilePath, storeSpecificConfig.Path, err)
	}

	// Pass backendName to Init for the store to know its logical name
	if err := localStore.Init(map[string]interface{}{"name": backendName}); err != nil { // Corrected quotes
		return nil, fmt.Errorf("failed to initialize localfs store for backend '%s' at %s: %w", backendName, storeSpecificConfig.Path, err)
	}

	// localStore.SetName(backendName) // SetName is now implicitly handled by Init

	return localStore, nil
}
