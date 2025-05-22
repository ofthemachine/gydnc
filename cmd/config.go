package cmd

import (
	"fmt"
	"log/slog"

	"gydnc/service"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage gydnc configuration (View implemented, Set/Get not implemented in MVP)",
	Long:  `Allows viewing and modifying the gydnc configuration.`,
}

var configViewCmd = &cobra.Command{
	Use:   "view",
	Short: "View the current gydnc configuration",
	Long:  `Prints the currently loaded gydnc configuration to standard output in YAML format.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		slog.Debug("Starting 'config view' command execution")

		// Check if app context is initialized
		if appContext == nil || appContext.Config == nil {
			return fmt.Errorf("configuration not loaded; run 'gydnc init' or check config")
		}

		// Get config from app context
		cfg := appContext.Config

		// Create a config service to get the effective config path
		configService := service.NewConfigService(appContext)
		loadedPath, err := configService.GetEffectiveConfigPath(cfgFile)
		if err == nil && loadedPath != "" {
			fmt.Printf("# Configuration loaded from: %s\n", loadedPath)
		} else {
			fmt.Println("# Configuration is using default values (not loaded from a file).")
		}
		fmt.Println("# ---") // Separator

		yamlData, err := yaml.Marshal(cfg)
		if err != nil {
			slog.Error("Failed to marshal current config to YAML", "error", err)
			return fmt.Errorf("failed to marshal config to YAML: %w", err)
		}
		fmt.Println(string(yamlData))
		slog.Debug("'config view' command finished successfully")
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a specific configuration value (Not implemented in MVP)",
	Long:  `Retrieves and displays a specific configuration value by its key. Not implemented in MVP.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		slog.Info("'config get' command called", "key", key)
		fmt.Printf("Command 'config get %s' is not implemented in MVP.\n", key)
		return fmt.Errorf("command 'config get' not implemented in MVP")
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a specific configuration value (Not implemented in MVP)",
	Long:  `Sets a configuration value by its key. Not implemented in MVP.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]
		slog.Info("'config set' command called", "key", key, "value", value)
		fmt.Printf("Command 'config set %s %s' is not implemented in MVP.\n", key, value)
		return fmt.Errorf("command 'config set' not implemented in MVP")
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configViewCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)

	// Flags for config set/get could be added here, e.g. --global for user-level config vs project config.
}
