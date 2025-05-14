package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"gydnc/config"

	"github.com/spf13/cobra"
	// "os/exec" // For optional git init
)

const (
	defaultGuidancePath    = ".gydnc_store"
	defaultConfigFileName  = "gydnc.conf"
	defaultTagOntologyFile = "TAG_ONTOLOGY.md"
	defaultBackendName     = "defaultLocal"
)

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a new gydnc repository and configuration in the specified path or current directory",
	Long: `Creates a default configuration file (gydnc.conf), a local guidance
storage directory (.gydnc_store), and a TAG_ONTOLOGY.md file within it.
If a path is provided, initialization occurs there. Otherwise, it uses the current directory.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		slog.Debug("Starting 'init' command execution")

		var targetBasePath string
		var err error

		if len(args) > 0 {
			targetBasePath, err = filepath.Abs(args[0])
			if err != nil {
				return fmt.Errorf("failed to get absolute path for target directory '%s': %w", args[0], err)
			}
			// Ensure the target base path itself exists if specified
			if err := os.MkdirAll(targetBasePath, 0755); err != nil {
				return fmt.Errorf("failed to create target directory '%s': %w", targetBasePath, err)
			}
		} else {
			targetBasePath, err = os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current working directory: %w", err)
			}
		}
		slog.Info("Target base path for initialization set", "path", targetBasePath)

		guidanceStorePath := filepath.Join(targetBasePath, defaultGuidancePath)
		configFilePath := filepath.Join(targetBasePath, defaultConfigFileName)
		tagOntologyFilePath := filepath.Join(guidanceStorePath, defaultTagOntologyFile)

		slog.Info("Initializing gydnc repository", "config_path", configFilePath, "store_path", guidanceStorePath)

		// Check if config file already exists
		if _, err := os.Stat(configFilePath); err == nil {
			return fmt.Errorf("configuration file '%s' already exists. Use --force to overwrite (not implemented).", configFilePath)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("failed to check for existing config file '%s': %w", configFilePath, err)
		}

		// Create guidance store directory (MkdirAll on targetBasePath above should handle base, this handles .gydnc_store)
		if err := os.MkdirAll(guidanceStorePath, 0755); err != nil {
			return fmt.Errorf("failed to create guidance store directory '%s': %w", guidanceStorePath, err)
		}
		fmt.Printf("Created guidance store directory: %s\n", guidanceStorePath)

		// Create TAG_ONTOLOGY.md
		tagOntologyContent := []byte("# Tag Ontology\n\nThis file defines the taxonomy of tags used for guidance entities.\n\n## Categories\n\n- category:example_category\n  description: An example category for tags.\n\n## Tags\n\n- tag:example_tag\n  description: An example tag.\n  category: example_category\n")
		if err := os.WriteFile(tagOntologyFilePath, tagOntologyContent, 0644); err != nil {
			return fmt.Errorf("failed to create TAG_ONTOLOGY.md at '%s': %w", tagOntologyFilePath, err)
		}
		fmt.Printf("Created TAG_ONTOLOGY.md: %s\n", tagOntologyFilePath)

		// Create default gydnc.conf
		newCfg := config.NewDefaultConfig()
		newCfg.DefaultBackend = defaultBackendName
		newCfg.StorageBackends[defaultBackendName] = &config.StorageConfig{
			Type: "localfs",
			LocalFS: &config.LocalFSConfig{
				Path: guidanceStorePath, // Store absolute path to the store
			},
		}

		if err := config.Save(newCfg, configFilePath); err != nil {
			return fmt.Errorf("failed to save configuration file '%s': %w", configFilePath, err)
		}
		fmt.Printf("Created configuration file: %s\n", configFilePath)

		// Optional: Initialize Git repository in the store path (remains commented for MVP)
		// ...

		slog.Info("'init' command finished successfully.")
		fmt.Printf("gydnc initialized successfully in %s\n", targetBasePath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	// initCmd.Flags().Bool("force", false, "Overwrite existing configuration if found")
}
