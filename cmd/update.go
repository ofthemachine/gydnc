package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices" // For slices.Sort and slices.Equal

	"gydnc/core/content"
	"gydnc/model"
	"gydnc/service" // For AppContext
	"gydnc/storage"
	"gydnc/storage/localfs"

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
	Long: `Updates metadata or content of an existing guidance entity.

The entity is identified by its alias. If the alias exists in multiple backends,
the command will error unless a specific backend is targetable (future feature).

Metadata fields (title, description, tags) can be updated via flags.
If content is piped via stdin, it will replace the existing body of the guidance.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		alias := args[0]

		if appContext == nil || appContext.Config == nil {
			return fmt.Errorf("application context or configuration not initialized")
		}

		// Discover the entity across backends using the appContext
		backend, actualPath, backendName, err := discoverEntityAcrossBackends(appContext, alias) // Pass appContext
		if err != nil {
			return fmt.Errorf("failed to discover entity '%s': %w", alias, err)
		}

		if backend == nil {
			return fmt.Errorf("entity '%s' not found or backend could not be determined", alias)
		}

		slog.Debug("Found entity for update", "alias", alias, "backendName", backendName, "pathInBackend", actualPath)

		originalContentBytes, entityMetadata, err := backend.Read(actualPath) // actualPath is the alias for localfs
		if err != nil {
			return fmt.Errorf("failed to read entity '%s' from backend '%s': %w", alias, backendName, err)
		}

		parsedContent, err := content.ParseG6E(originalContentBytes)
		if err != nil {
			return fmt.Errorf("failed to parse G6E content for '%s' ('%s'): %w", alias, actualPath, err)
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

		prospectiveTags := make([]string, len(parsedContent.Tags))
		copy(prospectiveTags, parsedContent.Tags)

		if cmd.Flags().Changed("add-tag") || cmd.Flags().Changed("remove-tag") {
			tagsSet := make(map[string]struct{})
			for _, tag := range prospectiveTags {
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

		slices.Sort(prospectiveTags)

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
				newBodyBytes = newBodyBytes[:len(newBodyBytes)-1]
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

		err = backend.Write(actualPath, updatedContentBytes, nil)
		if err != nil {
			return fmt.Errorf("failed to write updated entity '%s' ('%s') to backend '%s': %w", alias, actualPath, backendName, err)
		}

		displayPath := alias + ".g6e"
		if backend != nil {
			if lsStore, ok := backend.(*localfs.Store); ok {
				bp := lsStore.GetBasePath()
				if bp != "" {
					displayPath = filepath.Join(bp, alias+".g6e")
				}
			}
		}

		slog.Debug("Updated guidance file", "path", displayPath)
		return nil
	},
}

// discoverEntityAcrossBackends iterates all configured localfs backends to find the entity.
// It is given an alias and attempts to read it from each backend.
// Returns the backend instance, the path relative to the backend (which is the alias itself for localfs),
// the backend's name, and an error if not found.
func discoverEntityAcrossBackends(appCtx *service.AppContext, alias string) (storage.Backend, string, string, error) {
	var lastError error
	var foundBackends []string
	var foundBackend storage.Backend
	var foundPath string
	var foundBackendName string

	if appCtx == nil || appCtx.Config == nil {
		return nil, "", "", fmt.Errorf("appContext or its Config is nil in discoverEntityAcrossBackends")
	}
	cfg := appCtx.Config

	configDir := ""
	if appCtx.ConfigPath != "" {
		configDir = filepath.Dir(appCtx.ConfigPath)
	} else {
		// If ConfigPath is empty, behavior for relative paths is undefined or defaults to CWD for localfs.NewStore.
		// This case should ideally be prevented by initConfig setting ConfigPath.
		slog.Warn("appContext.ConfigPath is empty in discoverEntityAcrossBackends; relative backend paths may not resolve as expected.")
		// localfs.NewStore will use CWD if configDir is "" and path is relative.
	}

	for name, backendConfig := range cfg.StorageBackends {
		if backendConfig.Type != "localfs" {
			continue
		}
		if backendConfig.LocalFS == nil || backendConfig.LocalFS.Path == "" {
			continue
		}
		// Pass configDir to localfs.NewStore
		tempStore, err := localfs.NewStore(*backendConfig.LocalFS, configDir)
		if err != nil {
			lastError = fmt.Errorf("failed to init temp store for backend %s (path: %s, configDir: %s): %w", name, backendConfig.LocalFS.Path, configDir, err)
			continue
		}
		if initErr := tempStore.Init(map[string]interface{}{"name": name}); initErr != nil {
			lastError = fmt.Errorf("failed to initialize temp store for backend %s: %w", name, initErr)
			continue
		}
		// tempStore.SetName(name) // Init now handles setting the name

		_, stats, readErr := tempStore.Read(alias)
		if readErr == nil && stats != nil {
			foundBackends = append(foundBackends, name)
			if foundBackend == nil {
				foundBackend = tempStore
				foundPath = alias
				foundBackendName = name
			}
		}
	}

	if len(foundBackends) > 1 {
		return nil, "", "", fmt.Errorf("entity '%s' found in multiple backends (%v); please specify which backend to update or ensure the entity exists in only one backend", alias, foundBackends)
	}

	if len(foundBackends) == 1 {
		return foundBackend, foundPath, foundBackendName, nil
	}

	if lastError != nil {
		return nil, "", "", lastError
	}
	return nil, "", "", fmt.Errorf("entity '%s' not found in any backend", alias)
}

func init() {
	rootCmd.AddCommand(updateCmd)
	// "update" command takes flags to modify metadata elements
	updateCmd.Flags().StringVar(&updateTitle, "title", "", "New title for the guidance file")
	updateCmd.Flags().StringVar(&updateDescription, "description", "", "New description for the guidance file")
	updateCmd.Flags().StringSliceVar(&addTags, "add-tag", nil, "Tags to add to the guidance file (comma-separated)")
	updateCmd.Flags().StringSliceVar(&removeTags, "remove-tag", nil, "Tags to remove from the guidance file (comma-separated)")
}
