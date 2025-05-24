package service

import (
	"errors"
	"fmt"
	"io/fs"
	"sort"

	"gydnc/core/content"
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
					sort.Strings(entity.Tags) // Ensure tags are sorted
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
// Entities from all backends are collected. If filterString is provided, only entities matching the filter will be returned.
// It ensures alias uniqueness, prioritizing the default backend, then lexical backend order for duplicates.
// The final list is sorted by Alias.
func (s *EntityService) ListEntitiesMerged(prefix string, filterString string) ([]model.Entity, map[string]error) {
	// Get the entities from all backends, grouped by backend name
	backendEntitiesMap, backendErrors := s.ListEntities(prefix) // This already sorts tags within each entity

	defaultBackendName := s.ctx.Config.DefaultBackend
	entitiesByAlias := make(map[string]model.Entity)
	// Keep track of which backends an alias was found in, for logging duplicates
	foundInBackendsByAlias := make(map[string][]string)

	// Collect all entities and track sources
	// var allCollectedEntitiesForProcessing []model.Entity // Removed, was unused

	// Get backend names and sort them for deterministic processing order of non-default backends
	var backendNames []string
	for name := range backendEntitiesMap {
		backendNames = append(backendNames, name)
	}
	sort.Strings(backendNames)

	// First, process the default backend if it exists and has entities
	if defaultBackendName != "" {
		if entities, ok := backendEntitiesMap[defaultBackendName]; ok {
			for _, entity := range entities {
				entitiesByAlias[entity.Alias] = entity // Default backend version takes precedence
				foundInBackendsByAlias[entity.Alias] = append(foundInBackendsByAlias[entity.Alias], entity.SourceBackend)
			}
		}
	}

	// Process other backends, sorted by name for deterministic conflict resolution among non-defaults
	for _, backendName := range backendNames {
		if backendName == defaultBackendName {
			continue // Already processed or doesn't exist
		}
		if entitiesFromBackend, ok := backendEntitiesMap[backendName]; ok {
			for _, entity := range entitiesFromBackend {
				foundInBackendsByAlias[entity.Alias] = append(foundInBackendsByAlias[entity.Alias], entity.SourceBackend)
				if _, exists := entitiesByAlias[entity.Alias]; !exists {
					// If not already taken by default or a previous lexically smaller non-default backend
					entitiesByAlias[entity.Alias] = entity
					// It exists, meaning it was from default or an earlier (lexically) backend.
					// The current entity is a duplicate we will ignore based on prioritization.
					// Logging of this is handled after all entities are processed.
				}
			}
		}
	}

	// Log warnings for duplicates
	for alias, foundInBackends := range foundInBackendsByAlias {
		if len(foundInBackends) > 1 {
			chosenEntity := entitiesByAlias[alias]
			var ignoredSources []string
			for _, backendName := range foundInBackends {
				if backendName != chosenEntity.SourceBackend {
					ignoredSources = append(ignoredSources, backendName)
				}
			}
			if len(ignoredSources) > 0 {
				s.ctx.Logger.Warn("Alias found in multiple backends. Prioritizing version.",
					"alias", alias,
					"chosen_from_backend", chosenEntity.SourceBackend,
					"all_found_in_backends", foundInBackends, // For more detailed debug if needed
					"ignored_backends", ignoredSources)
			}
		}
	}

	// Convert map to slice for filtering and sorting
	var uniqueEntities []model.Entity
	for _, entity := range entitiesByAlias {
		uniqueEntities = append(uniqueEntities, entity)
	}

	// Apply filter if provided
	mergedAndFilteredEntities := uniqueEntities
	if filterString != "" {
		var err error
		mergedAndFilteredEntities, err = s.FilterEntities(uniqueEntities, filterString)
		if err != nil {
			s.ctx.Logger.Warn("Error applying filter to merged entities", "filter", filterString, "error", err)
		}
	}

	// Sort final list by Alias only (SourceBackend and Title were for tie-breaking if needed, but alias is now unique)
	sort.Slice(mergedAndFilteredEntities, func(i, j int) bool {
		return mergedAndFilteredEntities[i].Alias < mergedAndFilteredEntities[j].Alias
	})

	return mergedAndFilteredEntities, backendErrors
}

// ListEntitiesFromBackend returns a list of entities from a specific backend that match the given prefix and filter.
// The list is sorted by Alias.
func (s *EntityService) ListEntitiesFromBackend(backendName string, prefix string, filterString string) ([]model.Entity, error) {
	s.ctx.Logger.Debug("Listing entities from specific backend", "backend", backendName, "prefix", prefix, "filter", filterString)

	backend, err := s.ctx.GetBackend(backendName)
	if err != nil {
		return nil, fmt.Errorf("failed to get backend '%s': %w", backendName, err)
	}

	aliases, err := backend.List(prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to list entity aliases from backend '%s' (prefix: '%s'): %w", backendName, prefix, err)
	}

	var entities []model.Entity
	for _, alias := range aliases {
		metadata, err := backend.Stat(alias)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				s.ctx.Logger.Info("Entity listed but not found on Stat, skipping.", "backend", backendName, "alias", alias)
				continue
			}
			s.ctx.Logger.Warn("Failed to get metadata for entity, skipping.", "backend", backendName, "alias", alias, "error", err)
			continue
		}

		entity := model.Entity{
			Alias:         alias,
			SourceBackend: backend.GetName(),
		}

		if metadata != nil {
			if title, ok := metadata["title"].(string); ok {
				entity.Title = title
			}
			if desc, ok := metadata["description"].(string); ok {
				entity.Description = desc
			}
			if tags, ok := metadata["tags"].([]string); ok {
				entity.Tags = tags
				sort.Strings(entity.Tags)
			}
			if cid, ok := metadata["cid"].(string); ok {
				entity.CID = cid
			}
			if pcid, ok := metadata["pcid"].(string); ok {
				entity.PCID = pcid
			}

			entity.CustomMetadata = make(map[string]interface{})
			for k, v := range metadata {
				switch k {
				case "title", "description", "tags", "cid", "pcid":
					// These are handled directly above or are internal, skip them for CustomMetadata
				default:
					entity.CustomMetadata[k] = v
				}
			}
		}
		entities = append(entities, entity)
	}

	filteredEntities := entities
	if filterString != "" {
		var filterErr error
		filteredEntities, filterErr = s.FilterEntities(entities, filterString)
		if filterErr != nil {
			s.ctx.Logger.Warn("Error applying filter to backend entities", "backend", backendName, "filter", filterString, "error", filterErr)
			return entities, fmt.Errorf("failed to apply filter: %w", filterErr)
		}
	}

	sort.Slice(filteredEntities, func(i, j int) bool {
		return filteredEntities[i].Alias < filteredEntities[j].Alias
	})

	s.ctx.Logger.Debug("Successfully listed entities from backend", "backend", backendName, "count", len(filteredEntities))
	return filteredEntities, nil
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

	var filteredEntities []model.Entity
	if f != nil {
		for _, entity := range entities {
			if f.Matches(entity) {
				filteredEntities = append(filteredEntities, entity)
			}
		}
	} else {
		filteredEntities = entities
	}

	s.ctx.Logger.Debug("Entities filtered (or all kept if no filter)", "count", len(filteredEntities))
	return filteredEntities, nil
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
func (s *EntityService) createEntityFromBackendData(alias string, backendName string, contentBytes []byte, metadata map[string]interface{}) model.Entity {
	entity := model.Entity{
		Alias:         alias,
		SourceBackend: backendName,
		// Body will be set from parsedData below
	}

	parsedData, parseErr := content.ParseG6E(contentBytes)
	if parseErr != nil {
		s.ctx.Logger.Error("Failed to parse G6E content in createEntityFromBackendData",
			"alias", alias, "backend", backendName, "error", parseErr)
		// If parsing fails, the body might remain the raw contentBytes converted to string,
		// or we could return an error/partially formed entity. For now, body will be empty if parse fails and not set otherwise.
		// Title, Description, Tags will rely on the metadata map if parsing fails.
		entity.Body = string(contentBytes) // Fallback: use raw content if parsing fails. This might be undesirable.
		// Consider if an error should propagate or if entity should have an ErrorState field.
	} else {
		entity.Title = parsedData.Title
		entity.Description = parsedData.Description
		entity.Tags = parsedData.Tags
		entity.Body = parsedData.Body // Correct: Use parsed body
		cidValue, err := parsedData.GetContentID()
		if err != nil {
			s.ctx.Logger.Warn("Failed to get ContentID from parsed data", "alias", alias, "error", err)
		} else {
			entity.CID = cidValue
		}
	}

	// Sort tags if they were populated (either from G6E or later from metadata)
	if len(entity.Tags) > 0 {
		sort.Strings(entity.Tags)
	}

	// Populate/override from metadata map (backend-provided, non-G6E source of truth for some fields or fallback)
	if metadata != nil {
		// If G6E parsing failed or didn't populate a field, metadata map can be a source.
		// Generally, G6E content should be king for T/D/T if successfully parsed.
		if title, ok := metadata["title"].(string); ok && entity.Title == "" {
			entity.Title = title
		}
		if desc, ok := metadata["description"].(string); ok && entity.Description == "" {
			entity.Description = desc
		}
		if tags, ok := metadata["tags"].([]string); ok && len(entity.Tags) == 0 {
			entity.Tags = tags
			sort.Strings(entity.Tags) // Ensure tags from metadata are also sorted
		}

		// CustomMetadata should only contain items from the backend's metadata map
		// that are not already standard G6E fields (Title, Description, Tags) or core model fields (CID, PCID, Alias, SourceBackend, Body).
		entity.CustomMetadata = make(map[string]interface{})
		for k, v := range metadata {
			isStandardField := false
			standardKeys := []string{"title", "description", "tags", "cid", "pcid", "alias", "sourceBackend", "body"}
			for _, sk := range standardKeys {
				if k == sk {
					isStandardField = true
					break
				}
			}
			if !isStandardField {
				entity.CustomMetadata[k] = v
			}
		}

		// Populate CID/PCID from metadata map if not already set from G6E parsing (or if they are primarily from backend meta)
		if cidMeta, ok := metadata["cid"].(string); ok && entity.CID == "" {
			entity.CID = cidMeta
		}
		if pcidMeta, ok := metadata["pcid"].(string); ok && entity.PCID == "" { // PCID not from G6E parse, so always from meta
			entity.PCID = pcidMeta
		}
	}

	return entity
}

