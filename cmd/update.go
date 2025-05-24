package cmd

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"slices" // For slices.Sort and slices.Equal
	"strings"

	// For AppContext

	"github.com/spf13/cobra"
	// "gopkg.in/yaml.v3" // May not be needed directly if content package handles it
)

var (
	updateTitle       string
	updateDescription string
	addTags           []string
	removeTags        []string
	// No explicit backend flag for update; it should operate on the entity's current backend.
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update <alias>",
	Short: "Update an existing guidance entity",
	Long: `Updates metadata or content of an existing guidance entity using the EntityService.

The entity is identified by its alias. The update will be applied to the entity
found in its source backend.

Metadata fields (title, description, tags) can be updated via flags.
If content is piped via stdin, it will replace the existing body of the guidance.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		alias := args[0]

		if appContext == nil || appContext.Config == nil || appContext.EntityService == nil {
			slog.Error("Application context, configuration, or entity service not initialized.")
			return fmt.Errorf("application context, configuration, or entity service not initialized")
		}

		slog.Debug("Starting 'update' command with EntityService", "alias", alias)

		// 1. Get the existing entity using EntityService
		entity, err := appContext.EntityService.GetEntity(alias, "") // Search in all backends
		if err != nil {
			slog.Error("Failed to get entity for update", "alias", alias, "error", err)
			return fmt.Errorf("failed to retrieve entity '%s' for update: %w", alias, err)
		}

		slog.Debug("Found entity for update", "alias", entity.Alias, "backend", entity.SourceBackend, "currentTitle", entity.Title)

		contentModified := false

		// Store original values for comparison to see if anything actually changed.
		originalTitle := entity.Title
		originalDescription := entity.Description
		originalTags := make([]string, len(entity.Tags))
		copy(originalTags, entity.Tags)
		originalBody := entity.Body

		// 2. Apply updates to the fetched model.Entity
		if cmd.Flags().Changed("title") {
			if originalTitle != updateTitle {
				slog.Debug("Updating title", "from", originalTitle, "to", updateTitle)
				entity.Title = updateTitle
				contentModified = true
			} else {
				entity.Title = originalTitle // Ensure it's set back if flag was present but value is same
			}
		}

		if cmd.Flags().Changed("description") {
			if originalDescription != updateDescription {
				slog.Debug("Updating description", "from", originalDescription, "to", updateDescription)
				entity.Description = updateDescription
				contentModified = true
			} else {
				entity.Description = originalDescription // Ensure it's set back
			}
		}

		// Handle tag modifications
		if cmd.Flags().Changed("add-tag") || cmd.Flags().Changed("remove-tag") {
			tagsSet := make(map[string]struct{})
			for _, tag := range entity.Tags { // Start with current tags
				tagsSet[tag] = struct{}{}
			}
			for _, tagToRemove := range removeTags {
				delete(tagsSet, tagToRemove)
			}
			for _, tagToAdd := range addTags {
				tagsSet[tagToAdd] = struct{}{}
			}
			updatedTags := make([]string, 0, len(tagsSet))
			for tag := range tagsSet {
				updatedTags = append(updatedTags, tag)
			}
			slices.Sort(updatedTags) // Keep tags sorted for consistency
			entity.Tags = updatedTags
			// contentModified will be checked later by comparing originalTags and entity.Tags
		}

		// Handle body update from stdin
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 { // Check if stdin is piped
			slog.Debug("Stdin is piped, reading new body content.")
			scanner := bufio.NewScanner(os.Stdin)
			var bodyBuilder strings.Builder // Use strings.Builder for efficiency
			for scanner.Scan() {
				bodyBuilder.WriteString(scanner.Text())
				bodyBuilder.WriteString("\n")
			}
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("error reading new body from stdin: %w", err)
			}
			newBody := bodyBuilder.String()
			// Remove trailing newline if body is not empty, G6E anager will add one if needed.
			// if len(newBody) > 0 && newBody[len(newBody)-1] == '\n' {
			// 	newBody = newBody[:len(newBody)-1]
			// }
			if originalBody != newBody {
				slog.Debug("Updating body from stdin.")
				entity.Body = newBody
				contentModified = true
			} else {
				entity.Body = originalBody // Ensure it's set back if stdin was same as original
			}
		}

		// Check if tags were actually modified after add/remove operations and sorting
		slices.Sort(originalTags)
		if !slices.Equal(entity.Tags, originalTags) {
			contentModified = true
			slog.Debug("Tags modified", "from", originalTags, "to", entity.Tags)
		}

		// 3. If no changes, inform user and exit
		if !contentModified {
			// fmt.Printf("No changes detected for entity '%s'. Update not performed.\n", alias)
			appContext.Logger.Info("No changes detected for entity. Update not performed.", "alias", alias)
			return nil
		}

		slog.Debug("Content modified, attempting to save updated entity.", "alias", entity.Alias)

		// 4. Save the updated entity using EntityService
		// The SourceBackend field of the fetched entity tells the service where to save it.
		savedBackendName, err := appContext.EntityService.OverwriteEntity(entity, entity.SourceBackend)
		if err != nil {
			slog.Error("Failed to save updated entity using EntityService", "alias", alias, "error", err)
			return fmt.Errorf("failed to update entity '%s': %w", alias, err)
		}

		// fmt.Printf("Successfully updated entity '%s' in backend '%s'\n", alias, entity.SourceBackend) // Removed, slog.Info below handles this
		slog.Info("Successfully updated entity.", "alias", alias, "backend", savedBackendName)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	// "update" command takes flags to modify metadata elements
	updateCmd.Flags().StringVar(&updateTitle, "title", "", "New title for the guidance file")
	updateCmd.Flags().StringVar(&updateDescription, "description", "", "New description for the guidance file")
	updateCmd.Flags().StringSliceVar(&addTags, "add-tag", nil, "Tags to add to the guidance file (comma-separated)")
	updateCmd.Flags().StringSliceVar(&removeTags, "remove-tag", nil, "Tags to remove from the guidance file (comma-separated)")
}
