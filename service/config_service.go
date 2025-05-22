package service

import (
	"fmt"
	"os"
	"path/filepath"

	"gydnc/model"
	"gydnc/util"
)

// ConfigService provides methods for managing configuration.
// It ensures backward compatibility while enabling service-oriented design.
type ConfigService struct {
	ctx *AppContext
}

// NewConfigService creates a new ConfigService with the provided context.
func NewConfigService(ctx *AppContext) *ConfigService {
	return &ConfigService{
		ctx: ctx,
	}
}

// InitConfig initializes a new configuration in the specified directory.
// Returns the path to the .gydnc directory and its config file.
// If the configuration already exists, it returns an error unless forceCreate is true.
func (s *ConfigService) InitConfig(targetDir string, backendType string, forceCreate bool) (string, error) {
	if targetDir == "" {
		var err error
		targetDir, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Create .gydnc directory
	gydncPath := filepath.Join(targetDir, ".gydnc")
	if _, err := os.Stat(gydncPath); err == nil && !forceCreate {
		return "", fmt.Errorf("guidance store already exists at %s", gydncPath)
	}

	if err := os.MkdirAll(gydncPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", gydncPath, err)
	}

	// Create a new config with explicit settings (no defaults)
	cfg := &model.Config{
		DefaultBackend:  "default_local",
		StorageBackends: make(map[string]*model.StorageConfig),
	}

	// Add the named backend
	storageConfig := &model.StorageConfig{
		Type: backendType,
	}

	if backendType == "localfs" {
		storageConfig.LocalFS = &model.LocalFSConfig{
			Path: gydncPath,
		}
	}

	cfg.StorageBackends["default_local"] = storageConfig

	// Save config
	configPath := filepath.Join(gydncPath, "config.yml")
	if err := s.SaveConfig(cfg, configPath); err != nil {
		return "", fmt.Errorf("failed to save config: %w", err)
	}

	return gydncPath, nil
}

// GetEffectiveConfigPath determines which configuration file to use based on the provided path
// or environment variables.
func (s *ConfigService) GetEffectiveConfigPath(cliConfigPath string) (string, error) {
	if cliConfigPath != "" {
		return cliConfigPath, nil
	}

	// Check environment variable
	envConfig := os.Getenv("GYDNC_CONFIG")
	if envConfig != "" {
		return envConfig, nil
	}

	// No configuration path available
	return "", fmt.Errorf("no config file specified via CLI or GYDNC_CONFIG environment variable")
}

// LoadConfig loads configuration from the specified path.
// A configuration must be explicitly provided either via CLI argument or environment variable.
func (s *ConfigService) LoadConfig(configPath string, requireConfig bool) (*model.Config, error) {
	effectiveConfigPath, err := s.GetEffectiveConfigPath(configPath)
	if err != nil {
		return nil, err
	}

	// Load from path
	return s.LoadFromPath(effectiveConfigPath, requireConfig)
}

// LoadFromPath loads configuration from a specific file path.
func (s *ConfigService) LoadFromPath(configFilePath string, requireConfig bool) (*model.Config, error) {
	// If no configuration file path is provided, always return an error
	if configFilePath == "" {
		return nil, fmt.Errorf("no config file found - configuration must be explicitly provided via CLI arg or GYDNC_CONFIG env var")
	}

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configFilePath, err)
	}

	return s.LoadConfigFromString(string(data))
}

// LoadConfigFromString parses configuration data from a string (useful for testing).
func (s *ConfigService) LoadConfigFromString(data string) (*model.Config, error) {
	cfg, err := util.LoadConfigYAML([]byte(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return cfg, nil
}

// SaveConfig writes the configuration to the specified path.
func (s *ConfigService) SaveConfig(cfg *model.Config, path string) error {
	if cfg == nil {
		return fmt.Errorf("cannot save a nil config")
	}
	if path == "" {
		return fmt.Errorf("config save path cannot be empty")
	}

	data, err := util.MarshalConfigYAML(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = os.WriteFile(path, data, 0600) // 0600 for read/write by owner only
	if err != nil {
		return fmt.Errorf("failed to write config file %s: %w", path, err)
	}
	return nil
}

// GetActiveStorageBackend returns the StorageConfig for the DefaultBackend.
func (s *ConfigService) GetActiveStorageBackend(cfg *model.Config) (*model.StorageConfig, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration is nil")
	}
	if cfg.DefaultBackend == "" {
		// It's okay to not have a default backend specified. Some commands might not need it.
		// The caller should handle this case if a backend is strictly required.
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
