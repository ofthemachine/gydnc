package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

var (
	listOutputFormat string
	listFilterQuery  string
)

var listCmd = &cobra.Command{
	Use:   "list [filter_query]",
	Short: "List available guidance aliases",
	Long: `Scans the configured backend for guidance entities and lists their aliases.
Supports basic filtering (e.g., "tags:mytag").`,
	RunE: func(cmd *cobra.Command, args []string) error {
		slog.Debug("Starting 'list' command execution")

		backend := GetActiveBackend()
		if backend == nil {
			return fmt.Errorf("active backend not initialized; run 'gydnc init' or check config")
		}

		// Use the flag first, then fallback to positional argument if provided
		filter := listFilterQuery
		if filter == "" && len(args) > 0 {
			filter = args[0]
		}

		slog.Debug("Listing guidance", "backend", backend.GetName(), "filter", filter)

		aliases, err := backend.List(filter)
		if err != nil {
			slog.Error("Failed to list guidance from backend", "backend", backend.GetName(), "error", err)
			return fmt.Errorf("listing guidance from backend '%s': %w", backend.GetName(), err)
		}

		if listOutputFormat == "json" {
			jsonData, jsonErr := json.MarshalIndent(aliases, "", "  ")
			if jsonErr != nil {
				slog.Error("Failed to marshal alias list to JSON", "error", jsonErr)
				return fmt.Errorf("marshalling alias list to JSON: %w", jsonErr)
			}
			fmt.Println(string(jsonData))
		} else {
			if len(aliases) == 0 {
				fmt.Println("No guidance aliases found matching the criteria.")
			} else {
				fmt.Printf("Found %d guidance alias(es) in backend '%s':\n", len(aliases), backend.GetName())
				for _, alias := range aliases {
					fmt.Println(alias)
				}
			}
		}

		slog.Debug("'list' command finished successfully", "aliases_found", len(aliases))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// listCmd.Flags().StringVarP(&listConfigPath, "config", "c", "config.yaml", "Path to configuration file") // Config is now global
	listCmd.Flags().StringVarP(&listOutputFormat, "output", "o", "text", "Output format (text, json)")
	listCmd.Flags().StringVarP(&listFilterQuery, "filter", "f", "", "Filter query for listing (e.g., \"tags:example AND tags:another\")")

}
