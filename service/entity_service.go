package service

import (
	"fmt"
	"io/fs"

	"gydnc/model"
	"gydnc/storage"
)

// EntityService provides methods for interacting with guidance entities.
type EntityService struct {
	ctx *AppContext
}

// NewEntityService creates a new EntityService with the provided context.
func NewEntityService(ctx *AppContext) *EntityService {
	return &EntityService{
		ctx: ctx,
	}
}

// ListEntities returns a list of entities from all configured backends that match the given prefix.
// Entities are organized by backend, and backend errors are returned separately.
func (s *EntityService) ListEntities(prefix string) (map[string][]model.Entity, map[string]error) {
	backends, backendErrors := s.ctx.GetAllBackends()
	results := make(map[string][]model.Entity)

	for name, backend := range backends {
		s.ctx.Logger.Debug("Listing entities from backend", "backend", name, "prefix", prefix)

		// Get the list of entity aliases from the backend
		aliases, err := backend.List(prefix)
		if err != nil {
			backendErrors[name] = fmt.Errorf("failed to list entities from backend %s: %w", name, err)
			continue
		}

		// Create a model.Entity for each alias
		var entities []model.Entity
		for _, alias := range aliases {
			// Get metadata for the entity
			metadata, err := backend.Stat(alias)
			if err != nil && err != fs.ErrNotExist {
				// Log the error but continue with other entities
				s.ctx.Logger.Warn("Failed to get metadata for entity", "backend", name, "alias", alias, "error", err)
				continue
			}

			// Create an Entity with the available information
			entity := model.Entity{
				Alias:         alias,
				SourceBackend: backend.GetName(),
			}

			// Extract common metadata fields if available
			if metadata != nil {
				if title, ok := metadata["title"].(string); ok {
					entity.Title = title
				}
				if desc, ok := metadata["description"].(string); ok {
					entity.Description = desc
				}
				if tags, ok := metadata["tags"].([]string); ok {
					entity.Tags = tags
				}
				// Additional metadata goes into CustomMetadata
				entity.CustomMetadata = make(map[string]interface{})
				for k, v := range metadata {
					switch k {
					case "title", "description", "tags":
						// Skip fields already handled
					default:
						entity.CustomMetadata[k] = v
					}
				}
			}

			// Add CID if available
			if cid, ok := metadata["cid"].(string); ok {
				entity.CID = cid
			}

			// Add pCID if available
			if pcid, ok := metadata["pcid"].(string); ok {
				entity.PCID = pcid
			}

			entities = append(entities, entity)
		}

		results[name] = entities
	}

	return results, backendErrors
}

// GetEntity retrieves a single entity from the specified backend.
// If backendName is empty, it searches all backends with priority given to the default backend.
func (s *EntityService) GetEntity(alias string, backendName string) (model.Entity, error) {
	var entity model.Entity
	var backendToUse storage.ReadOnlyBackend
	var err error

	// If backend name is provided, use that specific backend
	if backendName != "" {
		backendToUse, err = s.ctx.GetBackend(backendName)
		if err != nil {
			return entity, fmt.Errorf("failed to get backend %s: %w", backendName, err)
		}
	} else {
		// Otherwise, try to get the default backend
		backendToUse, err = s.ctx.GetDefaultBackend()
		if err != nil {
			return entity, fmt.Errorf("no backend specified and no default backend available: %w", err)
		}
	}

	// Read the entity content and metadata
	content, metadata, err := backendToUse.Read(alias)
	if err != nil {
		return entity, fmt.Errorf("failed to read entity %s from backend %s: %w", alias, backendToUse.GetName(), err)
	}

	// Create an Entity with the information
	entity = model.Entity{
		Alias:         alias,
		SourceBackend: backendToUse.GetName(),
		Body:          string(content),
	}

	// Extract common metadata fields if available
	if metadata != nil {
		if title, ok := metadata["title"].(string); ok {
			entity.Title = title
		}
		if desc, ok := metadata["description"].(string); ok {
			entity.Description = desc
		}
		if tags, ok := metadata["tags"].([]string); ok {
			entity.Tags = tags
		}
		// Additional metadata goes into CustomMetadata
		entity.CustomMetadata = make(map[string]interface{})
		for k, v := range metadata {
			switch k {
			case "title", "description", "tags":
				// Skip fields already handled
			default:
				entity.CustomMetadata[k] = v
			}
		}
	}

	// Add CID if available
	if cid, ok := metadata["cid"].(string); ok {
		entity.CID = cid
	}

	// Add pCID if available
	if pcid, ok := metadata["pcid"].(string); ok {
		entity.PCID = pcid
	}

	return entity, nil
}

