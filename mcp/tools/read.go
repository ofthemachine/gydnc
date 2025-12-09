package tools

import (
	"context"
	"fmt"

	"gydnc/mcp/tools/format"
	"gydnc/mcp/tools/types"
	"gydnc/service"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var GuidanceReadTool = &mcp.Tool{
	Name:        "gydnc_read",
	Description: "Read guidance entities from the gydnc knowledge base. Supports two operations: 'list' to discover available entities with optional tag filtering, and 'get' to retrieve full content of entities by alias. Use 'list' first to discover what guidance is available, then 'get' to fetch full content. Fetching multiple entities in one 'get' call is more efficient than separate calls.",
	Annotations: &mcp.ToolAnnotations{
		ReadOnlyHint: true,
	},
}

type GuidanceReadInput struct {
	Operation  string   `json:"operation" jsonschema:"the operation to perform: 'list' or 'get'"`
	FilterTags string   `json:"filter_tags,omitempty" jsonschema:"for 'list' operation: tag filter expression (e.g., 'scope:code quality:safety', '-deprecated', 'scope:*')"`
	Aliases    []string `json:"aliases,omitempty" jsonschema:"for 'get' operation: one or more guidance aliases to retrieve"`
}

type GuidanceReadOutput struct {
	Operation string      `json:"operation" jsonschema:"the operation that was performed"`
	Entities  interface{} `json:"entities" jsonschema:"list operation returns array of {alias, title, tags}; get operation returns array of {title, description, tags, body}"`
}

// Use types from the types package
type GuidanceListItem = types.GuidanceListItem
type GuidanceGetItem = types.GuidanceGetItem

func GuidanceRead(ctx context.Context, req *mcp.CallToolRequest, input GuidanceReadInput) (
	*mcp.CallToolResult,
	GuidanceReadOutput,
	error,
) {
	if AppContext == nil {
		return nil, GuidanceReadOutput{}, fmt.Errorf("application context not initialized")
	}

	entityService := service.NewEntityService(AppContext)

	switch input.Operation {
	case "list":
		return handleListOperation(ctx, entityService, input.FilterTags)
	case "get":
		return handleGetOperation(ctx, entityService, input.Aliases)
	default:
		return nil, GuidanceReadOutput{}, fmt.Errorf("invalid operation '%s': must be 'list' or 'get'", input.Operation)
	}
}

func handleListOperation(ctx context.Context, entityService *service.EntityService, filterTags string) (
	*mcp.CallToolResult,
	GuidanceReadOutput,
	error,
) {
	entities, backendErrors := entityService.ListEntitiesMerged("", filterTags)

	// Log backend errors but don't fail the request
	if len(backendErrors) > 0 {
		for backendName, err := range backendErrors {
			AppContext.Logger.Warn("Error accessing backend during list operation", "backend", backendName, "error", err)
		}
	}

	// Convert to output format (without description to reduce context bloat)
	items := make([]GuidanceListItem, len(entities))
	for i, entity := range entities {
		items[i] = GuidanceListItem{
			Alias: entity.Alias,
			Title: entity.Title,
			Tags:  entity.Tags,
		}
	}

	// Format as markdown using formatter
	markdown := format.FormatListOutput(items)

	return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: markdown,
				},
			},
		}, GuidanceReadOutput{
			Operation: "list",
			Entities:  items,
		}, nil
}

func handleGetOperation(ctx context.Context, entityService *service.EntityService, aliases []string) (
	*mcp.CallToolResult,
	GuidanceReadOutput,
	error,
) {
	if len(aliases) == 0 {
		return nil, GuidanceReadOutput{}, fmt.Errorf("at least one alias must be provided for 'get' operation")
	}

	items := make([]GuidanceGetItem, 0, len(aliases))

	for _, alias := range aliases {
		entity, err := entityService.GetEntity(alias, "")
		if err != nil {
			// Continue with other entities even if one fails
			AppContext.Logger.Warn("Failed to get entity", "alias", alias, "error", err)
			// Add error item to output
			items = append(items, GuidanceGetItem{
				Title: fmt.Sprintf("ERROR_FETCHING_CONTENT_FOR_%s", alias),
				Body:  fmt.Sprintf("Error: %v", err),
			})
			continue
		}

		items = append(items, GuidanceGetItem{
			Title:       entity.Title,
			Description: entity.Description,
			Tags:        entity.Tags,
			Body:        entity.Body,
		})
	}

	// Format as markdown using formatter
	markdown := format.FormatGetOutput(items)

	return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: markdown,
				},
			},
		}, GuidanceReadOutput{
			Operation: "get",
			Entities:  items,
		}, nil
}
