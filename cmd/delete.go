package cmd

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"gydnc/core/content"
	"gydnc/model"

	"log/slog"

	"github.com/spf13/cobra"
)

var forceDelete bool

var deleteCmd = &cobra.Command{
	Use:   "delete <alias1> [alias2 ...]",
	Short: "Delete one or more guidance entities by alias (from all backends)",
	Long: `Deletes one or more guidance entities by alias. Searches all configured backends for each alias.
Requires confirmation unless --force is specified.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		aliases := args

		// Check if app context is initialized
		if appContext == nil || appContext.Config == nil {
			return fmt.Errorf("active backend not initialized; run 'gydnc init' or check config")
		}

		cfg := appContext.Config
		var toDelete []model.Entity
		var notFound []string

		// Track which aliases have been found
		foundAliases := make(map[string]bool)

		for backendName, backendConfig := range cfg.StorageBackends {
			backend, err := InitializeBackendFromConfig(backendName, backendConfig)
			if err != nil {
				continue
			}
			entityIDs, err := backend.List("")
			if err != nil {
				continue
			}
			slog.Debug("Entity IDs found in backend", "backendName", backendName, "entityIDs", entityIDs)
			for _, entityID := range entityIDs {
				for _, alias := range aliases {
					slog.Debug("Matching alias", "alias", alias, "entityID", entityID)
					if entityID == alias {
						contentBytes, meta, err := backend.Read(entityID)
						if err == nil && meta != nil {
							parsed, _ := content.ParseG6E(contentBytes)
							entity := model.Entity{
								Alias:          entityID,
								SourceBackend:  backendName,
								Title:          parsed.Title,
								Description:    parsed.Description,
								Tags:           parsed.Tags,
								CustomMetadata: meta,
								Body:           parsed.Body,
							}
							cid, _ := parsed.GetContentID()
							entity.CID = cid
							toDelete = append(toDelete, entity)
							foundAliases[alias] = true
							slog.Debug("Entity marked for deletion", "entity", entity)
						}
					}
				}
			}
		}

		// Track not found aliases
		for _, alias := range aliases {
			if !foundAliases[alias] {
				notFound = append(notFound, alias)
			}
		}

		if len(toDelete) == 0 {
			appContext.Logger.Info("No matching entities found to delete.")
			if len(notFound) > 0 {
				appContext.Logger.Info("Some aliases provided were not found.", "aliases", strings.Join(notFound, ", "))
			}
			return nil
		}

		// Print summary and confirm
		if !forceDelete {
			fmt.Println("The following entities will be deleted:")
			for _, e := range toDelete {
				path := ""
				if p, ok := e.CustomMetadata["path"].(string); ok {
					path = p
				}
				title := e.Title
				if title == "" {
					title = "(no title)"
				}
				fmt.Printf("- %s (backend: %s, path: %s, title: %s)\n", e.Alias, e.SourceBackend, path, title)
			}
			fmt.Print("Proceed with deletion? [y/N]: ")
			reader := bufio.NewReader(os.Stdin)
			resp, _ := reader.ReadString('\n')
			resp = strings.TrimSpace(strings.ToLower(resp))
			if resp != "y" && resp != "yes" {
				appContext.Logger.Info("Deletion aborted by user.")
				return nil
			}
		}

		// Sort toDelete slice by SourceBackend descending (be2 before be1, etc.) for test determinism
		if len(toDelete) > 1 {
			sort.Slice(toDelete, func(i, j int) bool {
				return toDelete[i].SourceBackend > toDelete[j].SourceBackend
			})
		}

		// Perform deletions
		var deleted, failed []string
		for _, e := range toDelete {
			backend, err := InitializeBackendFromConfig(e.SourceBackend, cfg.StorageBackends[e.SourceBackend])
			if err != nil {
				failed = append(failed, fmt.Sprintf("%s (backend: %s): %v", e.Alias, e.SourceBackend, err))
				continue
			}
			if err := backend.Delete(e.Alias); err != nil {
				failed = append(failed, fmt.Sprintf("%s (backend: %s): %v", e.Alias, e.SourceBackend, err))
			} else {
				deleted = append(deleted, fmt.Sprintf("%s (backend: %s)", e.Alias, e.SourceBackend))
			}
		}

		if len(deleted) > 0 {
			deletedItems := make([]string, len(deleted))
			copy(deletedItems, deleted)
			appContext.Logger.Info("Entities deleted.", "items", deletedItems)
		} else {
			appContext.Logger.Info("No entities were deleted in this operation (either none matched or all failed).", "items", []string{})
		}
		if len(failed) > 0 {
			failedItems := make([]string, len(failed))
			copy(failedItems, failed)
			appContext.Logger.Error("Failed to delete some entities.", "items", failedItems)
		}
		if len(notFound) > 0 && len(toDelete) > 0 {
			appContext.Logger.Info("Some aliases provided were not found (and were not processed for deletion).", "aliases", strings.Join(notFound, ", "))
		}

		// Print available guidance entities summary (same as list command)
		appContext.Logger.Debug("Starting summary of available entities post-deletion.")
		backendNames := make([]string, 0, len(cfg.StorageBackends))
		for backendName := range cfg.StorageBackends {
			backendNames = append(backendNames, backendName)
		}
		sort.Strings(backendNames)
		foundEntities := 0
		for _, backendName := range backendNames {
			backendConfigEntry := cfg.StorageBackends[backendName]
			backend, err := InitializeBackendFromConfig(backendName, backendConfigEntry)
			if err != nil {
				appContext.Logger.Warn("Error initializing backend for post-delete summary list.", "backend", backendName, "error", err)
				continue
			}
			entities, err := backend.List("")
			if err != nil {
				appContext.Logger.Warn("Error listing from backend for post-delete summary list.", "backend", backendName, "error", err)
				continue
			}
			if len(entities) == 0 {
				appContext.Logger.Debug("No entities found in backend for post-delete summary.", "backend", backendName)
				continue
			}
			foundEntitiesInBackend := 0
			for _, entityID := range entities {
				contentBytes, _, readErr := backend.Read(entityID)
				if readErr != nil {
					appContext.Logger.Debug("Error reading entity for summary list, skipping.", "id", entityID, "backend", backendName, "error", readErr)
					continue
				}
				parsed, parseErr := content.ParseG6E(contentBytes)
				if parseErr != nil {
					appContext.Logger.Debug("Error parsing entity for summary list, skipping.", "id", entityID, "backend", backendName, "error", parseErr)
					continue
				}
				appContext.Logger.Debug("Found entity post-delete.", "alias", entityID, "backend", backendName, "title", parsed.Title)
				foundEntitiesInBackend++
			}
			foundEntities += foundEntitiesInBackend
		}
		if foundEntities == 0 {
			appContext.Logger.Info("No guidance entities found across all configured backends post-delete.")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "Delete without confirmation")
}
