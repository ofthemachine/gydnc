package config

import (
	"fmt"
	"os"
	"path/filepath"

	"log/slog"

	"gopkg.in/yaml.v3"
)

// GlobalConfig holds the application-wide configuration.
// It is populated by Load and accessed via Get.
var globalConfig *Config
var loadedConfigActualPath string // Store the actual path from which config was loaded

// LocalFSConfig defines settings specific to the local filesystem backend.
// For the MVP, Git integration settings are omitted and considered a future enhancement.
type LocalFSConfig struct {
	Path string `yaml:"path"`
}

// StorageConfig defines the configuration for a storage backend.
// Only one backend type (e.g., LocalFS) should be configured at a time per named backend instance.
// For MVP, we'll assume a simple structure where only one type of backend config is present.
// A more complex setup might use a map[string]interface{} and type assertions.
type StorageConfig struct {
	Type    string         `yaml:"type"`              // e.g., "localfs"
	LocalFS *LocalFSConfig `yaml:"localfs,omitempty"` // Pointer to allow omitempty
	// Other backend types like S3Config, DBConfig etc. would go here
}

// Config defines the structure of the gydnc.conf file.
// It supports multiple named storage backends.
type Config struct {
	DefaultBackend  string                    `yaml:"default_backend"`
	StorageBackends map[string]*StorageConfig `yaml:"storage_backends"`
	// Future global settings can go here, e.g., relating to canonicalization or hashing defaults
	// Canonicalization struct {
	// 	 HashAlgorithm string   `yaml:"hash_algorithm"`
	// 	 IncludeFields []string `yaml:"include_fields"`
	// } `yaml:"canonicalization"`
}

// NewDefaultConfig creates a config with some default values.
// This might be used if no config file is found.
func NewDefaultConfig() *Config {
	defaultBackendName := "implicit_cwd_store" // Name for the implicitly created backend
	cfg := &Config{
		DefaultBackend:  defaultBackendName,
		StorageBackends: make(map[string]*StorageConfig),
	}
	cfg.StorageBackends[defaultBackendName] = &StorageConfig{
		Type: "localfs",
		LocalFS: &LocalFSConfig{
			Path: ".", // Point to the current working directory
		},
	}
	return cfg
}

// Load reads the configuration from the specified path or environment variable.
// Priority: --config CLI arg, then GYDNC_CONFIG env var.
// If cliConfigPath is empty, it checks the environment variable.
// If both are empty or file not found, it returns a default config and no error (or specific error if needed).
func Load(cliConfigPath string) (*Config, error) {
	configFilePath := cliConfigPath

	if configFilePath == "" {
		configFilePath = os.Getenv("GYDNC_CONFIG")
	}

	// If still no path, try ~/.gydnc/config.yml
	if configFilePath == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			homeGydncPath := filepath.Join(homeDir, ".gydnc", "config.yml")
			if _, err := os.Stat(homeGydncPath); err == nil {
				configFilePath = homeGydncPath
				slog.Debug("No explicit config path, using ~/.gydnc/config.yml", "path", configFilePath)
			}
		}
	}

	if configFilePath == "" {
		slog.Debug("No configuration file path specified or found, using default configuration.")
		cfg := NewDefaultConfig()
		globalConfig = cfg
		loadedConfigActualPath = "" // Explicitly clear if using default
		return cfg, nil
	}

	slog.Debug("Attempting to load configuration from", "path", configFilePath)
	loadedConfigActualPath = configFilePath // Store path *before* attempting to read

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		slog.Warn("Failed to read config file, using default configuration.", "path", configFilePath, "error", err)
		cfg := NewDefaultConfig()
		globalConfig = cfg
		loadedConfigActualPath = "" // Using defaults, so no "actual loaded path"
		return cfg, fmt.Errorf("failed to read config file %s: %w. Using default config", configFilePath, err)
	}

	cfgFromFile, err := LoadConfigFromString(string(data))
	if err != nil {
		slog.Warn("Failed to parse config file, using default configuration.", "path", configFilePath, "error", err)
		cfg := NewDefaultConfig()
		globalConfig = cfg
		loadedConfigActualPath = "" // Using defaults, so no "actual loaded path"
		return cfg, fmt.Errorf("failed to parse config file %s: %w. Using default config", configFilePath, err)
	}

	globalConfig = cfgFromFile
	// loadedConfigActualPath was already set to configFilePath if we reached here successfully.
	return cfgFromFile, nil
}

