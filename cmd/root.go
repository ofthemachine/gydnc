package cmd

import (
	"fmt"
	"os"

	"gydnc/config"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	// Weitere globale Flags hier, z.B. backendName
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gydnc",
	Short: "A CLI tool for managing content-addressable guidance.",
	Long: `gydnc is a command-line interface for creating, managing, and retrieving
guidance entities. It supports various backends and aims to provide
a robust system for AI guidance versioning and discovery.`,
	SilenceErrors: true,
	SilenceUsage:  true,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err) // Use fmt.Fprintln to os.Stderr for errors
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is empty, load via GYDNC_CONFIG env var or explicit path)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// Example for a global backend flag - to be wired into specific commands later
	// rootCmd.PersistentFlags().StringVar(&backendName, "backend", "", "Name of the storage backend to use (defined in config)")

	rootCmd.AddCommand(llmCmd) // llmCmd is defined in llm.go
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// config.Load will handle the logic of checking cfgFile (from --config flag),
	// then GYDNC_CONFIG env var, and then loading defaults if neither is present.
	_, err := config.Load(cfgFile) // cfgFile is populated by the --config persistent flag
	if err != nil {
		// It might be too early to fatal here. config.Load itself might return a default config
		// and an error if a specific file wasn't found but defaults are usable.
		// Or, some commands might not need a config file at all.
		// For now, let's print the error if one occurs during explicit load attempt.
		// Specific commands should check if the config they need is available.
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		// Potentially os.Exit(1) if config is absolutely mandatory for all operations from the start,
		// but many CLIs allow some operations without full config.
	}

	// Initialize the active backend after loading the configuration.
	if err := InitActiveBackend(); err != nil {
		// An error here means the backend specified in the config could not be initialized.
		// Some commands (like `version`, `init` itself, or `config view`) might still work.
		// Other commands that rely on a backend will fail later if GetActiveBackend() returns nil
		// or if they try to use a non-functional backend.
		// We print a warning but don't exit, allowing the CLI to proceed for commands
		// that don't require a fully initialized backend.
		fmt.Fprintf(os.Stderr, "Warning: could not initialize active backend: %v\n", err)
	}
}
