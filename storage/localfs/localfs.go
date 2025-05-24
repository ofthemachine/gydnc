package localfs

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gydnc/core/content"
	"gydnc/model"
	// "gydnc/storage" // REMOVED to break import cycle. Errors like ErrEntityNotFound will be handled by callers or via stdlib errors.
)

const g6eExt = ".g6e"

// Store implements the storage.Backend interface for local filesystem storage.
type Store struct {
	name     string
	basePath string
	// capabilitiesMap stores the capabilities of this backend instance.
	// The Capabilities() method from the interface will be used for external access.
	capabilitiesMap map[string]bool
	// ignoredFiles []string // Removed, as model.LocalFSConfig does not have IgnoredFiles
	fsys fs.FS // For testing, allow injecting a filesystem. For real use, os.DirFS(resolvedPath)
}

// NewStore creates a new Store instance for local filesystem operations.
// configDir is the directory of the main gydnc config file, used to resolve cfg.Path if it's relative.
func NewStore(cfg model.LocalFSConfig, configDir string) (*Store, error) {
	resolvedPath := cfg.Path
	if !filepath.IsAbs(resolvedPath) {
		if configDir == "" {
			return nil, fmt.Errorf("configDir is required to resolve relative path: %s", cfg.Path)
		}
		resolvedPath = filepath.Join(configDir, resolvedPath)
	}

	// Ensure the base path exists
	if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
		if err := os.MkdirAll(resolvedPath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create base directory '%s': %w", resolvedPath, err)
		}
	}
	return &Store{
		name:     "localfs", // Default name, can be overridden by SetName
		basePath: resolvedPath,
		// ignoredFiles: cfg.IgnoredFiles, // Removed
		capabilitiesMap: map[string]bool{ // Renamed field
			"listable":  true,
			"readable":  true,
			"writable":  true,
			"deletable": true,
		},
		fsys: os.DirFS(resolvedPath),
	}, nil
}

// Init initializes the local filesystem store.
// The 'name' for the store can be passed via initConfig["name"].
func (s *Store) Init(initConfig map[string]interface{}) error {
	if name, ok := initConfig["name"].(string); ok {
		s.name = name
	}
	if s.basePath == "" {
		return fmt.Errorf("basePath is not set for localfs store '%s'", s.name)
	}
	_, err := os.Stat(s.basePath)
	if os.IsNotExist(err) {
		return os.MkdirAll(s.basePath, 0755)
	}
	return err
}

// GetName returns the name of this backend store instance.
func (s *Store) GetName() string {
	return s.name
}

// GetBasePath returns the resolved, absolute base path of the store.
func (s *Store) GetBasePath() string {
	return s.basePath
}

// SetName sets the name of this backend store instance.
// This is useful if the store is created and then later associated with a named backend config.
func (s *Store) SetName(name string) {
	s.name = name
}

// Capabilities returns the capabilities of this backend.
func (s *Store) Capabilities() map[string]bool {
	// Return a copy to prevent external modification if that's a concern,
	// or return s.capabilitiesMap directly if it's considered read-only by convention.
	// For now, returning the direct map.
	if s.capabilitiesMap == nil { // Ensure it's initialized if Store was somehow created without NewStore
		s.capabilitiesMap = map[string]bool{
			"listable":  true,
			"readable":  true,
			"writable":  true,
			"deletable": true,
		}
	}
	return s.capabilitiesMap
}

// IsWritable indicates if the backend supports write operations.
func (s *Store) IsWritable() bool {
	// Check the capability map, default to true for localfs if not set.
	if val, ok := s.capabilitiesMap["writable"]; ok {
		return val
	}
	return true // Default for localfs
}

