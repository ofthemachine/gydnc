package cmd

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

var hashCmd = &cobra.Command{
	Use:   "hash [alias]",
	Short: "Calculate and display the G3A CID for a guidance entity (Not implemented in MVP)",
	Long: `This command will calculate the G3A Content Identifier (CID) for a given
guidance entity. In the current MVP, it is not implemented.`,
	Args: cobra.ExactArgs(1), // Requires an alias
	RunE: func(cmd *cobra.Command, args []string) error {
		alias := args[0]
		slog.Info("'hash' command called", "alias", alias)
		fmt.Printf("Command 'hash' for alias '%s' is not implemented in MVP.\n", alias)
		return fmt.Errorf("command 'hash' not implemented in MVP")
	},
}

func init() {
	rootCmd.AddCommand(hashCmd)
	// Flags for hash might include selecting canonicalization profile or hash algorithm if made configurable.
}
