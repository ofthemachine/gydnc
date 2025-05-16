package cmd

import (
	"fmt"
	// "log/slog" // For structured logging, if needed

	"gydnc/config"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		// slog.Debug("Starting 'list' command")
		cfg := config.Get()
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}

		if len(cfg.StorageBackends) == 0 {
			fmt.Println("No storage backends configured.")
			return nil
		}

		fmt.Println("Available guidance entities:")
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
			for _, entityID := range entities { // entities is now []string
				fmt.Printf("    - %s\n", entityID)
				foundEntities++
			}
		}

		if foundEntities == 0 {
			fmt.Println("No guidance entities found across all configured backends.")
		}

		return nil
	},
}

// InitializeBackendFromConfig attempts to create and initialize a backend instance from its config.
// `beConfig` should be a pointer to the config struct, e.g., *config.StorageConfig.
func InitializeBackendFromConfig(name string, beConfig *config.StorageConfig) (storage.Backend, error) {
	if beConfig.Type == "localfs" {
		if beConfig.LocalFS == nil || beConfig.LocalFS.Path == "" {
			return nil, fmt.Errorf("localfs config for backend '%s' is missing or path is empty", name)
		}
		store, err := localfs.NewStore(*beConfig.LocalFS) // Use the imported localfs
		if err != nil {
			return nil, fmt.Errorf("failed to create localfs store for backend '%s': %w", name, err)
		}
		if initErr := store.Init(nil); initErr != nil { // Assuming Init(nil) is okay for now
			return nil, fmt.Errorf("failed to initialize localfs store for backend '%s': %w", name, initErr)
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
