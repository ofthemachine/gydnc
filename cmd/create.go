package cmd

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

var (
	createTitle string
	// createTags  []string // For future use
	// createEditor bool     // For future use: open in editor
)

var createCmd = &cobra.Command{
	Use:   "create [alias]",
	Short: "Create a new guidance entity (Not implemented in MVP)",
	Long: `This command will eventually allow for creating new guidance entities interactively
or from a template. In the current MVP, it is not implemented.`,
	Args: cobra.ExactArgs(1), // Requires an alias
	RunE: func(cmd *cobra.Command, args []string) error {
		alias := args[0]
		slog.Info("'create' command called", "alias", alias, "title", createTitle)
		fmt.Printf("Command 'create' for alias '%s' with title '%s' is not implemented in MVP.\n", alias, createTitle)
		return fmt.Errorf("command 'create' not implemented in MVP")
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringVarP(&createTitle, "title", "t", "", "Title for the new guidance entity (required)")
	// _ = createCmd.MarkFlagRequired("title") // Mark as required once implemented
	// createCmd.Flags().StringSliceVarP(&createTags, "tags", "g", []string{}, "Comma-separated tags for the new guidance")
	// createCmd.Flags().BoolVarP(&createEditor, "editor", "e", false, "Open in default editor after creation")
}
