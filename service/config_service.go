package service

import (
	"fmt"
	"os"
	"path/filepath"

	"gydnc/config"
)

// ConfigService provides methods for managing configuration.
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

	// Create default config
	cfg := &config.Config{
		DefaultBackend:  "default_local",
		StorageBackends: make(map[string]*config.StorageConfig),
	}

	// Add the default backend
	storageConfig := &config.StorageConfig{
		Type: backendType,
	}

	if backendType == "localfs" {
		storageConfig.LocalFS = &config.LocalFSConfig{
			Path: gydncPath,
		}
	}

	cfg.StorageBackends["default_local"] = storageConfig

	// Save config
	configPath := filepath.Join(gydncPath, "config.yml")
	if err := config.Save(cfg, configPath); err != nil {
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
