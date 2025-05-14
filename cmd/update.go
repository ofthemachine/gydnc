package cmd

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

// var ( // Flags for update, e.g., for changing title, tags, or opening in editor
// 	updateTitle string
// 	updateTags  []string
// 	updateEditor bool
// )

var updateCmd = &cobra.Command{
	Use:   "update [alias]",
	Short: "Update an existing guidance entity (Not implemented in MVP)",
	Long: `This command will eventually allow for updating the content or metadata
of existing guidance entities. In the current MVP, it is not implemented.`,
	Args: cobra.ExactArgs(1), // Requires an alias
	RunE: func(cmd *cobra.Command, args []string) error {
		alias := args[0]
		slog.Info("'update' command called", "alias", alias)
		fmt.Printf("Command 'update' for alias '%s' is not implemented in MVP.\n", alias)
		return fmt.Errorf("command 'update' not implemented in MVP")
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	// updateCmd.Flags().StringVarP(&updateTitle, "title", "t", "", "New title for the guidance entity")
	// updateCmd.Flags().StringSliceVarP(&updateTags, "tags", "g", []string{}, "New comma-separated tags (replaces existing)")
	// updateCmd.Flags().BoolVarP(&updateEditor, "editor", "e", false, "Open in default editor")
}
