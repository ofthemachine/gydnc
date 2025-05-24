package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

// NEW SIMPLIFIED STRUCT for "structured" (default) JSON output
type SimplifiedStructuredOutput struct {
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Body        string   `json:"body"`
}

var getCmd = &cobra.Command{
	Use:   "get <id1> [id2...]",
	Short: "Retrieves and displays one or more guidance entities by their ID(s) as JSON.",
	Long: `Retrieves and displays the content of one or more guidance entities
from the configured backend, based on their IDs. Output is always in JSON format
containing title, description, tags, and body.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		idsToGet := args

		if appContext == nil || appContext.Config == nil || appContext.EntityService == nil {
			slog.Error("Application context, configuration, or entity service not initialized.")
			return fmt.Errorf("application context, configuration, or entity service not initialized")
		}

		var results []SimplifiedStructuredOutput
		if len(idsToGet) > 1 {
			results = make([]SimplifiedStructuredOutput, 0, len(idsToGet))
		}

		for _, id := range idsToGet {
			entity, err := appContext.EntityService.GetEntity(id, "")

			if err != nil {
				slog.Error("Failed to get entity using EntityService", "id", id, "error", err)
				if len(idsToGet) > 1 {
					results = append(results, SimplifiedStructuredOutput{Title: "ERROR_FETCHING_CONTENT_FOR_" + id, Body: fmt.Sprintf("Error: %v", err)})
				}
				continue
			}

			structuredData := SimplifiedStructuredOutput{
				Title:       entity.Title,
				Description: entity.Description,
				Tags:        entity.Tags,
				Body:        entity.Body,
			}

			if len(idsToGet) > 1 {
				results = append(results, structuredData)
			} else {
				jsonBytes, marshalErr := json.MarshalIndent(structuredData, "", "  ")
				if marshalErr != nil {
					slog.Error("Failed to marshal structured data to JSON", "id", id, "error", marshalErr)
					continue
				}
				fmt.Fprintln(os.Stdout, string(jsonBytes))
			}
		}

		if len(idsToGet) > 1 && len(results) > 0 {
			finalJsonBytes, marshalErr := json.MarshalIndent(results, "", "  ")
			if marshalErr != nil {
				slog.Error("Failed to marshal final structured JSON array", "error", marshalErr)
				return fmt.Errorf("marshalling final structured JSON array: %w", marshalErr)
			}
			fmt.Fprintln(os.Stdout, string(finalJsonBytes))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
