package service

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigService_InitConfig(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create a test application context
	ctx := NewAppContext(nil, nil)

	// Create the config service
	service := NewConfigService(ctx)

	tests := []struct {
		name        string
		targetDir   string
		backendType string
		forceCreate bool
		wantErr     bool
	}{
		{
			name:        "create new config in current directory",
			targetDir:   tmpDir,
			backendType: "localfs",
			forceCreate: false,
			wantErr:     false,
		},
		{
			name:        "fail if config already exists",
			targetDir:   tmpDir, // Same directory as first test
			backendType: "localfs",
			forceCreate: false,
			wantErr:     true,
		},
		{
			name:        "force create if config exists",
			targetDir:   tmpDir, // Same directory as before
			backendType: "localfs",
			forceCreate: true,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath, err := service.InitConfig(tt.targetDir, tt.backendType, tt.forceCreate)

			if (err != nil) != tt.wantErr {
				t.Errorf("InitConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Check that the config file was created
				configPath := filepath.Join(gotPath, "config.yml")
				if _, err := os.Stat(configPath); os.IsNotExist(err) {
					t.Errorf("Config file not created at %s", configPath)
				}
			}
		})
	}
}

func TestConfigService_GetEffectiveConfigPath(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")

	// Create a test file
	if err := os.WriteFile(configPath, []byte("# Test config"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a test application context
	ctx := NewAppContext(nil, nil)

	// Create the config service
	service := NewConfigService(ctx)

	// Save original environment
	originalEnv := os.Getenv("GYDNC_CONFIG")
	defer os.Setenv("GYDNC_CONFIG", originalEnv)

	tests := []struct {
		name          string
		cliConfigPath string
		envConfigPath string
		wantPath      string
		wantErr       bool
	}{
		{
			name:          "CLI path takes precedence",
			cliConfigPath: configPath,
			envConfigPath: "/non/existent/path",
			wantPath:      configPath,
			wantErr:       false,
		},
		{
			name:          "Environment variable used if CLI path not provided",
			cliConfigPath: "",
			envConfigPath: configPath,
			wantPath:      configPath,
			wantErr:       false,
		},
		{
			name:          "Error if no path available",
			cliConfigPath: "",
			envConfigPath: "",
			wantPath:      "",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envConfigPath != "" {
				os.Setenv("GYDNC_CONFIG", tt.envConfigPath)
			} else {
				os.Unsetenv("GYDNC_CONFIG")
			}

			gotPath, err := service.GetEffectiveConfigPath(tt.cliConfigPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetEffectiveConfigPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if gotPath != tt.wantPath {
				t.Errorf("GetEffectiveConfigPath() = %v, want %v", gotPath, tt.wantPath)
			}
		})
	}
}