// determineWriteBackend is a helper to select the appropriate writable backend.
// If backendName is provided, it's used. Otherwise, it tries entity.SourceBackend (if isOverwrite)
// then default backend, then sole available backend. Returns ErrAmbiguousBackend if needed.
func (s *EntityService) determineWriteBackend(entityAlias string, explicitBackendName string, sourceBackendName string, isOverwrite bool) (storage.Backend, error) {
	var backendToUse storage.ReadOnlyBackend
	var err error

	if explicitBackendName != "" {
		backendToUse, err = s.ctx.GetBackend(explicitBackendName)
		if err != nil {
			return nil, fmt.Errorf("failed to get specified backend '%s': %w", explicitBackendName, err)
		}
	} else if isOverwrite && sourceBackendName != "" {
		backendToUse, err = s.ctx.GetBackend(sourceBackendName)
		if err != nil {
			// This case should ideally not happen if GetEntity populated SourceBackend correctly.
			return nil, fmt.Errorf("failed to get source backend '%s' for overwrite: %w", sourceBackendName, err)
		}
	} else {
		backendToUse, err = s.ctx.GetDefaultBackend()
		if err != nil {
			// No default backend, or other error getting it. Check if single backend exists.
			if !errors.Is(err, storage.ErrNoDefaultBackend) {
				// Unexpected error from GetDefaultBackend
				return nil, fmt.Errorf("error getting default backend: %w", err)
			}

			// It is ErrNoDefaultBackend, so try to find a sole available backend.
			allBackends, _ := s.ctx.GetAllBackends() // Ignore errors here as we primarily care about count
			if len(allBackends) == 0 {
				return nil, fmt.Errorf("no backend specified, no default backend, and no backends configured")
			}
			if len(allBackends) == 1 {
				s.ctx.Logger.Debug("No explicit/default backend, using sole available backend", "alias", entityAlias)
				for _, singleBackend := range allBackends { // Get the single backend from map
					backendToUse = singleBackend
					break
				}
			} else {
				// Multiple backends available, and no default/explicit choice.
				return nil, storage.ErrAmbiguousBackend
			}
		}
	}

	if !backendToUse.IsWritable() {
		return nil, fmt.Errorf("target backend '%s' is read-only", backendToUse.GetName())
	}
	writableBackend, ok := backendToUse.(storage.Backend)
	if !ok {
		// This should not happen if IsWritable is true, but as a safeguard.
		return nil, fmt.Errorf("target backend '%s' is writable but not a full storage.Backend implementation", backendToUse.GetName())
	}
	return writableBackend, nil
}

