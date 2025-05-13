package cmd

import (
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"
)

//go:embed llm.txt
var llmGuidanceContent string

var llmCmd = &cobra.Command{
	Use:   "llm",
	Short: "Provides instructions for LLM interaction with gydnc (internal use)",
	Long: `Outputs the standard interaction protocol for LLMs interacting with the Gydnc service.
This is primarily intended for internal use during development and testing.
It prints the expected CLI usage and interaction flow for guidance management.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Slog can be added later if logging is integrated into the root command setup
		fmt.Print(llmGuidanceContent)
		return nil
	},
}

func init() {
	// Command registration is handled in root.go
	// rootCmd.AddCommand(llmCmd)
}
