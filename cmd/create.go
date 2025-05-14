package cmd

import (
	"fmt"
	// "log/slog" // To be used later
	"os"
	"path/filepath"
	"strings"

	"gydnc/config"
	"gydnc/core/content" // For GuidanceContent and ToFileContent

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	createTitle       string
	createDescription string
	createTags        []string
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
All necessary parent directories will be created.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		aliasOrPath := args[0]
		// slog.Debug("Starting 'create' command", "aliasOrPath", aliasOrPath, "title", createTitle, "description", createDescription, "tags", createTags)

		cfg := config.Get() // Assumes config is loaded by rootPersistentPreRun
		activeBackend, err := config.GetActiveStorageBackend(cfg)
		if err != nil {
			return fmt.Errorf("could not get active storage backend: %w", err)
		}
		if activeBackend.LocalFS == nil {
			return fmt.Errorf("active backend '%s' is not a localfs backend or not configured", cfg.DefaultBackend)
		}
		storeBasePath := activeBackend.LocalFS.Path

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

		// Generate initial content
		newID := uuid.New().String()
		title := createTitle
		if title == "" {
			// Derive title from filename, removing .g6e and replacing hyphens/underscores
			base := filepath.Base(targetFileName)
			ext := filepath.Ext(base)
			title = base[0 : len(base)-len(ext)]
			// TODO: Make title prettier (e.g., replace hyphens, underscores with spaces, capitalize)
		}
		description := createDescription // Default is empty if not provided
		tags := createTags               // Default is empty if not provided

		// Create an instance of StandardFrontmatter to be marshalled.
		frontmatterData := content.StandardFrontmatter{
			ID:          newID,
			Title:       title,
			Description: description,
			Tags:        tags,
		}

		// For the file content, we need the body.
		// Use the title from the frontmatterData for consistency.
		fileBodyContent := fmt.Sprintf("# %s\n\nGuidance content for '%s' goes here.\n", frontmatterData.Title, frontmatterData.Title)

		frontmatterBytes, err := yaml.Marshal(&frontmatterData)
		if err != nil {
			return fmt.Errorf("failed to serialize frontmatter: %w", err)
		}

		fileContent := append([]byte("---\n"), frontmatterBytes...)
		fileContent = append(fileContent, []byte("---\n")...)
		fileContent = append(fileContent, []byte(fileBodyContent)...)

		// Write file
		if err := os.WriteFile(targetFilePath, fileContent, 0644); err != nil {
			return fmt.Errorf("failed to write guidance file '%s': %w", targetFilePath, err)
		}

		fmt.Printf("Created guidance file: %s\n", targetFilePath)
		// slog.Info("Successfully created guidance file", "path", targetFilePath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringVarP(&createTitle, "title", "t", "", "Title for the new guidance entity")
	createCmd.Flags().StringVarP(&createDescription, "description", "d", "", "Description for the new guidance entity")
	createCmd.Flags().StringSliceVarP(&createTags, "tags", "g", []string{}, "Comma-separated tags (e.g., tag1,category:value2)")
	// Example of how to use a StringArray flag if preferred over StringSlice for comma separation handling by Cobra
	// createCmd.Flags().StringArrayVarP(&createTags, "tags", "g", []string{}, "Tags for the new guidance (can be specified multiple times)")
}
