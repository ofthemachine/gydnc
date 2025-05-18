package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	// "path/filepath" // Not strictly used in this iteration but often useful
	"slices" // Go 1.21+ for slices.Contains & Sort

	"gydnc/config"
	"gydnc/core/content"
	"gydnc/model"
	"gydnc/storage"
	"gydnc/storage/localfs" // Specific for localfs.NewStore

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
	Short: "Update an existing guidance entity file (.g6e)",
	Long: `Updates an existing .g6e guidance file's frontmatter or body.

<alias> specifies the entity to update, resolved via configured backends.

Flags allow modification of title, description, and tags.
The body of the guidance can be updated by piping new content via stdin.
If no changes are detected after applying flags and new body (if any),
the file will not be modified.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		alias := args[0]
		// slog.Debug("Starting 'update' command", "alias", alias, "title", updateTitle, "description", updateDescription, "addTags", addTags, "removeTags", removeTags)

		cfg := config.Get()
		var backend storage.Backend
		var backendName string // To store the name of the resolved backend
		var originalContentBytes []byte
		var entityMetadata map[string]interface{}
		var err error
		var actualPath string

		// Try active backend first
		activeBkend, activeBkendName := GetActiveBackend() // Ensure this line correctly unpacks two values
		if activeBkend != nil {
			originalContentBytes, entityMetadata, err = activeBkend.Read(alias)
			if err == nil {
				val, ok := entityMetadata["path"]
				if ok && val != nil {
					actualPath, ok = val.(string)
					if ok {
						backend = activeBkend
						backendName = activeBkendName // Store name from active backend
					}
				}
			}
		}

		// If not found via active backend, or active backend is nil, or path was not in metadata
		if backend == nil {
			// discoverEntityAcrossBackends now returns backend name as 3rd string value
			discoveredBackend, discoveredPath, discoveredBackendNameVal, discoverErr := discoverEntityAcrossBackends(cfg, alias)
			if discoverErr != nil {
				return fmt.Errorf("failed to find entity '%s' in any backend: %w", alias, discoverErr)
			}
			backend = discoveredBackend
			actualPath = discoveredPath
			backendName = discoveredBackendNameVal // Keep this assignment

			// Read the content again using the discovered backend and path
			// discoverEntityAcrossBackends only checks for existence and gets metadata like path, not full content.
			// entityMetadata is not used in this branch, so assign to _ to avoid ineffectual assignment.
			originalContentBytes, _, err = backend.Read(actualPath)
			if err != nil {
				return fmt.Errorf("failed to read entity from discovered path '%s' from backend '%s': %w", actualPath, backendName, err)
			}
		}

		parsedContent, err := content.ParseG6E(originalContentBytes)
		if err != nil {
			return fmt.Errorf("failed to parse existing content for '%s' ('%s'): %w", alias, actualPath, err)
		}

		entity := model.Entity{
			Alias:          alias,
			SourceBackend:  backendName,
			Title:          parsedContent.Title,
			Description:    parsedContent.Description,
			Tags:           parsedContent.Tags,
			CustomMetadata: entityMetadata,
			Body:           parsedContent.Body,
		}
		cid, _ := parsedContent.GetContentID()
		entity.CID = cid

		contentModified := false
		originalForTagComparison, _ := content.ParseG6E(originalContentBytes)

		if cmd.Flags().Changed("title") {
			if entity.Title != updateTitle {
				parsedContent.Title = updateTitle
				entity.Title = updateTitle
				contentModified = true
			}
		}

		if cmd.Flags().Changed("description") {
			if entity.Description != updateDescription {
				parsedContent.Description = updateDescription
				entity.Description = updateDescription
				contentModified = true
			}
		}

		// Tags processing
		// originalForTagComparison.Tags contains tags as read from the file (original order).
		// parsedContent.Tags also currently holds these original tags.

		prospectiveTags := make([]string, len(parsedContent.Tags))
		copy(prospectiveTags, parsedContent.Tags) // Start with a copy of current tags from parsedContent

		if cmd.Flags().Changed("add-tag") || cmd.Flags().Changed("remove-tag") {
			// If tag modification flags are used, rebuild the list from a set
			tagsSet := make(map[string]struct{})
			for _, tag := range prospectiveTags { // Use the current list before modifications
				tagsSet[tag] = struct{}{}
			}
			for _, tagToRemove := range removeTags {
				delete(tagsSet, tagToRemove)
			}
			for _, tagToAdd := range addTags {
				tagsSet[tagToAdd] = struct{}{}
			}
			prospectiveTags = make([]string, 0, len(tagsSet))
			for tag := range tagsSet {
				prospectiveTags = append(prospectiveTags, tag)
			}
		}

		// Always sort the prospectiveTags list (either newly built or the original list)
		slices.Sort(prospectiveTags)

		// Compare the sorted prospective tags with the original tags (as read from file, unsorted)
		// If they are different (either because content changed or because sorting changed the order),
		// then the content is considered modified.
		if !slices.Equal(prospectiveTags, originalForTagComparison.Tags) {
			contentModified = true
		}
		parsedContent.Tags = prospectiveTags
		entity.Tags = prospectiveTags

		var newBodyBytes []byte
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			scanner := bufio.NewScanner(os.Stdin)
			var bodyBuilder bytes.Buffer
			for scanner.Scan() {
				bodyBuilder.Write(scanner.Bytes())
				bodyBuilder.WriteString("\n")
			}
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("error reading from stdin: %w", err)
			}
			newBodyBytes = bodyBuilder.Bytes()
			if len(newBodyBytes) > 0 && newBodyBytes[len(newBodyBytes)-1] == '\n' {
				newBodyBytes = newBodyBytes[:len(newBodyBytes)-1] // Trim trailing newline from stdin read
			}

			if string(newBodyBytes) != entity.Body {
				parsedContent.Body = string(newBodyBytes)
				entity.Body = string(newBodyBytes)
				contentModified = true
			}
		}

		if !contentModified {
			fmt.Printf("No changes applied to %s.\n", actualPath)
			return nil
		}

		updatedContentBytes, err := parsedContent.ToFileContent()
		if err != nil {
			return fmt.Errorf("failed to serialize updated content for '%s' ('%s'): %w", alias, actualPath, err)
		}

		if bytes.Equal(originalContentBytes, updatedContentBytes) {
			fmt.Printf("No effective changes detected for %s after serialization. File not modified.\n", actualPath)
			return nil
		}

		// The `id` for Write for LocalFSBackend is the path relative to its root.
		// `actualPath` from `entityMetadata["path"]` should be this.
		err = backend.Write(actualPath, updatedContentBytes, nil)
		if err != nil {
			return fmt.Errorf("failed to write updated entity '%s' ('%s') to backend '%s': %w", alias, actualPath, backendName, err)
		}

		// Construct display path using GetBasePath if available
		displayPath := alias + ".g6e" // Fallback to alias.g6e
		if backend != nil {
			// Diagnostic: Try direct type assertion for localfs.Store
			if lsStore, ok := backend.(*localfs.Store); ok {
				// slog.Debug("Backend is localfs.Store, getting base path directly")
				bp := lsStore.GetBasePath() // Call on concrete type
				if bp != "" {
					displayPath = filepath.Join(bp, alias+".g6e")
				}
			}
		}

		fmt.Printf("Updated guidance: %s\n", displayPath)
		// slog.Info("Successfully updated guidance file", "path", displayPath)
		return nil
	},
}

// discoverEntityAcrossBackends iterates all configured localfs backends to find the entity.
// It is given an alias and attempts to read it from each backend.
// Returns the backend instance, the path relative to the backend (which is the alias itself for localfs),
// the backend's name, and an error if not found.
func discoverEntityAcrossBackends(cfg *config.Config, alias string) (storage.Backend, string, string, error) {
	var lastError error
	for name, backendConfig := range cfg.StorageBackends {
		if backendConfig.Type != "localfs" {
			continue
		}
		if backendConfig.LocalFS == nil || backendConfig.LocalFS.Path == "" {
			continue
		}
		tempStore, err := localfs.NewStore(*backendConfig.LocalFS)
		if err != nil {
			lastError = fmt.Errorf("failed to init temp store for backend %s: %w", name, err)
			continue
		}
		if initErr := tempStore.Init(nil); initErr != nil {
			lastError = fmt.Errorf("failed to initialize temp store for backend %s: %w", name, initErr)
			continue
		}

		_, meta, readErr := tempStore.Read(alias)
		if readErr == nil && meta != nil {
			if pathVal, ok := meta["path"]; ok {
				if pathStr, isStr := pathVal.(string); isStr && pathStr != "" {
					return tempStore, pathStr, name, nil
				}
			}
		}
		if readErr != nil {
			lastError = readErr
		}
	}
	if lastError != nil {
		return nil, "", "", fmt.Errorf("entity '%s' not found after checking all localfs backends, last error: %w", alias, lastError)
	}
	return nil, "", "", fmt.Errorf("entity '%s' not found in any configured localfs backend", alias)
}

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().StringVarP(&updateTitle, "title", "t", "", "New title for the guidance entity")
	updateCmd.Flags().StringVarP(&updateDescription, "description", "d", "", "New description for the guidance entity (provide empty string to clear)")
	updateCmd.Flags().StringSliceVar(&addTags, "add-tag", []string{}, "Tag to add (can be specified multiple times)")
	updateCmd.Flags().StringSliceVar(&removeTags, "remove-tag", []string{}, "Tag to remove (can be specified multiple times)")
	// Note: Unlike create, update does not take a --backend flag. It finds the entity in existing backends.
}
