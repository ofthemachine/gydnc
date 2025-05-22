package util

import (
	"fmt"
	"gydnc/model"

	"gopkg.in/yaml.v3"
)

// LoadConfigYAML unmarshals YAML data into a Config struct.
func LoadConfigYAML(data []byte) (*model.Config, error) {
	var cfg model.Config
	// Initialize map if it's nil, to avoid panic during unmarshal if storage_backends is empty
	cfg.StorageBackends = make(map[string]*model.StorageConfig)

	err := yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config data: %w", err)
	}
	return &cfg, nil
}

// MarshalConfigYAML marshals a Config struct into YAML data.
func MarshalConfigYAML(cfg *model.Config) ([]byte, error) {
	if cfg == nil {
		return nil, fmt.Errorf("cannot marshal nil config")
	}

	return yaml.Marshal(cfg)
}
