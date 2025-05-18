package cmd

import (
	"fmt"
	// "log/slog" // Removed for temporary debug setup
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
	SilenceErrors: false,
	SilenceUsage:  false,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err) // Print error to stderr
		os.Exit(1)
	}
}

func init() {
	// --- Temporary Slog Debug Setup WAS HERE --- REMOVED
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
	// Determine if the current command is 'init' or 'version' (bootstrap commands)
	requireConfig := true
	if len(os.Args) > 1 {
		cmd := os.Args[1]
		if cmd == "init" || cmd == "version" {
			requireConfig = false
		}
	}
	_, err := config.Load(cfgFile, requireConfig) // Pass requireConfig
	if err != nil {
		fmt.Fprintf(os.Stderr, "active backend not initialized; run 'gydnc init' or check config\n")
		os.Exit(1)
	}

	// Initialize the active backend after loading the configuration.
	// config.Get() should now be safe from panic.
	if err := InitActiveBackend(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not initialize active backend: %v\n", err)
	}
}
