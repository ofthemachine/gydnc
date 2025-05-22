package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"gydnc/internal/logging"
	"gydnc/service"
)

var (
	cfgFile    string
	verbosity  int
	quiet      bool
	appContext *service.AppContext // Exposed to be used by other files in cmd package
)

var rootCmd = &cobra.Command{
	Use:   "gydnc",
	Short: "A tool for managing guidance documents",
	Long: `gydnc streamlines the creation, management, and discovery of guidance documents.
It aids in creating and maintaining documentation tailored to agent behavior and LLM prompting.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is empty, load via GYDNC_CONFIG env var or explicit path)")
	rootCmd.PersistentFlags().CountVarP(&verbosity, "verbose", "v", "Increase logging verbosity (default: WARN, -v: INFO, -vv: DEBUG)")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress non-error log messages (equivalent to log level ERROR)")

	rootCmd.AddCommand(llmCmd) // llmCmd is defined in llm.go
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Set up logging based on verbosity/quiet flags
	logging.SetupLogger(verbosity, quiet)

	// Determine if the current command is 'init' or 'version' (bootstrap commands)
	requireConfig := true
	cmdName := ""
	if len(os.Args) > 1 {
		cmdName = os.Args[1]
		if cmdName == "init" || cmdName == "version" {
			requireConfig = false
		}
	}

	// For commands that don't require config (init, version), exit early
	if !requireConfig {
		return
	}

	// Create app context and config service
	appContext = service.NewAppContext(nil, nil)
	configService := service.NewConfigService(appContext)

	// Load config using the service layer
	configPath, err := configService.GetEffectiveConfigPath(cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "active backend not initialized; run 'gydnc init' or check config\n")
		os.Exit(1)
	}

	config, err := configService.LoadFromPath(configPath, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "active backend not initialized; run 'gydnc init' or check config\n")
		os.Exit(1)
	}

	// Update the app context with the loaded config
	appContext.Config = config

	// Initialize the active backend
	if err := InitActiveBackend(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not initialize active backend: %v\n", err)
	}
}
