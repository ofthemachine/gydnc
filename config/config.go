package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// GlobalConfig holds the application-wide configuration.
// It is populated by Load and accessed via Get.
var globalConfig *Config

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
	return &Config{
		// DefaultBackend: "defaultLocal", // Example, might require a defaultLocal to be defined
		StorageBackends: make(map[string]*StorageConfig),
	}
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

	if configFilePath == "" {
		// No explicit config path provided, return a basic default config.
		// Commands requiring specific paths (like backend path) will need to handle this.
		// fmt.Println("No config file specified via --config or GYDNC_CONFIG. Using defaults.")
		cfg := NewDefaultConfig()
		globalConfig = cfg
		return cfg, nil
	}

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configFilePath, err)
	}

	cfg, err := LoadConfigFromString(string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configFilePath, err)
	}
	globalConfig = cfg
	return cfg, nil
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