// SaveEntity saves an entity to the specified backend.
// If the backend is read-only, an error is returned.
func (s *EntityService) SaveEntity(entity model.Entity, backendName string) error {
	var backendToUse storage.ReadOnlyBackend
	var err error

	// If backend name is provided, use that specific backend
	if backendName != "" {
		backendToUse, err = s.ctx.GetBackend(backendName)
		if err != nil {
			return fmt.Errorf("failed to get backend %s: %w", backendName, err)
		}
	} else {
		// Otherwise, try to get the default backend
		backendToUse, err = s.ctx.GetDefaultBackend()
		if err != nil {
			return fmt.Errorf("no backend specified and no default backend available: %w", err)
		}
	}

	// Check if the backend is writable
	if !backendToUse.IsWritable() {
		return fmt.Errorf("backend %s is read-only", backendToUse.GetName())
	}

	// Cast to Backend (writable) interface
	writableBackend, ok := backendToUse.(storage.Backend)
	if !ok {
		return fmt.Errorf("backend %s does not implement the writable Backend interface", backendToUse.GetName())
	}

	// Prepare commit message details (map[string]string required by Backend.Write)
	commitMsg := map[string]string{
		"action": "save",
		"alias":  entity.Alias,
	}

	// Add CID and pCID to commitMsg if available
	if entity.CID != "" {
		commitMsg["cid"] = entity.CID
	}
	if entity.PCID != "" {
		commitMsg["pcid"] = entity.PCID
	}

	// Convert entity body to bytes
	content := []byte(entity.Body)

	// Write the entity
	err = writableBackend.Write(entity.Alias, content, commitMsg)
	if err != nil {
		return fmt.Errorf("failed to write entity %s to backend %s: %w", entity.Alias, writableBackend.GetName(), err)
	}

	return nil
}

// DeleteEntity deletes an entity from the specified backend.
// If the backend is read-only, an error is returned.
func (s *EntityService) DeleteEntity(alias string, backendName string) error {
	var backendToUse storage.ReadOnlyBackend
	var err error

	// If backend name is provided, use that specific backend
	if backendName != "" {
		backendToUse, err = s.ctx.GetBackend(backendName)
		if err != nil {
			return fmt.Errorf("failed to get backend %s: %w", backendName, err)
		}
	} else {
		// Otherwise, try to get the default backend
		backendToUse, err = s.ctx.GetDefaultBackend()
		if err != nil {
			return fmt.Errorf("no backend specified and no default backend available: %w", err)
		}
	}

	// Check if the backend is writable
	if !backendToUse.IsWritable() {
		return fmt.Errorf("backend %s is read-only", backendToUse.GetName())
	}

	// Cast to Backend (writable) interface
	writableBackend, ok := backendToUse.(storage.Backend)
	if !ok {
		return fmt.Errorf("backend %s does not implement the writable Backend interface", backendToUse.GetName())
	}

	// Delete the entity
	err = writableBackend.Delete(alias)
	if err != nil {
		return fmt.Errorf("failed to delete entity %s from backend %s: %w", alias, writableBackend.GetName(), err)
	}

	return nil
}