// isIgnored checks if a given filename matches any of the ignored patterns.
// Currently uses simple string equality. Could be expanded to glob patterns.
// This method was assuming s.ignoredFiles, which has been removed.
// If ignore functionality is needed, it must be re-implemented based on a proper config source.
func (s *Store) isIgnored(name string) bool {
	// for _, pattern := range s.ignoredFiles { // s.ignoredFiles is removed
	// 	if pattern == name {
	// 		return true
	// 	}
	// }
	// Example of how it might work if IgnoredFiles were part of model.LocalFSConfig and passed to Store:
	// if s.config != nil { // Assuming Store had a field like `config model.LocalFSConfig`
	// 	for _, pattern := range s.config.IgnoredFiles {
	// 		if pattern == name {
	// 			return true
	// 		}
	// 	}
	// }
	return false // Placeholder: No ignore patterns currently configured this way
}

// Read retrieves the content of a guidance entity and its parsed G6E frontmatter as metadata.
func (s *Store) Read(alias string) ([]byte, map[string]interface{}, error) {
	fileName := alias + g6eExt
	if s.isIgnored(fileName) {
		return nil, nil, fmt.Errorf("%w: entity is ignored: %s", fs.ErrNotExist, alias)
	}
	filePath := filepath.Join(s.basePath, fileName)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, fs.ErrNotExist // Standard library error
		}
		return nil, nil, err
	}

	parsedG6E, err := content.ParseG6E(data)
	if err != nil {
		// Log parsing error but still return raw content, metadata will be minimal.
		slog.Warn("Failed to parse G6E frontmatter during Read", "alias", alias, "path", filePath, "error", err)
		// Return basic metadata even if parsing fails, or an empty map.
		// For consistency, it's better to ensure metadata map is not nil.
		return data, make(map[string]interface{}), fmt.Errorf("failed to parse G6E content for %s: %w", alias, err)
	}

	metadata := map[string]interface{}{
		"title":       parsedG6E.Title,
		"description": parsedG6E.Description,
		"tags":        parsedG6E.Tags, // These are already []string from ParseG6E
		// Include other known frontmatter fields if necessary, or add them to CustomMetadata
	}
	// Add any other raw frontmatter fields to metadata if ParseG6E exposes them
	// For example, if parsedG6E.RawFrontmatter is a map[string]interface{}:
	// for k, v := range parsedG6E.RawFrontmatter {
	//  if _, exists := metadata[k]; !exists { // Avoid overwriting structured fields
	//   metadata[k] = v
	//  }
	// }

	return data, metadata, nil
}

// Write creates or updates a guidance entity.
func (s *Store) Write(alias string, data []byte, commitMsgDetails map[string]string) error {
	if !s.IsWritable() {
		return fs.ErrPermission // Standard library error for read-only or permission issues
	}
	fileName := alias + g6eExt
	if s.isIgnored(fileName) {
		return fmt.Errorf("cannot write to ignored entity: %s", alias)
	}
	filePath := filepath.Join(s.basePath, fileName)
	// Ensure the directory for the file exists if alias contains path separators
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory for entity '%s': %w", alias, err)
		}
	}
	return os.WriteFile(filePath, data, 0644)
}

// List retrieves a list of all guidance entity aliases (filenames without .g6e).
// The prefix parameter is not deeply implemented here yet for hierarchical listing;
// it currently lists all .g6e files under basePath.
func (s *Store) List(prefix string) ([]string, error) {
	var aliases []string
	// Convert basepath to use OS-specific separators for WalkDir
	searchPath := filepath.FromSlash(s.basePath)

	err := filepath.WalkDir(searchPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Log the error but try to continue if possible, unless it's a critical path error.
			slog.Warn("Error during filepath.WalkDir for List operation", "path", path, "error", err)
			// If the root searchPath itself is inaccessible, return the error.
			if path == searchPath && os.IsNotExist(err) {
				return fmt.Errorf("base path for store does not exist: %s; %w", searchPath, err)
			}
			// For other errors (e.g., permission denied on a sub-object), skip and continue.
			return nil // Continue walking if it's a non-critical error on a specific file/dir
		}

		// Process only files, not directories
		if !d.IsDir() {
			// Check if it's a .g6e file
			if strings.HasSuffix(d.Name(), ".g6e") {
				// Calculate alias relative to the basePath
				relPath, err := filepath.Rel(searchPath, path)
				if err != nil {
					slog.Warn("Could not determine relative path for List operation", "basePath", searchPath, "filePath", path, "error", err)
					return nil // Continue walking
				}
				alias := strings.TrimSuffix(filepath.ToSlash(relPath), ".g6e") // Use ToSlash for consistent alias format
				if !s.isIgnored(d.Name()) {                                    // Check if the original filename would be ignored
					// Apply prefix filter if present
					if prefix == "" || strings.HasPrefix(alias, prefix) {
						aliases = append(aliases, alias)
					}
				}
			}
		}
		return nil
	})

	if err != nil {
		// This error is from filepath.WalkDir if it was halted by a returned error.
		// Most errors within the walk function are handled to allow continuation.
		return nil, fmt.Errorf("error walking directory '%s': %w", searchPath, err)
	}
	return aliases, nil
}

