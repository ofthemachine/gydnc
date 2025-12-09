package format

import (
	"fmt"
	"strings"

	"gydnc/mcp/tools/types"
)

// FormatListOutput formats a list of guidance list items as markdown
func FormatListOutput(items []types.GuidanceListItem) string {
	var markdown strings.Builder
	markdown.WriteString(fmt.Sprintf("## Found %d guidance entities\n\n", len(items)))

	if len(items) > 0 {
		for _, item := range items {
			markdown.WriteString(fmt.Sprintf("### %s\n", item.Title))
			markdown.WriteString(fmt.Sprintf("**Alias:** `%s`\n", item.Alias))
			if len(item.Tags) > 0 {
				tagList := strings.Join(item.Tags, ", ")
				markdown.WriteString(fmt.Sprintf("**Tags:** %s\n", tagList))
			}
			markdown.WriteString("\n")
		}
	}

	return markdown.String()
}

// FormatGetOutput formats a list of guidance get items as markdown
func FormatGetOutput(items []types.GuidanceGetItem) string {
	var markdown strings.Builder
	for i, item := range items {
		if len(items) > 1 {
			markdown.WriteString(fmt.Sprintf("---\n\n## Entity %d of %d\n\n", i+1, len(items)))
		}

		// Check if this is an error item
		isError := strings.HasPrefix(item.Title, "ERROR_FETCHING_CONTENT_FOR_")

		if isError {
			markdown.WriteString(fmt.Sprintf("## ❌ Error\n\n**Failed to fetch:** `%s`\n\n%s\n",
				strings.TrimPrefix(item.Title, "ERROR_FETCHING_CONTENT_FOR_"), item.Body))
		} else {
			markdown.WriteString(fmt.Sprintf("# %s\n\n", item.Title))

			if item.Description != "" {
				markdown.WriteString(fmt.Sprintf("**Description:** %s\n\n", item.Description))
			}

			if len(item.Tags) > 0 {
				tagList := strings.Join(item.Tags, ", ")
				markdown.WriteString(fmt.Sprintf("**Tags:** %s\n\n", tagList))
			}

			if item.Body != "" {
				markdown.WriteString("---\n\n")
				markdown.WriteString(item.Body)
				if !strings.HasSuffix(item.Body, "\n") {
					markdown.WriteString("\n")
				}
			}
		}

		if i < len(items)-1 {
			markdown.WriteString("\n\n")
		}
	}

	return markdown.String()
}

// FormatWriteSuccessOutput formats a successful write operation as markdown
func FormatWriteSuccessOutput(output types.GuidanceWriteOutput) string {
	var emoji string
	var action string
	switch output.Operation {
	case "create":
		emoji = "✅"
		action = "Created"
	case "update":
		emoji = "✅"
		action = "Updated"
	default:
		emoji = "✅"
		action = "Completed"
	}

	return fmt.Sprintf("## %s Successfully %s\n\n**Alias:** `%s`\n**Backend:** `%s`\n", emoji, action, output.Alias, output.Backend)
}

// FormatWriteErrorOutput formats a failed write operation as markdown
func FormatWriteErrorOutput(output types.GuidanceWriteOutput) string {
	var action string
	switch output.Operation {
	case "create":
		action = "Create"
	case "update":
		action = "Update"
	default:
		action = "Operation"
	}

	return fmt.Sprintf("## ❌ Failed to %s\n\n**Alias:** `%s`\n**Error:** %s\n", action, output.Alias, output.Message)
}
