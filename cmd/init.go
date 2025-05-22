package cmd

import (
	_ "embed"
	"fmt"
	"gydnc/service"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

//go:embed init_tag_ontology.md
var tagOntologyContent []byte

const (
	defaultBackendName         = "default"
	defaultBackendType         = "localfs"
	defaultTagOntologyFileName = "tag_ontology.md"
)

var (
	forceInit bool
)

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a new gydnc repository and configuration in the specified path or current directory",
	Long: `Creates a configuration file and tag ontology in the .gydnc directory of the target path.
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

		// Create a temporary app context for the init command
		// This doesn't depend on any existing config
		ctx := service.NewAppContext(nil, nil)
		configService := service.NewConfigService(ctx)

		// Initialize the config using our service
		gydncDirPath, err := configService.InitConfig(targetBasePath, defaultBackendType, forceInit)
		if err != nil {
			return err
		}

		fmt.Printf("Created guidance store: %s\n", gydncDirPath)

		// Create tag_ontology.md directly in the init command
		tagOntologyPath := filepath.Join(gydncDirPath, defaultTagOntologyFileName)
		if err := os.WriteFile(tagOntologyPath, tagOntologyContent, 0644); err != nil {
			return fmt.Errorf("failed to create tag_ontology.md at '%s': %w", tagOntologyPath, err)
		}
		fmt.Printf("Created tag_ontology.md: %s\n", tagOntologyPath)

		configFilePath := filepath.Join(gydncDirPath, "config.yml")
		fmt.Printf("Created configuration file: %s\n", configFilePath)

		fmt.Printf("gydnc initialized successfully in %s\n", targetBasePath)
		fmt.Println("\nTo activate this configuration for your current session, you can run:")
		fmt.Printf("  export GYDNC_CONFIG=\"%s\"\n", configFilePath)
		fmt.Println("Consider adding this line to your shell configuration file (e.g., ~/.zshrc or ~/.bashrc) for persistent use.")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&forceInit, "force", false, "Overwrite existing configuration if found")
}
