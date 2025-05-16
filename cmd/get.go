package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"gydnc/core/content"

	"github.com/spf13/cobra"
)

var outputFormatGet string

// --- Structs for "structured" JSON output ---
type StructuredGuidanceOutput struct {
	ID          string          `json:"id"`
	Frontmatter FrontmatterData `json:"frontmatter"`
	Body        string          `json:"body"`
}

type FrontmatterData struct {
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// --- End structs for "structured" JSON output ---

// Helper struct for "json-frontmatter" output (was jsonFrontmatterOutput)
type JsonFrontmatterOnlyOutput struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

var getCmd = &cobra.Command{
	Use:   "get <id1> [id2...]",
	Short: "Retrieves and displays one or more guidance entities by their ID(s) via the backend.",
	Long: `Retrieves and displays the content of one or more guidance entities
from the configured backend, based on their IDs.

You can specify the output format using the --output flag:
- structured (default): Displays ID, parsed frontmatter, and body as a JSON object (or array).
- json-frontmatter: Displays only ID and parsed frontmatter as a JSON object (or array).
- yaml-frontmatter: Displays only parsed frontmatter as YAML.
- body: Displays only the Markdown body content.
- raw: Displays the raw file content as returned by the backend (includes frontmatter and body delimiters).`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		idsToGet := args
		backend, _ := GetActiveBackend()
		if backend == nil {
			slog.Error("Active backend not initialized")
			return fmt.Errorf("active backend not initialized; run 'gydnc init' or check config")
		}

		// Prepare slices for collecting multiple results for JSON array outputs
		var structuredResults []StructuredGuidanceOutput
		var jsonFrontmatterResults []JsonFrontmatterOnlyOutput

		if len(idsToGet) > 1 {
			switch outputFormatGet {
			case "structured":
				structuredResults = make([]StructuredGuidanceOutput, 0, len(idsToGet))
			case "json-frontmatter":
				jsonFrontmatterResults = make([]JsonFrontmatterOnlyOutput, 0, len(idsToGet))
			}
		}

		for i, id := range idsToGet {
			slog.Debug("Attempting to get guidance from backend", "id", id, "format", outputFormatGet, "backend", backend.GetName())

			contentBytes, _, err := backend.Read(id)
			if err != nil {
				slog.Error("Failed to read guidance from backend", "id", id, "backend", backend.GetName(), "error", err)
				fmt.Fprintf(os.Stderr, "Error getting ID %s: %v\n", id, err)
				// Add placeholder for error to respective results slice if in multi-ID JSON mode
				if len(idsToGet) > 1 {
					switch outputFormatGet {
					case "structured":
						structuredResults = append(structuredResults, StructuredGuidanceOutput{ID: id, Body: "ERROR_FETCHING_CONTENT"})
					case "json-frontmatter":
						jsonFrontmatterResults = append(jsonFrontmatterResults, JsonFrontmatterOnlyOutput{ID: id, Title: "ERROR_FETCHING_CONTENT"})
					}
				}
				continue
			}

			// Separators for multi-output non-JSON array formats
			if outputFormatGet != "structured" && outputFormatGet != "json-frontmatter" {
				if len(idsToGet) > 1 && i > 0 {
					if outputFormatGet == "yaml-frontmatter" {
						fmt.Fprintln(os.Stdout, "---")
					} else {
						fmt.Fprintf(os.Stdout, "\n--- Content for ID: %s ---\n", id)
					}
				} else if len(idsToGet) > 1 && i == 0 && outputFormatGet != "yaml-frontmatter" && outputFormatGet != "raw" {
					// Header for first item in non-yaml/raw multi-item output
					fmt.Fprintf(os.Stdout, "--- Content for ID: %s ---\n", id)
				}
			}

			parsedContent, parseErr := content.ParseG6E(contentBytes)
			if parseErr != nil {
				slog.Error("Failed to parse G6E content", "id", id, "format", outputFormatGet, "error", parseErr)
				fmt.Fprintf(os.Stderr, "Error parsing ID %s: %v\n", id, parseErr)
				// Add placeholder for error to respective results slice if in multi-ID JSON mode and parsing is needed
				if len(idsToGet) > 1 {
					switch outputFormatGet {
					case "structured":
						structuredResults = append(structuredResults, StructuredGuidanceOutput{ID: id, Body: "ERROR_PARSING_CONTENT"})
					case "json-frontmatter":
						jsonFrontmatterResults = append(jsonFrontmatterResults, JsonFrontmatterOnlyOutput{ID: id, Title: "ERROR_PARSING_CONTENT"})
						// yaml-frontmatter and body also need parsing but don't collect into a single JSON array at the end
					}
				}
				// For raw, we don't parse, so we don't hit this error before the switch.
				// For formats that need parsing, continue to next ID if parse fails.
				if outputFormatGet != "raw" {
					continue
				}
			}

			switch outputFormatGet {
			case "structured":
				// parseErr already handled above for this case
				contentID, idErr := parsedContent.GetContentID()
				if idErr != nil {
					slog.Error("Failed to compute content ID", "source_id_arg", id, "error", idErr)
					// Decide how to handle this - perhaps use a placeholder or the original id_arg?
					contentID = "ERROR_COMPUTING_ID_" + id // Placeholder
				}
				structuredData := StructuredGuidanceOutput{
					ID: contentID,
					Frontmatter: FrontmatterData{
						Title:       parsedContent.Title,
						Description: parsedContent.Description,
						Tags:        parsedContent.Tags,
					},
					Body: parsedContent.Body,
				}
				if len(idsToGet) > 1 {
					structuredResults = append(structuredResults, structuredData)
				} else {
					jsonBytes, err := json.MarshalIndent(structuredData, "", "  ")
					if err != nil {
						slog.Error("Failed to marshal structured data to JSON", "id", id, "error", err)
						fmt.Fprintf(os.Stderr, "Error marshalling structured JSON for ID %s: %v\n", id, err)
						continue
					}
					fmt.Fprintln(os.Stdout, string(jsonBytes))
				}
			case "json-frontmatter":
				// parseErr already handled above for this case
				contentID, idErr := parsedContent.GetContentID()
				if idErr != nil {
					slog.Error("Failed to compute content ID for json-frontmatter", "source_id_arg", id, "error", idErr)
					contentID = "ERROR_COMPUTING_ID_" + id // Placeholder
				}
				jsonData := JsonFrontmatterOnlyOutput{
					ID:          contentID,
					Title:       parsedContent.Title,
					Description: parsedContent.Description,
					Tags:        parsedContent.Tags,
				}
				if len(idsToGet) > 1 {
					jsonFrontmatterResults = append(jsonFrontmatterResults, jsonData)
				} else {
					jsonBytes, err := json.MarshalIndent(jsonData, "", "  ")
					if err != nil {
						slog.Error("Failed to marshal frontmatter to JSON", "id", id, "error", err)
						fmt.Fprintf(os.Stderr, "Error marshalling JSON frontmatter for ID %s: %v\n", id, err)
						continue
					}
					fmt.Fprintln(os.Stdout, string(jsonBytes))
				}
			case "yaml-frontmatter": // Was "yaml"
				// parseErr already handled above for this case
				yamlBytes, err := parsedContent.MarshalFrontmatter() // This method produces YAML
				if err != nil {
					slog.Error("Failed to marshal frontmatter to YAML", "id", id, "error", err)
					fmt.Fprintf(os.Stderr, "Error marshalling YAML for ID %s: %v\n", id, err)
					continue
				}
				fmt.Fprint(os.Stdout, string(yamlBytes))
			case "raw": // "full" is an alias for raw conceptually, just print bytes
				// No parsing needed for raw, so parseErr check above is skipped for this case by design.
				fmt.Fprint(os.Stdout, string(contentBytes))
				// Ensure newline for consistent multi-output or single item display
				if len(contentBytes) > 0 && contentBytes[len(contentBytes)-1] != '\n' {
					fmt.Fprintln(os.Stdout)
				}

			case "body":
				// parseErr already handled above for this case
				fmt.Fprint(os.Stdout, parsedContent.Body)
				// Ensure newline for consistent multi-output or single item display
				if len(parsedContent.Body) > 0 && parsedContent.Body[len(parsedContent.Body)-1] != '\n' {
					fmt.Fprintln(os.Stdout)
				}
			default:
				slog.Error("Unknown output format specified", "format", outputFormatGet)
				return fmt.Errorf("unknown output format: %s. Valid formats are: structured, json-frontmatter, yaml-frontmatter, body, raw", outputFormatGet)
			}
		}

		// Finalize JSON array outputs if multiple IDs were processed
		if len(idsToGet) > 1 {
			switch outputFormatGet {
			case "structured":
				finalJsonBytes, err := json.MarshalIndent(structuredResults, "", "  ")
				if err != nil {
					slog.Error("Failed to marshal final structured JSON array", "error", err)
					return fmt.Errorf("marshalling final structured JSON array: %w", err)
				}
				fmt.Fprintln(os.Stdout, string(finalJsonBytes))
			case "json-frontmatter":
				finalJsonBytes, err := json.MarshalIndent(jsonFrontmatterResults, "", "  ")
				if err != nil {
					slog.Error("Failed to marshal final JSON frontmatter array", "error", err)
					return fmt.Errorf("marshalling final JSON frontmatter array: %w", err)
				}
				fmt.Fprintln(os.Stdout, string(finalJsonBytes))
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().StringVarP(&outputFormatGet, "output", "o", "structured", "Output format (structured, json-frontmatter, yaml-frontmatter, body, raw)")
}
