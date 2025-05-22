package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigLoadPriority(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir := t.TempDir()
	workDir := filepath.Join(tmpDir, "work")

	// Create directories
	if err := os.MkdirAll(filepath.Join(workDir, ".gydnc"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create test config files with different content to identify which one is loaded
	configs := map[string]string{
		filepath.Join(tmpDir, "explicit_config.yml"): "default_backend: explicit",
		filepath.Join(tmpDir, "env_var_config.yml"):  "default_backend: envvar",
	}

	for path, content := range configs {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Save original environment and working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(originalWd) }()

	// Change to the test working directory
	if err := os.Chdir(workDir); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		cliArg        string
		envVar        string
		expectedValue string
		expectError   bool
	}{
		{
			name:          "CLI argument takes precedence",
			cliArg:        filepath.Join(tmpDir, "explicit_config.yml"),
			envVar:        filepath.Join(tmpDir, "env_var_config.yml"),
			expectedValue: "explicit",
			expectError:   false,
		},
		{
			name:          "Environment variable is used if CLI arg not provided",
			cliArg:        "",
			envVar:        filepath.Join(tmpDir, "env_var_config.yml"),
			expectedValue: "envvar",
			expectError:   false,
		},
		{
			name:          "Error when no CLI arg or env var provided",
			cliArg:        "",
			envVar:        "",
			expectedValue: "",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset environment between tests
			if tt.envVar != "" {
				os.Setenv("GYDNC_CONFIG", tt.envVar)
			} else {
				os.Unsetenv("GYDNC_CONFIG")
			}

			// Load config
			cfg, err := Load(tt.cliArg, false)

			if tt.expectError {
				if err == nil {
					t.Errorf("Load() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			if cfg.DefaultBackend != tt.expectedValue {
				t.Errorf("Load() got default_backend = %v, want %v", cfg.DefaultBackend, tt.expectedValue)
			}
		})
	}
}

func TestConfigLoadNoConfig(t *testing.T) {
	// Create a clean temporary directory
	tmpDir := t.TempDir()

	// Save original environment and working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(originalWd) }()

	// Change to the test directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	os.Unsetenv("GYDNC_CONFIG")

	// Load config with no files present - should error
	cfg, err := Load("", false)

	// Verify an error is returned
	if err == nil {
		t.Error("Load() should have returned an error when no config provided")
	}

	// Verify the config is nil
	if cfg != nil {
		t.Error("Load() should have returned nil config when no config provided")
	}
}
