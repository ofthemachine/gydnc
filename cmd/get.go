package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"gydnc/core/content"
	"gydnc/model"
	"gydnc/storage"

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

		if appContext == nil || appContext.Config == nil {
			slog.Error("Application context or configuration not initialized.")
			return fmt.Errorf("application context or configuration not initialized")
		}

		// Get all initialized backends from the appContext
		// The GetAllBackends method in AppContext should handle initialization and path resolution internally.
		allConfiguredBackends, backendErrors := appContext.GetAllBackends()

		if len(allConfiguredBackends) == 0 && len(backendErrors) > 0 {
			slog.Error("No backends could be initialized. Please check backend configurations.")
			for name, err := range backendErrors {
				fmt.Fprintf(os.Stderr, "Error initializing backend '%s': %v\n", name, err)
			}
			return fmt.Errorf("no backends could be initialized")
		}
		if len(allConfiguredBackends) == 0 {
			slog.Error("No backends available or configured.")
			return fmt.Errorf("no backends available or configured")
		}

		var results []SimplifiedStructuredOutput
		if len(idsToGet) > 1 {
			results = make([]SimplifiedStructuredOutput, 0, len(idsToGet))
		}

		for _, id := range idsToGet {
			var foundEntity *model.Entity
			var lastReadError error

			// Iterate through the map of initialized ReadOnlyBackend instances
			for backendName, currentBackendStore := range allConfiguredBackends {
				if currentBackendStore == nil { // Should ideally not happen if GetAllBackends filters failed ones
					slog.Warn("Encountered nil backend store, skipping.", "backendName", backendName)
					continue
				}

				slog.Debug("Attempting to get guidance from backend", "id", id, "backend", currentBackendStore.GetName())
				contentBytes, meta, readErr := currentBackendStore.Read(id)

				if readErr == nil {
					parsedData, parseErr := content.ParseG6E(contentBytes)
					if parseErr != nil {
						slog.Error("Failed to parse G6E content after successful read", "id", id, "backend", currentBackendStore.GetName(), "error", parseErr)
						lastReadError = fmt.Errorf("parsing %s from %s: %w", id, currentBackendStore.GetName(), parseErr)
						foundEntity = nil
						break
					}

					cidValue, _ := parsedData.GetContentID()
					foundEntity = &model.Entity{
						Alias:          id,
						SourceBackend:  currentBackendStore.GetName(),
						Title:          parsedData.Title,
						Description:    parsedData.Description,
						Tags:           parsedData.Tags,
						CustomMetadata: meta,
						Body:           parsedData.Body,
						CID:            cidValue,
					}
					lastReadError = nil
					break
				} else {
					if os.IsNotExist(readErr) || readErr == storage.ErrEntityNotFound { // Corrected to use ErrEntityNotFound
						slog.Debug("Entity not found in this backend", "id", id, "backend", currentBackendStore.GetName())
					} else {
						slog.Warn("Error reading from backend (will try others if available)", "id", id, "backend", currentBackendStore.GetName(), "error", readErr)
					}
					lastReadError = readErr
				}
			} // End of backend iteration loop

			// Log any errors encountered during backend initialization for this specific ID's get attempt, if not already covered
			// This is more for context if all backends failed for other reasons before even trying to read.
			for name, err := range backendErrors {
				slog.Warn("Note: Backend initialization failed earlier, which might affect availability.", "id", id, "failedBackendName", name, "initError", err)
			}

			if foundEntity == nil {
				if lastReadError == nil {
					lastReadError = fmt.Errorf("entity '%s' not found in any backend and no specific error recorded", id)
				}
				slog.Error("Failed to get entity from any backend or post-read processing failed", "id", id, "finalError", lastReadError)
				fmt.Fprintf(os.Stderr, "Error getting ID %s: %v\n", id, lastReadError)

				if len(idsToGet) > 1 {
					results = append(results, SimplifiedStructuredOutput{Title: "ERROR_FETCHING_CONTENT_FOR_" + id, Body: fmt.Sprintf("Error: %v", lastReadError)})
				}
				continue
			}

			structuredData := SimplifiedStructuredOutput{
				Title:       foundEntity.Title,
				Description: foundEntity.Description,
				Tags:        foundEntity.Tags,
				Body:        foundEntity.Body,
			}
			if len(idsToGet) > 1 {
				results = append(results, structuredData)
			} else {
				jsonBytes, err := json.MarshalIndent(structuredData, "", "  ")
				if err != nil {
					slog.Error("Failed to marshal structured data to JSON", "id", id, "error", err)
					fmt.Fprintf(os.Stderr, "Error marshalling structured JSON for ID %s: %v\n", id, err)
					continue
				}
				fmt.Fprintln(os.Stdout, string(jsonBytes))
			}
		} // End of id iteration loop

		if len(idsToGet) > 1 && len(results) > 0 {
			finalJsonBytes, err := json.MarshalIndent(results, "", "  ")
			if err != nil {
				slog.Error("Failed to marshal final structured JSON array", "error", err)
				return fmt.Errorf("marshalling final structured JSON array: %w", err)
			}
			fmt.Fprintln(os.Stdout, string(finalJsonBytes))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
