package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog" // Added for global logger in panic/early exit
	"os"

	// "sort" // No longer needed directly here if service sorts
	// "path/filepath" // No longer needed directly here
	// "gydnc/core/content" // No longer needed directly here
	// "gydnc/filter" // No longer needed directly here
	"gydnc/model"   // Added import for model.Entity
	"gydnc/service" // Import the service package

	// "gydnc/storage/localfs" // No longer needed directly here

	"github.com/spf13/cobra"
)

var (
	// listJSON bool // Flag becomes effectively obsolete as JSON is default
	filterTags      string
	extendedOutput  bool
	listBackendName string
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available guidance entities",
	Long: `Lists all available guidance entities across configured storage backends.
Supports tag filtering with the --filter-tags flag using syntax like:
- "scope:code quality:safety" (include tags)
- "NOT deprecated" or "-deprecated" (exclude tags)
- "scope:* -deprecated" (wildcards and negation)
Output is always in JSON format.`, // Updated Long description
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		defer func() {
			if r := recover(); r != nil {
				// Use global slog here as appContext might not be initialized or could be part of the panic
				slog.Error("A critical error occurred in the list command", "details", fmt.Sprintf("%v", r))
				os.Exit(1)
			}
		}()

		if appContext == nil {
			// Use global slog here as appContext is nil
			slog.Error("Application context not initialized.")
			os.Exit(1)
		}

		entityService := service.NewEntityService(appContext)
		var allEntities []model.Entity
		var backendErrors map[string]error // Only relevant for merged list
		var listErr error                  // For single backend list errors

		if listBackendName != "" {
			appContext.Logger.Debug("Listing entities for specific backend", "backend", listBackendName, "filter", filterTags)
			allEntities, listErr = entityService.ListEntitiesFromBackend(listBackendName, "", filterTags)
			if listErr != nil {
				// Log the error using the structured logger if available
				if appContext.Logger != nil {
					appContext.Logger.Error("Failed to list entities from backend", "backend", listBackendName, "error", listErr)
				} else {
					fmt.Fprintf(os.Stderr, "Error listing entities from backend '%s': %v\n", listBackendName, listErr)
				}
				os.Exit(1)
			}
			// backendErrors is not populated in this path, as we deal with a single backend.
		} else {
			appContext.Logger.Debug("Listing merged entities from all backends", "filter", filterTags)
			allEntities, backendErrors = entityService.ListEntitiesMerged("", filterTags)
		}

		// Log any backend errors encountered by the service (only for merged list).
		// For single backend list, errors are fatal and handled above.
		if listBackendName == "" && len(backendErrors) > 0 {
			for backendName, err := range backendErrors {
				// Prefer structured logging if available and configured in appContext.
				if appContext.Logger != nil {
					appContext.Logger.Warn("Error accessing backend during list operation", "backend", backendName, "error", err)
				} else {
					fmt.Fprintf(os.Stderr, "Warning: Error accessing backend '%s': %v\n", backendName, err)
				}
			}
		}

		// Output is always JSON
		if len(allEntities) == 0 {
			fmt.Println("[]") // Output empty JSON array
		} else {
			var outputEntities interface{}
			if extendedOutput {
				outputEntities = allEntities
			} else {
				type CompactEntity struct {
					Alias string `json:"alias"`
					// SourceBackend string `json:"source_backend"` // Removed as per user request
					Title       string   `json:"title"`
					Description string   `json:"description"`
					Tags        []string `json:"tags"`
				}
				compactEntities := make([]CompactEntity, len(allEntities))
				for i, entity := range allEntities {
					compactEntities[i] = CompactEntity{
						Alias: entity.Alias,
						// SourceBackend: entity.SourceBackend, // Removed
						Title:       entity.Title,
						Description: entity.Description,
						Tags:        entity.Tags,
					}
				}
				outputEntities = compactEntities
			}

			jsonBytes, err := json.MarshalIndent(outputEntities, "", "  ")
			if err != nil {
				// Prefer structured logging for errors if available.
				if appContext.Logger != nil {
					appContext.Logger.Error("Failed to marshal entities to JSON", "error", err)
				} else {
					fmt.Fprintf(os.Stderr, "Error marshaling entities to JSON: %v\n", err)
				}
				os.Exit(1)
			}
			fmt.Println(string(jsonBytes))
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	// listCmd.Flags().BoolVar(&listJSON, "json", false, "Output in JSON format") // Flag removed, JSON is default
	listCmd.Flags().StringVar(&filterTags, "filter-tags", "", "Filter by tags (e.g., \"scope:code -deprecated\")")
	listCmd.Flags().BoolVar(&extendedOutput, "extended", false, "Include extended metadata in JSON output (includes source_backend)") // Clarified extended output
	listCmd.Flags().StringVar(&listBackendName, "backend", "", "List entities only from a specific backend")
}
