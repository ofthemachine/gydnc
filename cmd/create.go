package cmd

import (
	"bufio"
	"errors" // Added for errors.Is
	"fmt"
	"log/slog" // To be used for debug logging
	"os"

	"strings"

	// For GuidanceContent and ToFileContent
	"gydnc/model"
	"gydnc/storage" // Added for storage.ErrAmbiguousBackend

	"github.com/spf13/cobra"
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
	Short: "Create a new guidance entity",
	Long: `Creates a new guidance entity using the EntityService.

The alias_or_path is used as the entity's alias.
Metadata (title, description, tags) is provided via flags.
Body content can be provided via stdin, --body, or --body-from-file.

The command will fail if the entity already exists in the target backend.
All write operations are handled by the configured storage backend via the EntityService.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		alias := args[0] // Changed from aliasOrPath to just alias, as path resolution is now backend's concern
		slog.Debug("Starting 'create' command with EntityService",
			"alias", alias,
			"title", createTitle,
			"description", createDescription,
			"tags", createTags,
			"backend", createBackend,
			"bodyFromFile", createBodyFromFile,
			"bodyFlagUsed", cmd.Flags().Changed("body"))

		// Check if app context and entity service are initialized
		if appContext == nil || appContext.Config == nil || appContext.EntityService == nil {
			slog.Error("Application context, configuration, or entity service not initialized.")
			return fmt.Errorf("application context, configuration, or entity service not initialized")
		}

		// Determine body content
		var actualBodyContent string
		var bodySourceUsed bool

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
				// Ensure trailing newline if content is not empty
				if !strings.HasSuffix(actualBodyContent, "\n") {
					actualBodyContent += "\n"
				}
			} else {
				actualBodyContent = "" // Explicitly empty for empty stdin
			}
			bodySourceUsed = true
		} else if bodyFlagUsed {
			actualBodyContent = createBody
			// Ensure trailing newline if content is not empty and doesn't have one
			if actualBodyContent != "" && !strings.HasSuffix(actualBodyContent, "\n") {
				actualBodyContent += "\n"
			}
			bodySourceUsed = true
		}

		// Use default title if not provided - user wants blank if not specified
		titleToUse := createTitle

		// Use default body if none provided
		if !bodySourceUsed || actualBodyContent == "" {
			if titleToUse == "" {
				actualBodyContent = "#\n\nGuidance content for '' goes here.\n" // Corrected: '' for empty title placeholder
			} else {
				actualBodyContent = fmt.Sprintf("# %s\n\nGuidance content for '%s' goes here.\n", titleToUse, titleToUse)
			}
		}

		// Create the model.Entity to be saved
		entityToSave := model.Entity{
			Alias:       alias,
			Title:       titleToUse,
			Description: createDescription,
			Tags:        createTags,
			Body:        actualBodyContent,
			// CID and PCID will be handled by the backend/storage layer or if they become part of standard creation flow
			// CustomMetadata can be added here if there's a mechanism to pass it via flags, for now it's empty.
		}

		slog.Debug("Attempting to save entity via EntityService", "alias", entityToSave.Alias, "backend", createBackend)

		// Save the entity using EntityService
		savedBackendName, err := appContext.EntityService.SaveEntity(entityToSave, createBackend)
		if err != nil {
			slog.Error("Failed to save entity using EntityService", "alias", alias, "error", err)
			if errors.Is(err, storage.ErrAmbiguousBackend) {
				return fmt.Errorf("failed to create guidance '%s': %w. Please specify a backend using --backend or set default_backend in config", alias, err)
			}
			return fmt.Errorf("failed to create guidance '%s': %w", alias, err)
		}

		slog.Info("Successfully created guidance.", "alias", alias, "backend", savedBackendName)

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