// LoadConfigFromString parses configuration data from a string (useful for testing).
func LoadConfigFromString(data string) (*Config, error) {
	var cfg Config
	// Initialize map if it's nil, to avoid panic during unmarshal if storage_backends is empty
	cfg.StorageBackends = make(map[string]*StorageConfig)

	err := yaml.Unmarshal([]byte(data), &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config data: %w", err)
	}
	return &cfg, nil
}

// Save writes the current configuration to the specified path.
// This is primarily for `gydnc config set` and `gydnc init`.
func Save(cfg *Config, path string) error {
	if cfg == nil {
		return fmt.Errorf("cannot save a nil config")
	}
	if path == "" {
		return fmt.Errorf("config save path cannot be empty")
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Ensure directory exists if path includes directories
	// dir := filepath.Dir(path)
	// if _, err := os.Stat(dir); os.IsNotExist(err) {
	// 	 if err := os.MkdirAll(dir, 0750); err != nil {
	// 	 	 return fmt.Errorf("failed to create directory for config file %s: %w", dir, err)
	// 	 }
	// }

	// For MVP, keep it simple and assume `gydnc init` might create a root .gydnc.conf
	// or `gydnc config set` operates on an existing one.
	// Directory creation logic can be added if needed for more complex scenarios.

	err = os.WriteFile(path, data, 0600) // 0600 for read/write by owner only
	if err != nil {
		return fmt.Errorf("failed to write config file %s: %w", path, err)
	}
	return nil
}

// Get returns the loaded global configuration.
// It panics if Load has not been called successfully.
func Get() *Config {
	if globalConfig == nil {
		// This state should ideally be prevented by ensuring Load is called early,
		// or by having Load return a default config instead of erroring out completely
		// if no file is found (which it currently does).
		// For now, let's ensure Load always sets globalConfig, even to a default.
		panic("config not loaded; Load() must be called before Get()")
	}
	return globalConfig
}

// GetLoadedConfigActualPath returns the actual file path from which the global configuration was loaded.
// Returns an empty string if the configuration was not loaded from a file (e.g., using defaults).
func GetLoadedConfigActualPath() string {
	return loadedConfigActualPath
}

// GetActiveStorageBackend returns the StorageConfig for the DefaultBackend.
// It returns an error if the DefaultBackend is not defined or not found in StorageBackends.
func GetActiveStorageBackend(cfg *Config) (*StorageConfig, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration is nil")
	}
	if cfg.DefaultBackend == "" {
		// It's okay to not have a default backend specified. Some commands might not need it.
		// The caller should handle this case if a backend is strictly required.
		// slog.Debug("No DefaultBackend specified in configuration.")
		return nil, fmt.Errorf("notice: No DefaultBackend specified in configuration. Some commands may not function")
	}
	backendConfig, ok := cfg.StorageBackends[cfg.DefaultBackend]
	if !ok {
		return nil, fmt.Errorf("default backend '%s' not found in storage_backends configuration", cfg.DefaultBackend)
	}
	if backendConfig == nil { // Should not happen if key exists, but good practice
		return nil, fmt.Errorf("configuration for default backend '%s' is nil", cfg.DefaultBackend)
	}
	return backendConfig, nil
}

// Example gydnc.conf content:
// default_backend: "prodLocal"
// storage_backends:
//   prodLocal:
//     type: "localfs"
//     localfs:
//       path: "./guidance_prod"
//   stagingLocal:
//     type: "localfs"
//     localfs:
//       path: "./guidance_staging"
