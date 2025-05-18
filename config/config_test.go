package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigLoadPriority(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	workDir := filepath.Join(tmpDir, "work")

	// Create directories
	if err := os.MkdirAll(filepath.Join(homeDir, ".gydnc"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(workDir, ".gydnc"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create test config files with different content to identify which one is loaded
	configs := map[string]string{
		filepath.Join(homeDir, ".gydnc/config.yml"):  "default_backend: home",
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

	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Change to the test working directory
	if err := os.Chdir(workDir); err != nil {
		t.Fatal(err)
	}

	// Set test home directory
	os.Setenv("HOME", homeDir)

	tests := []struct {
		name          string
		cliArg        string
		envVar        string
		expectedValue string
	}{
		{
			name:          "CLI argument takes precedence",
			cliArg:        filepath.Join(tmpDir, "explicit_config.yml"),
			envVar:        filepath.Join(tmpDir, "env_var_config.yml"),
			expectedValue: "explicit",
		},
		{
			name:          "Environment variable is second priority",
			cliArg:        "",
			envVar:        filepath.Join(tmpDir, "env_var_config.yml"),
			expectedValue: "envvar",
		},
		{
			name:          "Home directory is last priority",
			cliArg:        "",
			envVar:        "",
			expectedValue: "home",
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

			// For the home directory fallback test, remove the CWD .gydnc/config.yml
			if tt.name == "Home directory is last priority" {
				t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))
				// For this specific test case, also ensure GYDNC_CONFIG is NOT set
				t.Setenv("GYDNC_CONFIG", "") // Unset it explicitly for the test

				// TODO: This if block was empty and caused a linting error (SA9003: empty branch).
				// If there was intended logic here, it needs to be implemented.
				// if tt.name == "Home directory is last priority" {
				// }
			}

			// Load config
			cfg, err := Load(tt.cliArg, false)
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

	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Change to the test directory and set it as home
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	os.Setenv("HOME", tmpDir)
	os.Unsetenv("GYDNC_CONFIG")

	// Load config with no files present
	cfg, err := Load("", false)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg == nil {
		t.Error("Load() returned nil config")
		return
	}
	if cfg.StorageBackends == nil {
		t.Error("Load() returned config with nil StorageBackends")
		return
	}
	if len(cfg.StorageBackends) != 0 {
		t.Error("Load() returned config with non-empty StorageBackends")
	}
}
