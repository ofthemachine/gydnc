package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LocalFSSettings mirrors the 'settings' for a 'localfs' backend in config.yaml
type LocalFSSettings struct {
	RootDir           string   `yaml:"rootDir"`
	GuidanceLocations []string `yaml:"guidanceLocations"`
	// Git settings were removed as per user request.
}

// BackendConfig mirrors the structure of a single backend entry in config.yaml
type BackendConfig struct {
	Name     string          `yaml:"name"`
	Type     string          `yaml:"type"`
	Settings LocalFSSettings `yaml:"settings"`
}

// Config mirrors the structure of the entire gydnc/config.yaml file
type Config struct {
	Backends []BackendConfig `yaml:"backends"`
}

// Load reads and parses the gydnc configuration file from the given path.
func Load(configPath string) (Config, error) {
	var cfg Config

	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		return cfg, fmt.Errorf("reading config file '%s': %w", configPath, err)
	}

	err = yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("unmarshalling YAML from '%s': %w", configPath, err)
	}

	// Basic validation: ensure there's at least one backend configured.
	if len(cfg.Backends) == 0 {
		return cfg, fmt.Errorf("no backends configured in '%s'", configPath)
	}

	for i, backend := range cfg.Backends {
		if backend.Name == "" {
			return cfg, fmt.Errorf("backend %d in '%s' is missing a name", i, configPath)
		}
		if backend.Type == "" {
			return cfg, fmt.Errorf("backend '%s' in '%s' is missing a type", backend.Name, configPath)
		}
		if backend.Type == "localfs" {
			if backend.Settings.RootDir == "" {
				return cfg, fmt.Errorf("localfs backend '%s' in '%s' is missing 'rootDir' in settings", backend.Name, configPath)
			}
			if len(backend.Settings.GuidanceLocations) == 0 {
				return cfg, fmt.Errorf("localfs backend '%s' in '%s' has no 'guidanceLocations' in settings", backend.Name, configPath)
			}
		}
		// Add validation for other backend types if introduced
	}

	return cfg, nil
}

// Global verbosity and quiet state, managed by logging setup.
// These are placeholders; actual implementation will be in internal/logging.
var globalVerbosity int = 0
var globalQuiet bool = false

func SetVerbosity(level int) {
	globalVerbosity = level
}

func GetVerbosity() int {
	return globalVerbosity
}

func SetQuiet(quiet bool) {
	globalQuiet = quiet
}

func IsQuiet() bool {
	return globalQuiet
}
