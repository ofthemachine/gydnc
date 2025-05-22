package cmd

import (
	"bufio"
	"fmt"

	// "log/slog" // To be used later
	"os"
	"path/filepath"
	"strings"

	"gydnc/core/content" // For GuidanceContent and ToFileContent
	"gydnc/model"
	"gydnc/service"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	createTitle        string
	createDescription  string
	createTags         []string
	createBackend      string // Added for backend selection
	createBodyFromFile string
	createBody         string
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create <alias_or_path>",
	Short: "Create a new guidance entity file (.g6e)",
	Long: `Creates a new .g6e guidance file.

If <alias_or_path> is a simple name (e.g., "my-guidance"), the file
is created in the default guidance store directory with the .g6e extension
(e.g., <store_path>/my-guidance.g6e).

If <alias_or_path> includes slashes (e.g., "category/my-guidance"),
it is treated as a path relative to the default guidance store directory.
The .g6e extension will be added if not present.

The command will fail if the target file already exists.
All necessary parent directories will be created.

Body content can be provided via one of three methods (mutually exclusive):
- Piping to stdin (e.g., echo "Body content" | gydnc create ...)
- Using the --body flag (e.g., gydnc create ... --body "Body content")
- Using the --body-from-file flag (e.g., gydnc create ... --body-from-file path/to/body.txt)

If no body is provided, a default placeholder body will be generated.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		aliasOrPath := args[0]
		// slog.Debug("Starting 'create' command", "aliasOrPath", aliasOrPath, "title", createTitle, "description", createDescription, "tags", createTags, "backend", createBackend)

		// Check if app context is initialized
		if appContext == nil || appContext.Config == nil {
			return fmt.Errorf("configuration not loaded; run 'gydnc init' or check config")
		}

		// Get config from app context and initialize config service
		cfg := appContext.Config
		configService := service.NewConfigService(appContext)

		var targetBackendName string
		var chosenBackendConfig *model.StorageConfig

		if createBackend != "" {
			targetBackendName = createBackend
			backendConfig, ok := cfg.StorageBackends[targetBackendName]
			if !ok {
				return fmt.Errorf("specified backend '%s' not found in configuration", targetBackendName)
			}
			if backendConfig == nil {
				return fmt.Errorf("configuration for specified backend '%s' is nil (should not happen if key exists)", targetBackendName)
			}
			chosenBackendConfig = backendConfig
		} else if cfg.DefaultBackend != "" {
			targetBackendName = cfg.DefaultBackend
			backendConfig, ok := cfg.StorageBackends[targetBackendName]
			if !ok {
				return fmt.Errorf("default backend '%s' (from config) not found in storage_backends configuration", targetBackendName)
			}
			if backendConfig == nil {
				return fmt.Errorf("configuration for default backend '%s' is nil (should not happen if key exists)", targetBackendName)
			}
			chosenBackendConfig = backendConfig
		} else {
			if len(cfg.StorageBackends) == 0 {
				return fmt.Errorf("no storage backends configured")
			}
			if len(cfg.StorageBackends) == 1 {
				// If only one backend, use it by default
				for name, backendCfg := range cfg.StorageBackends {
					targetBackendName = name
					chosenBackendConfig = backendCfg
					// slog.Debug("Only one backend configured, using it by default", "backendName", targetBackendName)
					break
				}
			} else {
				// Multiple backends, no default, no explicit choice
				return fmt.Errorf("multiple backends configured and no default is set. Please specify a backend using --backend or set default_backend in config")
			}
		}

		if chosenBackendConfig == nil { // Should be caught by earlier checks, but as a safeguard
			return fmt.Errorf("failed to determine a target storage backend")
		}

		var storeBasePath string
		switch chosenBackendConfig.Type {
		case "localfs":
			if chosenBackendConfig.LocalFS == nil {
				return fmt.Errorf("backend '%s' is type 'localfs' but has no localfs settings configured", targetBackendName)
			}
			if chosenBackendConfig.LocalFS.Path == "" {
				return fmt.Errorf("localfs backend '%s' is missing the 'path' setting", targetBackendName)
			}

			// Resolve storeBasePath relative to the config file's directory
			rawPathFromConfig := chosenBackendConfig.LocalFS.Path
			if !filepath.IsAbs(rawPathFromConfig) {
				// Get the loaded config path using service
				loadedConfigPath, err := configService.GetEffectiveConfigPath(cfgFile)
				if err == nil && loadedConfigPath != "" {
					configDir := filepath.Dir(loadedConfigPath)
					storeBasePath = filepath.Join(configDir, rawPathFromConfig)
				} else {
					// If no config file path, treat relative path from config as relative to CWD
					storeBasePath = rawPathFromConfig
				}
			} else {
				storeBasePath = rawPathFromConfig
			}
			// Ensure storeBasePath is an absolute path for subsequent operations
			var err error
			storeBasePath, err = filepath.Abs(storeBasePath)
			if err != nil {
				return fmt.Errorf("failed to get absolute path for storeBasePath ('%s'): %w", storeBasePath, err)
			}
		default:
			return fmt.Errorf("backend '%s' has an unsupported type '%s' for the create command", targetBackendName, chosenBackendConfig.Type)
		}

		// Determine target file path
		targetFileName := aliasOrPath
		if filepath.Ext(targetFileName) == "" {
			targetFileName += ".g6e"
		}
		targetFilePath := filepath.Join(storeBasePath, targetFileName)

		// Ensure the path is within the storeBasePath (basic safety)
		absTargetFilePath, err := filepath.Abs(targetFilePath)
		if err != nil {
			return fmt.Errorf("could not get absolute path for target: %w", err)
		}
		absStoreBasePath, err := filepath.Abs(storeBasePath)
		if err != nil {
			return fmt.Errorf("could not get absolute path for store: %w", err)
		}
		// Check if the target path is safely within the store base path.
		// filepath.Rel will return an error if they don't share a common prefix,
		// or a path starting with ".." if target is outside base.
		relPath, err := filepath.Rel(absStoreBasePath, absTargetFilePath)
		if err != nil {
			// This case implies they don't share a common base or one is not abs, which Abs should prevent.
			// However, direct check for safety is good.
			return fmt.Errorf("target path '%s' is not relatable to store path '%s': %w", targetFilePath, storeBasePath, err)
		}
		// If relPath starts with "..", it means absTargetFilePath is outside absStoreBasePath.
		// Also, if relPath is exactly ".." (though Rel should produce a more specific path like "../target").
		if strings.HasPrefix(relPath, "..") {
			return fmt.Errorf("resolved target path '%s' attempts to navigate outside the configured store path '%s'", absTargetFilePath, absStoreBasePath)
		}

		// Safety Check: Fail if file already exists
		if _, err := os.Stat(targetFilePath); err == nil {
			return fmt.Errorf("guidance file '%s' already exists", targetFilePath)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("error checking for existing file '%s': %w", targetFilePath, err)
		}

		// Create parent directories if they don't exist
		targetDir := filepath.Dir(targetFilePath)
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory '%s': %w", targetDir, err)
		}

		var actualBodyContent string
		var bodySourceUsed bool

		// Check flags and stdin for body content
		bodyFromFileFlagUsed := cmd.Flags().Changed("body-from-file")
		bodyFlagUsed := cmd.Flags().Changed("body")

		stat, _ := os.Stdin.Stat()
		stdinIsPiped := (stat.Mode() & os.ModeCharDevice) == 0

		sourcesProvided := 0
		if bodyFromFileFlagUsed {
			sourcesProvided++
		}
		if bodyFlagUsed {
			sourcesProvided++
		}
		if stdinIsPiped {
			sourcesProvided++
		}

		if sourcesProvided > 1 {
			return fmt.Errorf("multiple body sources provided (--body-from-file, --body, stdin); please use only one")
		}

		if bodyFromFileFlagUsed {
			bodyBytes, err := os.ReadFile(createBodyFromFile)
			if err != nil {
				return fmt.Errorf("failed to read body from file '%s': %w", createBodyFromFile, err)
			}
			actualBodyContent = string(bodyBytes)
			bodySourceUsed = true
		} else if stdinIsPiped {
			scanner := bufio.NewScanner(os.Stdin)
			var lines []string
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("error reading body from stdin: %w", err)
			}

			if len(lines) > 0 {
				actualBodyContent = strings.Join(lines, "\n")
			} else {
				actualBodyContent = "" // Explicitly empty for empty stdin
			}
			bodySourceUsed = true // Considered used even if stdin was empty but piped
		} else if bodyFlagUsed {
			actualBodyContent = createBody
			if actualBodyContent != "" && !strings.HasSuffix(actualBodyContent, "\n") {
				actualBodyContent += "\n"
			}
			bodySourceUsed = true
		}

		// Generate initial content
		// newID := uuid.New().String() // Removed: ID is now content-addressable
		title := createTitle
		if title == "" {
			// Derive title from filename, removing .g6e and replacing hyphens/underscores
			base := filepath.Base(targetFileName)
			ext := filepath.Ext(base)
			title = base[0 : len(base)-len(ext)]
			// TODO: Make title prettier (e.g., replace hyphens, underscores with spaces, capitalize)
		}
		description := createDescription // Default is empty if not provided

		// Use default body if none provided
		if !bodySourceUsed || actualBodyContent == "" {
			actualBodyContent = fmt.Sprintf("# %s\n\nGuidance content for '%s' goes here.\n", title, title)
		}

		// Create content object with metadata
		content := content.GuidanceContent{
			Title:       title,
			Description: description,
			Tags:        createTags,
			Body:        actualBodyContent,
		}

		// Generate file content with frontmatter
		fileContent, err := content.ToFileContent()
		if err != nil {
			return fmt.Errorf("failed to generate guidance file content: %w", err)
		}

		// Write to file
		if err := os.WriteFile(targetFilePath, fileContent, 0644); err != nil {
			return fmt.Errorf("failed to write guidance file: %w", err)
		}

		// For test compatibility, we need to output relative paths
		workingDir, err := os.Getwd()
		if err == nil && strings.HasPrefix(targetFilePath, workingDir) {
			// Output relative path for test compatibility
			relativePath, err := filepath.Rel(workingDir, targetFilePath)
			if err == nil {
				fmt.Printf("Created guidance file: %s\n", relativePath)
			} else {
				fmt.Printf("Created guidance file: %s\n", targetFilePath)
			}
		} else {
			fmt.Printf("Created guidance file: %s\n", targetFilePath)
		}
		// slog.Info("Successfully created guidance file", "path", targetFilePath)

		// Return YAML representation of created entity for display
		metaDisplay := struct {
			Backend     string   `yaml:"backend"`
			Alias       string   `yaml:"alias"`
			Title       string   `yaml:"title"`
			Description string   `yaml:"description"`
			Tags        []string `yaml:"tags,omitempty"`
			Path        string   `yaml:"path"`
		}{
			Backend:     targetBackendName,
			Alias:       aliasOrPath,
			Title:       title,
			Description: description,
			Tags:        createTags,
			Path:        targetFilePath,
		}

		yamlData, err := yaml.Marshal(metaDisplay)
		if err != nil {
			return fmt.Errorf("failed to marshal entity metadata for display: %w", err)
		}
		fmt.Println(string(yamlData))

		return nil
	},
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringVarP(&createTitle, "title", "t", "", "Title for the new guidance entity")
	createCmd.Flags().StringVarP(&createDescription, "description", "d", "", "Description for the new guidance entity")
	createCmd.Flags().StringSliceVarP(&createTags, "tags", "g", []string{}, "Comma-separated tags (e.g., tag1,category:value2)")
	createCmd.Flags().StringVar(&createBackend, "backend", "", "Name of the storage backend to use (overrides default_backend from config)") // Added flag
	createCmd.Flags().StringVar(&createBodyFromFile, "body-from-file", "", "Path to a file containing the body for the new guidance")
	createCmd.Flags().StringVar(&createBody, "body", "", "Direct string content for the body of the new guidance")
	// Example of how to use a StringArray flag if preferred over StringSlice for comma separation handling by Cobra
	// createCmd.Flags().StringArrayVarP(&createTags, "tags", "g", []string{}, "Tags for the new guidance (can be specified multiple times)")
}
