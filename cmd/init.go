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
	// defaultGydncDirName is now the store directory itself
	defaultStoreDirName   = ".gydnc"
	defaultConfigFileName = "config.yml"
	// defaultStoreSubDirName is no longer used as .gydnc is the store
	// defaultTagOntologyInDirFile is no longer used, it's at the root
	defaultTagOntologyFileName = "TAG_ONTOLOGY.md"
	defaultBackendName         = "defaultLocal"
)

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a new gydnc repository and configuration in the specified path or current directory",
	Long: `Creates a configuration file (config.yml) and TAG_ONTOLOGY.md at the root of the target path,
and a .gydnc directory to serve as the default guidance store.
If a path is provided, initialization occurs there. Otherwise, it uses the current directory.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		slog.Debug("Starting 'init' command execution with new structure")

		var targetBasePath string
		var err error

		if len(args) > 0 {
			targetBasePath, err = filepath.Abs(args[0])
			if err != nil {
				return fmt.Errorf("failed to get absolute path for target directory '%s': %w", args[0], err)
			}
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

		configFilePath := filepath.Join(targetBasePath, defaultConfigFileName)
		guidanceStorePath := filepath.Join(targetBasePath, defaultStoreDirName)
		tagOntologyFilePath := filepath.Join(targetBasePath, defaultTagOntologyFileName)

		slog.Info("Initializing gydnc repository structure",
			"config_file", configFilePath,
			"store_dir", guidanceStorePath,
			"tag_ontology", tagOntologyFilePath)

		// Check if config file already exists to prevent accidental overwrite
		if _, err := os.Stat(configFilePath); err == nil {
			return fmt.Errorf("gydnc already initialized: '%s' exists. Use --force to overwrite (not implemented).", configFilePath)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("failed to check for existing '%s' file: %w", configFilePath, err)
		}

		// Create guidance store directory (.gydnc)
		if err := os.MkdirAll(guidanceStorePath, 0755); err != nil {
			return fmt.Errorf("failed to create guidance store directory '%s': %w", guidanceStorePath, err)
		}
		fmt.Printf("Created guidance store: %s\n", guidanceStorePath)

		// Create TAG_ONTOLOGY.md at the target base path
		tagOntologyContent := []byte("# Tag Ontology\n\nThis file defines the taxonomy of tags used for guidance entities.\n\n## Categories\n\n- category:example_category\n  description: An example category for tags.\n\n## Tags\n\n- tag:example_tag\n  description: An example tag.\n  category: example_category\n")
		if err := os.WriteFile(tagOntologyFilePath, tagOntologyContent, 0644); err != nil {
			return fmt.Errorf("failed to create TAG_ONTOLOGY.md at '%s': %w", tagOntologyFilePath, err)
		}
		fmt.Printf("Created TAG_ONTOLOGY.md: %s\n", tagOntologyFilePath)

		// Create default config.yml at the target base path
		newCfg := config.NewDefaultConfig()
		newCfg.DefaultBackend = defaultBackendName
		// The path for LocalFS should be the absolute path to the guidanceStorePath
		absGuidanceStorePath, err := filepath.Abs(guidanceStorePath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for guidance store '%s': %w", guidanceStorePath, err)
		}
		newCfg.StorageBackends[defaultBackendName] = &config.StorageConfig{
			Type: "localfs",
			LocalFS: &config.LocalFSConfig{
				Path: absGuidanceStorePath, // Use absolute path in config for clarity
			},
		}

		if err := config.Save(newCfg, configFilePath); err != nil {
			return fmt.Errorf("failed to save configuration file '%s': %w", configFilePath, err)
		}
		fmt.Printf("Created configuration file: %s\n", configFilePath)

		slog.Info("'init' command finished successfully.")
		fmt.Printf("gydnc initialized successfully in %s\n", targetBasePath)
		fmt.Println("\nTo activate this configuration for your current session, you can run:")
		fmt.Printf("  export GYDNC_CONFIG=\"%s\"\n", configFilePath)
		fmt.Println("Consider adding this line to your shell configuration file (e.g., ~/.zshrc or ~/.bashrc) for persistent use.")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	// initCmd.Flags().Bool("force", false, "Overwrite existing configuration if found")
}