// SaveEntity saves an entity to the specified backend.
// If the backend is read-only, an error is returned.
// If the entity already exists in the target backend, storage.ErrEntityAlreadyExists is returned.
// It returns the name of the backend used for saving, or an empty string if an error occurs.
func (s *EntityService) SaveEntity(entity model.Entity, backendName string) (string, error) {
	writableBackend, err := s.determineWriteBackend(entity.Alias, backendName, "", false) // For Save, sourceBackend is not used for selection, not an overwrite
	if err != nil {
		return "", err // Error already formatted by determineWriteBackend
	}

	// Check if entity already exists in this backend before attempting to write
	_, statErr := writableBackend.Stat(entity.Alias)
	if statErr == nil {
		return "", fmt.Errorf("cannot save entity '%s' to backend '%s': %w", entity.Alias, writableBackend.GetName(), storage.ErrEntityAlreadyExists)
	}
	if statErr != nil && !errors.Is(statErr, fs.ErrNotExist) {
		return "", fmt.Errorf("failed to stat entity '%s' in backend '%s' before save: %w", entity.Alias, writableBackend.GetName(), statErr)
	}

	// Prepare G6E content from model.Entity
	g6eContent := content.GuidanceContent{
		Title:       entity.Title,
		Description: entity.Description,
		Tags:        entity.Tags,
		Body:        entity.Body, // This is the textual body part, not the full G6E file string
		// Ensure CustomMetadata from entity is also passed if GuidanceContent supports it directly
		// or handle it separately if it needs to be in frontmatter.
		// For now, assuming CustomMetadata in model.Entity might be for other uses or needs specific mapping.
	}
	// Add other known/structured metadata from entity.CustomMetadata to g6eContent if applicable.
	// For example, if CID/PCID were stored in CustomMetadata and need to be in frontmatter.
	// However, entity.CID and entity.PCID are top-level fields, so they should be handled directly.

	// Handle CID and PCID (they might be part of frontmatter or separate)
	// If they are part of frontmatter, corecontent.GuidanceContent should handle them.
	// For now, let's assume they are separate and used in commitMsgDetails as before.

	fileBytes, err := g6eContent.ToFileContent() // This creates the full G6E string with frontmatter
	if err != nil {
		return "", fmt.Errorf("failed to serialize entity %s to G6E format: %w", entity.Alias, err)
	}

	// Prepare commit message details (map[string]string required by Backend.Write)
	commitMsg := map[string]string{
		"action": "save", // Or "create" / "update" based on context if available
		"alias":  entity.Alias,
	}

	// Add CID and PCID to commitMsg if available from the entity model
	if entity.CID != "" {
		commitMsg["cid"] = entity.CID
	}
	if entity.PCID != "" {
		commitMsg["pcid"] = entity.PCID
	}

	// Write the entity (using the fully serialized fileBytes)
	err = writableBackend.Write(entity.Alias, fileBytes, commitMsg)
	if err != nil {
		return "", fmt.Errorf("failed to write entity %s to backend %s: %w", entity.Alias, writableBackend.GetName(), err)
	}

	return writableBackend.GetName(), nil
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

// OverwriteEntity saves an entity to the specified backend, overwriting it if it already exists.
// If the backend is read-only, an error is returned.
// It returns the name of the backend used for overwriting, or an empty string if an error occurs.
func (s *EntityService) OverwriteEntity(entity model.Entity, backendName string) (string, error) {
	writableBackend, err := s.determineWriteBackend(entity.Alias, backendName, entity.SourceBackend, true) // For Overwrite, pass entity.SourceBackend and isOverwrite=true
	if err != nil {
		return "", err // Error already formatted by determineWriteBackend
	}

	// Prepare G6E content from model.Entity
	g6eContent := content.GuidanceContent{
		Title:       entity.Title,
		Description: entity.Description,
		Tags:        entity.Tags,
		Body:        entity.Body,
	}

	fileBytes, err := g6eContent.ToFileContent()
	if err != nil {
		return "", fmt.Errorf("failed to serialize entity %s to G6E format for overwrite: %w", entity.Alias, err)
	}

	commitMsg := map[string]string{
		"action": "overwrite", // Changed action for commit log
		"alias":  entity.Alias,
	}
	if entity.CID != "" {
		commitMsg["cid"] = entity.CID
	}
	if entity.PCID != "" {
		commitMsg["pcid"] = entity.PCID
	}

	err = writableBackend.Write(entity.Alias, fileBytes, commitMsg)
	if err != nil {
		return "", fmt.Errorf("failed to overwrite entity %s in backend %s: %w", entity.Alias, writableBackend.GetName(), err)
	}

	return writableBackend.GetName(), nil
}
