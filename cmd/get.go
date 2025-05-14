package cmd

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get [alias]",
	Short: "Retrieve and display a specific guidance entity",
	Long: `Retrieves a guidance entity by its alias from the configured backend
and prints its raw content to standard output.`,
	Args: cobra.ExactArgs(1), // Requires exactly one argument: the alias
	RunE: func(cmd *cobra.Command, args []string) error {
		alias := args[0]
		slog.Debug("Starting 'get' command execution", "alias", alias)

		backend := GetActiveBackend()
		if backend == nil {
			return fmt.Errorf("active backend not initialized; run 'gydnc init' or check config")
		}

		slog.Debug("Getting guidance content", "backend", backend.GetName(), "alias", alias)

		contentBytes, metadata, err := backend.Read(alias)
		if err != nil {
			slog.Error("Failed to read guidance from backend", "backend", backend.GetName(), "alias", alias, "error", err)
			return fmt.Errorf("reading guidance '%s' from backend '%s': %w", alias, backend.GetName(), err)
		}

		// For MVP, just print the raw content. Metadata is available if needed for future enhancements.
		// Slog the metadata if verbosity is high enough (assuming slog levels are configured).
		slog.Debug("Successfully read guidance", "alias", alias, "metadata", metadata, "content_length", len(contentBytes))

		fmt.Print(string(contentBytes))

		slog.Debug("'get' command finished successfully", "alias", alias)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	// No flags for getCmd in MVP, but could add --output-format (raw, frontmatter, body, json) later.
}
