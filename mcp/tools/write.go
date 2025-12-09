package tools

import (
	"context"
	"errors"
	"fmt"

	"gydnc/mcp/tools/format"
	"gydnc/mcp/tools/types"
	"gydnc/model"
	"gydnc/service"
	"gydnc/storage"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var GuidanceWriteTool = &mcp.Tool{
	Name:        "gydnc_write",
	Description: "Write (create or update) guidance entities in the gydnc knowledge base. Supports two operations: 'create' to add a new entity, and 'update' to modify an existing entity. Both operations share the same parameter structure: alias (required), title, description, tags, and body (all optional). For 'update', only provided fields will be modified; existing values are preserved for omitted fields.",
	Annotations: &mcp.ToolAnnotations{
		ReadOnlyHint: false,
	},
}

type GuidanceWriteInput struct {
	Operation   string   `json:"operation" jsonschema:"the operation to perform: 'create' or 'update'"`
	Alias       string   `json:"alias" jsonschema:"the unique identifier for the guidance entity (required)"`
	Title       string   `json:"title,omitempty" jsonschema:"the title of the guidance entity (optional, for update: empty string means don't update)"`
	Description string   `json:"description,omitempty" jsonschema:"the description of the guidance entity (optional, for update: empty string means don't update)"`
	Tags        []string `json:"tags,omitempty" jsonschema:"tags associated with the guidance entity (optional, for update: empty array means don't update)"`
	Body        string   `json:"body,omitempty" jsonschema:"the body content of the guidance entity (optional, for update: empty string means don't update)"`
	Backend     string   `json:"backend,omitempty" jsonschema:"name of the storage backend to use (optional, uses default if not specified)"`
}

// Use type from the types package
type GuidanceWriteOutput = types.GuidanceWriteOutput

func GuidanceWrite(ctx context.Context, req *mcp.CallToolRequest, input GuidanceWriteInput) (
	*mcp.CallToolResult,
	GuidanceWriteOutput,
	error,
) {
	if AppContext == nil {
		return nil, GuidanceWriteOutput{}, fmt.Errorf("application context not initialized")
	}

	if input.Alias == "" {
		return nil, GuidanceWriteOutput{}, fmt.Errorf("alias is required")
	}

	entityService := service.NewEntityService(AppContext)

	switch input.Operation {
	case "create":
		return handleCreateOperation(ctx, entityService, input)
	case "update":
		return handleUpdateOperation(ctx, entityService, input)
	default:
		return nil, GuidanceWriteOutput{}, fmt.Errorf("invalid operation '%s': must be 'create' or 'update'", input.Operation)
	}
}

func handleCreateOperation(ctx context.Context, entityService *service.EntityService, input GuidanceWriteInput) (
	*mcp.CallToolResult,
	GuidanceWriteOutput,
	error,
) {
	// Build entity for creation
	entity := model.Entity{
		Alias:       input.Alias,
		Title:       input.Title,
		Description: input.Description,
		Tags:        input.Tags,
		Body:        input.Body,
	}

	// Provide default body if none specified (matching CLI behavior)
	// Note: We check for empty string, but preserve whitespace-only strings
	if input.Body == "" {
		if entity.Title == "" {
			entity.Body = "#\n\nGuidance content for '' goes here.\n"
		} else {
			entity.Body = fmt.Sprintf("# %s\n\nGuidance content for '%s' goes here.\n", entity.Title, entity.Title)
		}
	}

	backendName := input.Backend

	savedBackendName, err := entityService.SaveEntity(entity, backendName)
	if err != nil {
		errorOutput := GuidanceWriteOutput{
			Operation: "create",
			Alias:     input.Alias,
			Success:   false,
			Message:   err.Error(),
		}
		errorMarkdown := format.FormatWriteErrorOutput(errorOutput)

		// For expected business logic errors (like entity already exists), return as success with error content
		// Only return actual errors for unexpected system failures
		if errors.Is(err, storage.ErrEntityAlreadyExists) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: errorMarkdown,
					},
				},
			}, errorOutput, nil
		}

		// For unexpected errors, return as error
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: errorMarkdown,
				},
			},
		}, errorOutput, err
	}

	result := GuidanceWriteOutput{
		Operation: "create",
		Alias:     input.Alias,
		Backend:   savedBackendName,
		Success:   true,
		Message:   fmt.Sprintf("Successfully created entity '%s' in backend '%s'", input.Alias, savedBackendName),
	}

	// Format as markdown using formatter
	markdown := format.FormatWriteSuccessOutput(result)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: markdown,
			},
		},
	}, result, nil
}

func handleUpdateOperation(ctx context.Context, entityService *service.EntityService, input GuidanceWriteInput) (
	*mcp.CallToolResult,
	GuidanceWriteOutput,
	error,
) {
	// Get existing entity first
	existingEntity, err := entityService.GetEntity(input.Alias, "")
	if err != nil {
		errorOutput := GuidanceWriteOutput{
			Operation: "update",
			Alias:     input.Alias,
			Success:   false,
			Message:   fmt.Sprintf("failed to retrieve entity for update: %v", err),
		}
		errorMarkdown := format.FormatWriteErrorOutput(errorOutput)
		// Entity not found is an expected business logic error, return as success with error content
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: errorMarkdown,
				},
			},
		}, errorOutput, nil
	}

	// Apply updates only to provided fields (empty string means "don't update")
	// Note: We check for empty string, but preserve whitespace-only strings
	// Empty string == not provided, so we only update if the field is non-empty
	if input.Title != "" {
		existingEntity.Title = input.Title
	}
	if input.Description != "" {
		existingEntity.Description = input.Description
	}
	// For tags, empty array means "don't update", non-empty array means "replace all tags"
	if len(input.Tags) > 0 {
		existingEntity.Tags = input.Tags
	}
	if input.Body != "" {
		existingEntity.Body = input.Body
	}

	// Use the entity's source backend for update (or override if specified)
	backendName := input.Backend
	if backendName == "" {
		backendName = existingEntity.SourceBackend
	}

	savedBackendName, err := entityService.OverwriteEntity(existingEntity, backendName)
	if err != nil {
		errorOutput := GuidanceWriteOutput{
			Operation: "update",
			Alias:     input.Alias,
			Success:   false,
			Message:   err.Error(),
		}
		errorMarkdown := format.FormatWriteErrorOutput(errorOutput)
		// Most update errors are expected business logic errors, return as success with error content
		// Only return actual errors for unexpected system failures
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: errorMarkdown,
				},
			},
		}, errorOutput, nil
	}

	result := GuidanceWriteOutput{
		Operation: "update",
		Alias:     input.Alias,
		Backend:   savedBackendName,
		Success:   true,
		Message:   fmt.Sprintf("Successfully updated entity '%s' in backend '%s'", input.Alias, savedBackendName),
	}

	// Format as markdown using formatter
	markdown := format.FormatWriteSuccessOutput(result)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: markdown,
			},
		},
	}, result, nil
}
