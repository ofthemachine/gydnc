package service

import (
	"fmt"
	"io/fs"

	"gydnc/filter"
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

// ListEntitiesMerged returns a list of entities from all configured backends that match the given prefix.
// Entities are deduplicated based on alias, with priority given to the default backend and then to
// backends in the order they are defined in the config.
// If filterString is provided, only entities matching the filter will be returned.
func (s *EntityService) ListEntitiesMerged(prefix string, filterString string) ([]model.Entity, map[string]error) {
	// Get the entities from all backends
	backendEntities, backendErrors := s.ListEntities(prefix)

	// Create a map to deduplicate entities by alias
	uniqueEntities := make(map[string]model.Entity)

	// Get the default backend name for prioritization
	defaultBackendName := s.ctx.Config.DefaultBackend

	// First, add entities from the default backend if available
	if defaultBackendName != "" {
		if entities, ok := backendEntities[defaultBackendName]; ok {
			for _, entity := range entities {
				uniqueEntities[entity.Alias] = entity
			}
		}
	}

	// Then add entities from other backends, preserving existing ones in case of duplicates
	for backendName, entities := range backendEntities {
		// Skip default backend as we already processed it
		if backendName == defaultBackendName {
			continue
		}

		for _, entity := range entities {
			// Only add if we haven't seen this alias before
			if _, exists := uniqueEntities[entity.Alias]; !exists {
				uniqueEntities[entity.Alias] = entity
			}
		}
	}

	// Convert map to slice
	mergedEntities := make([]model.Entity, 0, len(uniqueEntities))
	for _, entity := range uniqueEntities {
		mergedEntities = append(mergedEntities, entity)
	}

	// Apply filter if provided
	if filterString != "" {
		var err error
		mergedEntities, err = s.FilterEntities(mergedEntities, filterString)
		if err != nil {
			s.ctx.Logger.Warn("Error applying filter", "filter", filterString, "error", err)
		}
	}

	return mergedEntities, backendErrors
}

// FilterEntities applies a filter string to a list of entities.
// It uses the filter package to perform the filtering.
func (s *EntityService) FilterEntities(entities []model.Entity, filterString string) ([]model.Entity, error) {
	if filterString == "" {
		return entities, nil
	}

	// Import the filter package
	f, err := filter.NewFilterFromString(filterString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse filter string: %w", err)
	}

	return f.Filter(entities), nil
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

		// Read the entity content and metadata
		content, metadata, err := backendToUse.Read(alias)
		if err != nil {
			return entity, fmt.Errorf("failed to read entity %s from backend %s: %w", alias, backendToUse.GetName(), err)
		}

		// Create an Entity with the information
		entity = s.createEntityFromBackendData(alias, backendToUse.GetName(), content, metadata)
		return entity, nil
	} else {
		// If no backend specified, search through backends in priority order
		s.ctx.Logger.Debug("No backend specified, searching through all backends in priority order", "alias", alias)

		// Get all backends
		backends, backendErrors := s.ctx.GetAllBackends()
		if len(backends) == 0 {
			// No backends available
			return entity, fmt.Errorf("no backends available: %v", backendErrors)
		}

		// Try default backend first if configured
		defaultBackendName := s.ctx.Config.DefaultBackend
		if defaultBackendName != "" {
			if defaultBackend, ok := backends[defaultBackendName]; ok {
				content, metadata, err := defaultBackend.Read(alias)
				if err == nil {
					// Found in default backend
					entity = s.createEntityFromBackendData(alias, defaultBackend.GetName(), content, metadata)
					return entity, nil
				}
				// Log the error but continue with other backends
				s.ctx.Logger.Debug("Entity not found in default backend", "backend", defaultBackendName, "alias", alias, "error", err)
			}
		}

		// Try all other backends in the order defined in the config
		// Note: map iteration order is not guaranteed, but we're maintaining ordering based on config
		for name, backend := range backends {
			// Skip default backend as we already tried it
			if name == defaultBackendName {
				continue
			}

			content, metadata, err := backend.Read(alias)
			if err == nil {
				// Found in this backend
				entity = s.createEntityFromBackendData(alias, backend.GetName(), content, metadata)
				return entity, nil
			}
			// Log the error but continue with other backends
			s.ctx.Logger.Debug("Entity not found in backend", "backend", name, "alias", alias, "error", err)
		}

		// Entity not found in any backend
		return entity, fmt.Errorf("entity %s not found in any available backend", alias)
	}
}

// createEntityFromBackendData is a helper function to create an Entity from backend data
// This reduces duplication in the GetEntity method
func (s *EntityService) createEntityFromBackendData(alias string, backendName string, content []byte, metadata map[string]interface{}) model.Entity {
	entity := model.Entity{
		Alias:         alias,
		SourceBackend: backendName,
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

	return entity
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