// Delete removes a guidance entity file.
func (s *Store) Delete(alias string) error {
	if !s.IsWritable() { // Or check a specific "deletable" capability
		return fmt.Errorf("delete operation not supported by backend '%s': %w", s.name, fs.ErrPermission) // Use fs.ErrPermission
	}
	canDelete, ok := s.capabilitiesMap["deletable"]
	if !ok || !canDelete {
		return fmt.Errorf("%w: delete operation not supported by backend '%s'", fs.ErrPermission, s.name)
	}
	fileName := alias + g6eExt
	if s.isIgnored(fileName) {
		return fmt.Errorf("cannot delete ignored entity: %s", alias)
	}
	filePath := filepath.Join(s.basePath, fileName)
	err := os.Remove(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fs.ErrNotExist // Standard library error
		}
		return err
	}
	return nil
}

// Stat retrieves metadata about a guidance entity, including parsed G6E frontmatter.
func (s *Store) Stat(alias string) (map[string]interface{}, error) {
	fileName := alias + g6eExt
	if s.isIgnored(fileName) {
		return nil, fmt.Errorf("%w: entity is ignored: %s", fs.ErrNotExist, alias)
	}
	filePath := filepath.Join(s.basePath, fileName)

	// Read file content to parse frontmatter
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fs.ErrNotExist
		}
		return nil, fmt.Errorf("failed to read file for Stat %s: %w", alias, err)
	}

	parsedG6E, err := content.ParseG6E(data)
	if err != nil {
		// Log parsing error but proceed with basic file info if G6E parsing fails.
		slog.Warn("Failed to parse G6E frontmatter during Stat", "alias", alias, "path", filePath, "error", err)
		// Fallback to basic file info if parsing fails
		fileInfo, statErr := os.Stat(filePath)
		if statErr != nil { // This shouldn't happen if ReadFile succeeded, but good to check.
			return nil, fmt.Errorf("failed to stat file after G6E parse error for %s: %w", alias, statErr)
		}
		return map[string]interface{}{
			"name":     fileInfo.Name(),
			"size":     fileInfo.Size(),
			"mod_time": fileInfo.ModTime(),
			// Indicate parsing failure or incomplete metadata
			"g6e_parse_error": err.Error(),
		}, nil // Return basic info with error, or just the error: fmt.Errorf("failed to parse G6E content for Stat %s: %w", alias, err)
	}

	// Successfully parsed, return rich metadata
	metadata := map[string]interface{}{
		"title":       parsedG6E.Title,
		"description": parsedG6E.Description,
		"tags":        parsedG6E.Tags,          // These are already []string from ParseG6E
		"name":        filepath.Base(filePath), // Keep basic file info too
		// "size": // Size might be misleading if we only care about frontmatter for Stat.
		// "mod_time": // ModTime might still be relevant.
	}
	// If ParseG6E provided other frontmatter fields in a map, merge them here.
	// e.g., if parsedG6E.OtherFrontmatter exists:
	// for k, v := range parsedG6E.OtherFrontmatter {
	//  if _, exists := metadata[k]; !exists {
	//   metadata[k] = v
	//  }
	// }
	return metadata, nil
}
