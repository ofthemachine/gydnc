package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	// "log/slog" // For structured logging, if needed
	"path/filepath" // Added for path resolution

	"gydnc/core/content"
	"gydnc/filter"
	"gydnc/model"
	"gydnc/service"
	"gydnc/storage/localfs" // Added for localfs.NewStore

	"github.com/spf13/cobra"
)

var listJSON bool
var filterTags string
var extendedOutput bool

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available guidance entities",
	Long: `Lists all available guidance entities across configured storage backends.
Supports tag filtering with the --filter-tags flag using syntax like:
- "scope:code quality:safety" (include tags)
- "NOT deprecated" or "-deprecated" (exclude tags)
- "scope:* -deprecated" (wildcards and negation)`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintln(os.Stderr, "active backend not initialized; run 'gydnc init' or check config")
				os.Exit(1)
			}
		}()

		// Check if app context is initialized
		if appContext == nil || appContext.Config == nil {
			fmt.Fprintln(os.Stderr, "active backend not initialized; run 'gydnc init' or check config")
			os.Exit(1)
		}

		// Get config from app context
		cfg := appContext.Config
		if cfg == nil || len(cfg.StorageBackends) == 0 {
			fmt.Fprintln(os.Stderr, "active backend not initialized; run 'gydnc init' or check config")
			os.Exit(1)
		}

		// Create a config service to help with operations
		configService := service.NewConfigService(appContext)

		// Get the active storage backend config
		_, err := configService.GetActiveStorageBackend(cfg)
		if err != nil {
			fmt.Fprintln(os.Stderr, "active backend not initialized; run 'gydnc init' or check config")
			os.Exit(1)
		}
		// slog.Debug("Starting 'list' command")

		if !listJSON {
			fmt.Println("Available guidance entities:")
		}
		var allEntities []model.Entity
		foundEntities := 0

		// Get the config path for resolving relative paths
		configPath, err := configService.GetEffectiveConfigPath(cfgFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting config path: %v\n", err)
			os.Exit(1)
		}

		foundBackends := false // Track if we found any valid backends to query
		for backendName, backendCfg := range cfg.StorageBackends {
			if backendCfg.Type != "localfs" || backendCfg.LocalFS == nil {
				// Skip non-localfs or improperly configured backends
				if !listJSON {
					fmt.Printf("  Backend '%s' skipped (not a configured localfs)\n", backendName)
				}
				continue
			}

			foundBackends = true
			resolvedPath := backendCfg.LocalFS.Path
			// If path is relative, it's relative to the config file's directory.
			cfgDir := filepath.Dir(configPath)
			if !filepath.IsAbs(resolvedPath) {
				resolvedPath = filepath.Join(cfgDir, resolvedPath)
			}

			store, err := localfs.NewStore(model.LocalFSConfig{Path: resolvedPath})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Could not initialize backend '%s': %v\n", backendName, err)
				continue
			}

			// Set the name in the store for proper attribution in entity listings
			store.SetName(backendName)

			// List all entities in this backend
			aliases, err := store.List("")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Could not list entities from backend '%s': %v\n", backendName, err)
				continue
			}

			// Process each alias to build entity objects
			var backendEntities []model.Entity
			for _, alias := range aliases {
				contentBytes, metadata, err := store.Read(alias)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Error reading entity '%s' from backend '%s': %v\n", alias, backendName, err)
					continue
				}

				// Process content to extract frontmatter
				parsed, err := content.ParseG6E(contentBytes)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Error parsing entity '%s' from backend '%s': %v\n", alias, backendName, err)
					continue
				}

				entity := model.Entity{
					Alias:          alias,
					SourceBackend:  backendName,
					Title:          parsed.Title,
					Description:    parsed.Description,
					Tags:           parsed.Tags,
					CustomMetadata: metadata,
					Body:           parsed.Body,
				}

				// Sort tags for deterministic output and comparison
				sort.Strings(entity.Tags)

				cid, _ := parsed.GetContentID()
				entity.CID = cid

				backendEntities = append(backendEntities, entity)
			}

			// Apply filter if provided
			if filterTags != "" {
				filtered, err := filter.ApplyFilter(backendEntities, filterTags)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error applying filter: %v\n", err)
					os.Exit(1)
				}
				backendEntities = filtered
			}

			// Count the entities found after filtering
			count := len(backendEntities)
			foundEntities += count

			// Add to our combined list
			allEntities = append(allEntities, backendEntities...)

			// Display count
			if !listJSON {
				fmt.Printf("  Found %d entities in backend: %s\n", count, backendName)
			}
		}

		if !foundBackends {
			if !listJSON {
				fmt.Println("No configured localfs backends found.")
			}
		} else if foundEntities == 0 && !listJSON {
			fmt.Println("  No entities found.")
			fmt.Println("No guidance entities found across all configured backends.")
		}

		// Display the entities as JSON if requested
		if listJSON && len(allEntities) > 0 {
			// Sort entities by alias for consistent output
			sort.Slice(allEntities, func(i, j int) bool {
				return allEntities[i].Alias < allEntities[j].Alias
			})

			// Create compact or extended output
			var outputEntities interface{}

			if extendedOutput {
				// Use full entity structures
				outputEntities = allEntities
			} else {
				// Create compact representation with only essential fields
				compactEntities := make([]map[string]interface{}, len(allEntities))
				for i, entity := range allEntities {
					compactEntities[i] = map[string]interface{}{
						"alias":       entity.Alias,
						"title":       entity.Title,
						"description": entity.Description,
						"tags":        entity.Tags,
					}
				}
				outputEntities = compactEntities
			}

			// Output as JSON
			jsonBytes, err := json.MarshalIndent(outputEntities, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error marshaling entities to JSON: %v\n", err)
				os.Exit(1)
			}

			fmt.Println(string(jsonBytes))
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output in JSON format")
	listCmd.Flags().StringVar(&filterTags, "filter-tags", "", "Filter by tags (e.g., \"scope:code -deprecated\")")
	listCmd.Flags().BoolVar(&extendedOutput, "extended", false, "Include extended metadata in JSON output")
}
