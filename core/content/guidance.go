package content

import (
	"bytes"
	// "crypto/sha256" // No longer needed directly
	// "encoding/hex"  // No longer needed directly
	"errors"
	"fmt"
	"strings"

	"gydnc/internal/utils" // Added import for our utils package

	"gopkg.in/yaml.v3"
)

// IMPORTANT: g6e File Format Backward Compatibility Notice
//
// The g6e file format is considered "extend only" for backward compatibility.
// - The YAML frontmatter structure may be extended with new fields
// - Existing fields must not be renamed or removed
// - The format (---\n YAML \n---\n content) must be preserved
//
// Changes to this format must be carefully considered and documented.

const (
	frontmatterDelimiter = "---"
	delimiterNewLine     = "\n"
)

// GuidanceContent holds the parsed content of a .g6e file.
//
// EXTEND ONLY: This structure may be extended with new fields but existing
// fields must not be renamed or removed to maintain backward compatibility.
type GuidanceContent struct {
	Title       string   `yaml:"title"`
	Description string   `yaml:"description,omitempty"`
	Tags        []string `yaml:"tags,omitempty"`
	// Body is not part of YAML, it's the content after the second '---'
	Body string `yaml:"-"` // Ignored by YAML marshaller/unmarshaller
}

// frontmatterYAML is a temporary struct used for marshalling only the YAML frontmatter fields.
// This prevents the Body field of GuidanceContent from being included in the YAML output.
type frontmatterYAML struct {
	Title       string   `yaml:"title"`
	Description string   `yaml:"description,omitempty"`
	Tags        []string `yaml:"tags,omitempty"`
}

// StandardFrontmatter defines the complete set of metadata fields for a new guidance entity.
//
// EXTEND ONLY: This structure may be extended with new fields but existing
// fields must not be renamed or removed to maintain backward compatibility.
type StandardFrontmatter struct {
	// ID          string   `yaml:"id"` // Removed: ID is now content-addressable
	Title       string   `yaml:"title"`
	Description string   `yaml:"description,omitempty"`
	Tags        []string `yaml:"tags,omitempty"`
}

// ParseG6E takes the raw byte content of a .g6e file and parses it
// into its frontmatter (unmarshaled into GuidanceContent) and Markdown body.
// It enforces strict frontmatter delimiter rules: starts with "---\n" and has a closing "\n---\n".
func ParseG6E(fileContent []byte) (*GuidanceContent, error) {
	openingDelimiterBytes := []byte(frontmatterDelimiter + delimiterNewLine)
	closingDelimiterBytes := []byte(delimiterNewLine + frontmatterDelimiter + delimiterNewLine)

	if !bytes.HasPrefix(fileContent, openingDelimiterBytes) {
		return nil, errors.New("malformed guidance: must start with '---\n' delimiter")
	}

	// End of the opening delimiter
	startOfYaml := len(openingDelimiterBytes)

	// Find the start of the closing delimiter "\n---\n"
	// This search must begin *after* the opening delimiter's content.
	endOfYaml := bytes.Index(fileContent[startOfYaml:], closingDelimiterBytes)
	if endOfYaml == -1 {
		return nil, errors.New("malformed guidance: missing closing '\n---\n' delimiter for frontmatter")
	}
	// Adjust endOfYaml to be relative to the original fileContent slice
	endOfYaml += startOfYaml

	yamlData := fileContent[startOfYaml:endOfYaml]
	bodyContent := fileContent[endOfYaml+len(closingDelimiterBytes):]

	var gc GuidanceContent
	err := yaml.Unmarshal(yamlData, &gc)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML frontmatter: %w", err)
	}

	gc.Body = string(bodyContent)

	return &gc, nil
}

// ToFileContent serializes a GuidanceContent struct back into a byte slice
// formatted as a .g6e file (YAML frontmatter + Markdown body).
func (gc *GuidanceContent) ToFileContent() ([]byte, error) {
	fm := frontmatterYAML{
		Title:       gc.Title,
		Description: gc.Description,
		Tags:        gc.Tags,
	}
	yamlData, err := yaml.Marshal(&fm)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal YAML frontmatter: %w", err)
	}

	var buffer bytes.Buffer
	buffer.WriteString(frontmatterDelimiter + delimiterNewLine)
	buffer.Write(yamlData)
	// yaml.Marshal for structs usually adds a trailing newline if the output is not empty.
	// If yamlData is empty (e.g. all fields in frontmatterYAML were empty and omitempty took effect),
	// it might not have a newline. We need to ensure there's one before the closing delimiter.
	if len(yamlData) == 0 || (len(yamlData) > 0 && yamlData[len(yamlData)-1] != '\n') {
		buffer.WriteString(delimiterNewLine)
	}
	buffer.WriteString(frontmatterDelimiter + delimiterNewLine)
	body := gc.Body
	if body != "" && !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	buffer.WriteString(body)

	return buffer.Bytes(), nil
}

// MarshalFrontmatter serializes only the frontmatter-related fields (Title, Description, Tags)
// of the GuidanceContent to a YAML byte slice.
func (gc *GuidanceContent) MarshalFrontmatter() ([]byte, error) {
	fm := frontmatterYAML{ // Uses the internal, unexported struct
		Title:       gc.Title,
		Description: gc.Description,
		Tags:        gc.Tags,
	}
	return yaml.Marshal(&fm)
}

// GetContentID computes and returns the SHA256 hash of the Body content.
// This serves as the content-addressable ID for the guidance.
func (gc *GuidanceContent) GetContentID() (string, error) {
	if gc == nil {
		return "", errors.New("cannot get content ID from nil GuidanceContent")
	}
	return utils.Sha256([]byte(gc.Body)), nil
}
