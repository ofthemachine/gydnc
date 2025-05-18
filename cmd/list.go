package cmd

import (
	"fmt"
	"os"

	// "log/slog" // For structured logging, if needed
	"path/filepath" // Added for path resolution

	"gydnc/config"
	"gydnc/core/content"
	"gydnc/model"
	"gydnc/storage"         // Assuming storage.Backend is defined here
	"gydnc/storage/localfs" // Added for localfs.NewStore

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available guidance entities",
	Long: `Lists all available guidance entities across configured storage backends.
Future enhancements may include filtering by backend or prefix.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintln(os.Stderr, "active backend not initialized; run 'gydnc init' or check config")
				os.Exit(1)
			}
		}()
		if config.GetLoadedConfigActualPath() == "" {
			fmt.Fprintln(os.Stderr, "active backend not initialized; run 'gydnc init' or check config")
			os.Exit(1)
		}
		cfg := config.Get()
		if cfg == nil || len(cfg.StorageBackends) == 0 {
			fmt.Fprintln(os.Stderr, "active backend not initialized; run 'gydnc init' or check config")
			os.Exit(1)
		}
		_, err := config.GetActiveStorageBackend(cfg)
		if err != nil {
			fmt.Fprintln(os.Stderr, "active backend not initialized; run 'gydnc init' or check config")
			os.Exit(1)
		}
		// slog.Debug("Starting 'list' command")

		fmt.Println("Available guidance entities:")
		var allEntities []model.Entity
		foundEntities := 0

		for backendName, backendConfigEntry := range cfg.StorageBackends { // Renamed backendConfig to backendConfigEntry for clarity
			// slog.Debug("Listing entities for backend", "name", backendName, "type", backendConfigEntry.Type)

			var currentBackend storage.Backend
			var err error

			if backendConfigEntry.Type == "localfs" {
				tempBackend, errInit := InitializeBackendFromConfig(backendName, backendConfigEntry) // Pass pointer
				if errInit != nil {
					fmt.Printf("  Error initializing backend %s: %v\n", backendName, errInit)
					continue
				}
				currentBackend = tempBackend
			} else {
				// slog.Info("Skipping non-localfs backend or unhandled type for listing", "name", backendName, "type", backendConfigEntry.Type)
				// For now, we only attempt to list from localfs backends.
				// This should be expanded to support any backend type that implements List.
				fmt.Printf("  Skipping backend %s (type: %s) - only localfs supported for listing in this version.\n", backendName, backendConfigEntry.Type)
				continue
			}

			if currentBackend == nil { // Should be caught by errInit != nil, but as a safeguard
				fmt.Printf("  Could not get a backend instance for %s (was nil after init attempt)\n", backendName)
				continue
			}

			entities, err := currentBackend.List("")
			if err != nil {
				fmt.Printf("  Error listing entities from backend %s: %v\n", backendName, err)
				continue
			}

			if len(entities) == 0 {
				// slog.Debug("No entities found in backend", "name", backendName)
				// Optionally print something or just skip. For now, let's be verbose.
				fmt.Printf("  No entities found in backend: %s\n", backendName)
				continue
			}

			fmt.Printf("  Backend: %s (%s)\n", backendName, backendConfigEntry.Type)
			for _, entityID := range entities {
				contentBytes, meta, readErr := currentBackend.Read(entityID)
				if readErr != nil {
					fmt.Printf("    - %s (error reading: %v)\n", entityID, readErr)
					continue
				}
				parsed, parseErr := content.ParseG6E(contentBytes)
				if parseErr != nil {
					fmt.Printf("    - %s (error parsing: %v)\n", entityID, parseErr)
					continue
				}
				entity := model.Entity{
					Alias:          entityID,
					SourceBackend:  backendName,
					Title:          parsed.Title,
					Description:    parsed.Description,
					Tags:           parsed.Tags,
					CustomMetadata: meta, // Optionally filter meta fields
					Body:           parsed.Body,
				}
				cid, _ := parsed.GetContentID()
				entity.CID = cid
				allEntities = append(allEntities, entity)
				foundEntities++
			}
		}

		if foundEntities == 0 {
			fmt.Println("No guidance entities found across all configured backends.")
			return
		}
		for _, entity := range allEntities {
			fmt.Printf("- %s (backend: %s) | title: %s | tags: %v\n", entity.Alias, entity.SourceBackend, entity.Title, entity.Tags)
		}
	},
}

// InitializeBackendFromConfig attempts to create and initialize a backend instance from its config.
// `beConfig` should be a pointer to the config struct, e.g., *config.StorageConfig.
func InitializeBackendFromConfig(name string, beConfig *config.StorageConfig) (storage.Backend, error) {
	if beConfig.Type == "localfs" {
		if beConfig.LocalFS == nil || beConfig.LocalFS.Path == "" {
			return nil, fmt.Errorf("localfs config for backend '%s' is missing or path is empty", name)
		}

		cfgPath := config.GetLoadedConfigActualPath()
		resolvedPath := beConfig.LocalFS.Path

		if !filepath.IsAbs(resolvedPath) && cfgPath != "" {
			configFileDir := filepath.Dir(cfgPath)
			resolvedPath = filepath.Join(configFileDir, resolvedPath)
		}

		absResolvedPath, err := filepath.Abs(resolvedPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path for resolved localfs path '%s' for backend '%s': %w", resolvedPath, name, err)
		}

		storeSpecificConfig := config.LocalFSConfig{Path: absResolvedPath}
		store, err := localfs.NewStore(storeSpecificConfig) // Use the imported localfs
		if err != nil {
			return nil, fmt.Errorf("failed to create localfs store for backend '%s' (resolved path: %s): %w", name, absResolvedPath, err)
		}
		if initErr := store.Init(nil); initErr != nil { // Assuming Init(nil) is okay for now
			userFacingPath := beConfig.LocalFS.Path
			if absResolvedPath != userFacingPath {
				userFacingPath = fmt.Sprintf("%s (resolved to %s)", beConfig.LocalFS.Path, absResolvedPath)
			}
			return nil, fmt.Errorf("failed to initialize localfs store for backend '%s' at %s: %w", name, userFacingPath, initErr)
		}
		return store, nil
	}
	return nil, fmt.Errorf("backend type '%s' not supported by InitializeBackendFromConfig yet", beConfig.Type)
}

func init() {
	rootCmd.AddCommand(listCmd)
	// Add flags here if needed in the future, e.g.:
	// listCmd.Flags().StringP("backend", "b", "", "Filter by specific backend name")
	// listCmd.Flags().StringP("prefix", "p", "", "Filter by entity ID prefix")
}
