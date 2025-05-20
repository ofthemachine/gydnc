package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"gydnc/core/content"
	"gydnc/model"

	"github.com/spf13/cobra"
)

var outputFormatGet string

// --- Structs for "structured" JSON output ---
// OLD STRUCTS:
// type StructuredGuidanceOutput struct {
// 	ID          string          `json:"id"`
// 	Frontmatter FrontmatterData `json:"frontmatter"`
// 	Body        string          `json:"body"`
// }
//
// type FrontmatterData struct {
// 	Title       string   `json:"title"`
// 	Description string   `json:"description,omitempty"`
// 	Tags        []string `json:"tags,omitempty"`
// }

// NEW SIMPLIFIED STRUCT for "structured" (default) JSON output
type SimplifiedStructuredOutput struct {
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Body        string   `json:"body"`
}

// --- End structs for "structured" JSON output ---

// Helper struct for "json-frontmatter" output (was jsonFrontmatterOutput)
type JsonFrontmatterOnlyOutput struct {
	// ID          string   `json:"id"` // Removed: do not output 'id' in json-frontmatter mode
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
- body: Displays only the Markdown body content.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Set isJSONMode for machine-readable output
		if outputFormatGet == "structured" || outputFormatGet == "json-frontmatter" {
			isJSONMode = true
		}

		idsToGet := args
		backend, _ := GetActiveBackend()
		if backend == nil {
			slog.Error("Active backend not initialized")
			return fmt.Errorf("active backend not initialized; run 'gydnc init' or check config")
		}

		var simplifiedStructuredResults []SimplifiedStructuredOutput // NEW
		var jsonFrontmatterResults []JsonFrontmatterOnlyOutput

		if len(idsToGet) > 1 {
			switch outputFormatGet {
			case "structured":
				simplifiedStructuredResults = make([]SimplifiedStructuredOutput, 0, len(idsToGet)) // NEW
			case "json-frontmatter":
				jsonFrontmatterResults = make([]JsonFrontmatterOnlyOutput, 0, len(idsToGet))
			}
		}

		for _, id := range idsToGet {
			slog.Debug("Attempting to get guidance from backend", "id", id, "format", outputFormatGet, "backend", backend.GetName())

			contentBytes, meta, err := backend.Read(id)
			if err != nil {
				slog.Error("Failed to read guidance from backend", "id", id, "backend", backend.GetName(), "error", err)
				fmt.Fprintf(os.Stderr, "Error getting ID %s: %v\n", id, err)
				// Add placeholder for error to respective results slice if in multi-ID JSON mode
				if len(idsToGet) > 1 {
					switch outputFormatGet {
					case "structured":
						simplifiedStructuredResults = append(simplifiedStructuredResults, SimplifiedStructuredOutput{Title: "ERROR_FETCHING_CONTENT_FOR_" + id, Body: "ERROR_FETCHING_CONTENT"}) // NEW
					case "json-frontmatter":
						jsonFrontmatterResults = append(jsonFrontmatterResults, JsonFrontmatterOnlyOutput{Title: "ERROR_FETCHING_CONTENT"})
					}
				}
				continue
			}

			parsedContent, parseErr := content.ParseG6E(contentBytes)
			if parseErr != nil {
				slog.Error("Failed to parse G6E content", "id", id, "format", outputFormatGet, "error", parseErr)
				fmt.Fprintf(os.Stderr, "Error parsing ID %s: %v\n", id, parseErr)
				// Add placeholder for error to respective results slice if in multi-ID JSON mode and parsing is needed
				if len(idsToGet) > 1 {
					switch outputFormatGet {
					case "structured":
						simplifiedStructuredResults = append(simplifiedStructuredResults, SimplifiedStructuredOutput{Title: "ERROR_PARSING_CONTENT_FOR_" + id, Body: "ERROR_PARSING_CONTENT"}) // NEW
					case "json-frontmatter":
						jsonFrontmatterResults = append(jsonFrontmatterResults, JsonFrontmatterOnlyOutput{Title: "ERROR_PARSING_CONTENT"})
						// yaml-frontmatter and body also need parsing but don't collect into a single JSON array at the end
					}
				}
				// Since all remaining formats require parsing, if parsing fails, continue to the next ID.
				continue
			}

			entity := model.Entity{
				Alias:          id,
				SourceBackend:  backend.GetName(),
				Title:          parsedContent.Title,
				Description:    parsedContent.Description,
				Tags:           parsedContent.Tags,
				CustomMetadata: meta, // Optionally filter meta fields
				Body:           parsedContent.Body,
			}
			cid, _ := parsedContent.GetContentID()
			entity.CID = cid

			switch outputFormatGet {
			case "structured":
				structuredData := struct {
					Title       string   `json:"title"`
					Description string   `json:"description,omitempty"`
					Tags        []string `json:"tags,omitempty"`
					Body        string   `json:"body"`
				}{
					Title:       entity.Title,
					Description: entity.Description,
					Tags:        entity.Tags,
					Body:        entity.Body,
				}
				if len(idsToGet) > 1 {
					simplifiedStructuredResults = append(simplifiedStructuredResults, SimplifiedStructuredOutput{
						Title:       entity.Title,
						Description: entity.Description,
						Tags:        entity.Tags,
						Body:        entity.Body,
					})
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
				jsonData := JsonFrontmatterOnlyOutput{
					// ID:          entity.CID, // Removed: do not output 'id' in json-frontmatter mode
					Title:       entity.Title,
					Description: entity.Description,
					Tags:        entity.Tags,
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
				yamlBytes, err := parsedContent.MarshalFrontmatter() // This method produces YAML
				if err != nil {
					slog.Error("Failed to marshal frontmatter to YAML", "id", id, "error", err)
					fmt.Fprintf(os.Stderr, "Error marshalling YAML for ID %s: %v\n", id, err)
					continue
				}
				fmt.Fprint(os.Stdout, string(yamlBytes))
			case "body":
				fmt.Fprint(os.Stdout, entity.Body)
				if len(entity.Body) > 0 && entity.Body[len(entity.Body)-1] != '\n' {
					fmt.Fprintln(os.Stdout)
				}
			default:
				slog.Error("Unknown output format specified", "format", outputFormatGet)
				return fmt.Errorf("unknown output format: %s. Valid formats are: structured, json-frontmatter, yaml-frontmatter, body", outputFormatGet)
			}
		}

		// Finalize JSON array outputs if multiple IDs were processed
		if len(idsToGet) > 1 {
			switch outputFormatGet {
			case "structured":
				finalJsonBytes, err := json.MarshalIndent(simplifiedStructuredResults, "", "  ") // NEW
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

		// Do not print any entity list, summary, or extra output after the JSON/YAML/body output.
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().StringVarP(&outputFormatGet, "output", "o", "structured", "Output format (structured, json-frontmatter, yaml-frontmatter, body)")
}
