package localfs

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gydnc/config" // For config.LocalFSConfig
	// "gydnc/storage" // Not imported to avoid circular dependency if storage imports localfs indirectly
)

const g6eExt = ".g6e"

// Store implements the storage.Backend interface for local filesystem.
type Store struct {
	name     string
	basePath string
	fsys     fs.FS // For testing, allow injecting a filesystem
}

// NewStore creates a new local filesystem backend.
// The provided LocalFSConfig contains the root path for this store.
func NewStore(cfg config.LocalFSConfig) (*Store, error) {
	if cfg.Path == "" {
		return nil, fmt.Errorf("localfs path cannot be empty")
	}
	absPath, err := filepath.Abs(cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for localfs store: %w", err)
	}

	// Ensure the base directory exists
	// slog.Debug("Ensuring base directory for localfs store", "path", absPath)
	if err := os.MkdirAll(absPath, 0750); err != nil { // 0750: rwxr-x---
		return nil, fmt.Errorf("failed to create base directory '%s' for localfs store: %w", absPath, err)
	}

	return &Store{
		basePath: absPath,
		// Name can be set via Init or a dedicated setter if needed, or passed in NewStore
	}, nil
}

// Init initializes the localfs store. The name is derived from metadata if provided.
// For localfs, metadata is not strictly required for basic operation beyond setting the name.
func (s *Store) Init(metadata map[string]interface{}) error {
	// slog.Debug("Initializing localfs store", "basePath", s.basePath)
	if s.basePath == "" {
		return fmt.Errorf("basePath cannot be empty for localfs store, ensure NewStore was called with valid config")
	}
	// Check if basePath exists and is a directory
	info, err := os.Stat(s.basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("basePath '%s' does not exist for localfs store", s.basePath)
		}
		return fmt.Errorf("failed to stat basePath '%s': %w", s.basePath, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("basePath '%s' is not a directory", s.basePath)
	}

	if name, ok := metadata["name"].(string); ok {
		s.name = name
	}
	// The empty else block that was here has been removed to clear SA9003.
	// Fallback name logic (if any) would go here or GetName() would handle it.

	s.fsys = os.DirFS(s.basePath) // Use the real filesystem
	return nil
}

// GetName returns the name of the backend instance.
func (s *Store) GetName() string {
	if s.name == "" {
		// Attempt to derive a default name from the base path if not set
		return filepath.Base(s.basePath)
	}
	return s.name
}

// GetBasePath returns the root path of this localfs store.
func (s *Store) GetBasePath() string {
	return s.basePath
}

// resolvePath converts an ID (alias) to an absolute path within the store.
// It ensures the path is still within the store's basePath.
func (s *Store) resolvePath(id string) (string, error) {
	if strings.Contains(id, "..") { // Basic path traversal protection
		return "", fmt.Errorf("invalid id: '%s' contains '..'", id)
	}
	if filepath.IsAbs(id) {
		return "", fmt.Errorf("invalid id: '%s' must be a relative path/alias", id)
	}
	// Ensure it has the .g6e extension. This was previously handled in cmd,
	// but makes sense for the storage backend to enforce or manage its own file types.
	// However, the interface uses 'id' generically. For now, assume id is alias WITHOUT extension.
	fileName := id + g6eExt
	absPath := filepath.Join(s.basePath, fileName)

	// Security check: Ensure the resolved path is still within the basePath
	rel, err := filepath.Rel(s.basePath, absPath)
	if err != nil {
		return "", fmt.Errorf("could not make path relative: %w", err)
	}
	if strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("resolved path '%s' is outside of store basePath '%s'", absPath, s.basePath)
	}

	return absPath, nil
}

// Read retrieves the content of a guidance entity by its ID (alias).
// For localfs, id is the filename without extension.
func (s *Store) Read(id string) ([]byte, map[string]interface{}, error) {
	filePath, err := s.resolvePath(id)
	if err != nil {
		return nil, nil, fmt.Errorf("read: invalid id or path resolution for '%s': %w", id, err)
	}

	// slog.Debug("Reading file", "path", filePath)
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, fs.ErrNotExist // Use fs.ErrNotExist for consistency
		}
		return nil, nil, fmt.Errorf("failed to read file '%s': %w", filePath, err)
	}

	// For localfs, "path" in metadata is the id itself (relative path/alias)
	// "backend_name" should be s.GetName()
	metadata := map[string]interface{}{
		"path":         id,
		"backend_name": s.GetName(),
		"full_path":    filePath, // Could be useful for debugging or context
	}
	return content, metadata, nil
}

// Write saves the content of a guidance entity by its ID (alias).
// For localfs, id is the filename without extension.
// Metadata changed to map[string]string to align with compiler's view of storage.Backend interface.
func (s *Store) Write(id string, content []byte, metadata map[string]string) error {
	filePath, err := s.resolvePath(id)
	if err != nil {
		return fmt.Errorf("write: invalid id or path resolution for '%s': %w", id, err)
	}

	// Ensure the directory for the file exists (if id includes subdirs, e.g., "group/myalias")
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create directory '%s' for writing: %w", dir, err)
	}

	// slog.Debug("Writing file", "path", filePath, "metadata", metadata) // Metadata is map[string]string
	err = os.WriteFile(filePath, content, 0640) // rw-r-----
	if err != nil {
		return fmt.Errorf("failed to write file '%s': %w", filePath, err)
	}
	return nil
}

// Delete removes a guidance entity by its ID (alias).
func (s *Store) Delete(id string) error {
	filePath, err := s.resolvePath(id)
	if err != nil {
		return fmt.Errorf("delete: invalid id or path resolution for '%s': %w", id, err)
	}

	// slog.Debug("Deleting file", "path", filePath)
	err = os.Remove(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fs.ErrNotExist // Consistent error for not found
		}
		return fmt.Errorf("failed to delete file '%s': %w", filePath, err)
	}
	return nil
}

// Stat returns metadata about a guidance entity by its ID (alias).
func (s *Store) Stat(id string) (map[string]interface{}, error) {
	filePath, err := s.resolvePath(id)
	if err != nil {
		return nil, fmt.Errorf("stat: invalid id or path resolution for '%s': %w", id, err)
	}

	// slog.Debug("Stating file", "path", filePath)
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fs.ErrNotExist
		}
		return nil, fmt.Errorf("failed to stat file '%s': %w", filePath, err)
	}

	metadata := map[string]interface{}{
		"id":           id, // The alias
		"path":         id, // Relative path is the ID for localfs
		"full_path":    filePath,
		"size":         fileInfo.Size(),
		"mod_time":     fileInfo.ModTime(),
		"is_dir":       fileInfo.IsDir(), // Should always be false for .g6e files
		"backend_name": s.GetName(),
	}
	return metadata, nil
}

// List retrieves a list of guidance entity IDs (aliases) based on a prefix.
// If prefix is empty, it lists all entities in the backend.
// Returns a list of IDs (strings) and conforms to the updated storage.Backend interface.
func (s *Store) List(prefix string) ([]string, error) {
	// slog.Debug("Listing files in store", "basePath", s.basePath, "prefix", prefix)
	var results []string

	err := filepath.WalkDir(s.basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // Propagate errors
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), g6eExt) {
			relPath, relErr := filepath.Rel(s.basePath, path)
			if relErr != nil {
				return relErr
			}

			id := strings.TrimSuffix(relPath, g6eExt)

			if prefix == "" || strings.HasPrefix(id, prefix) {
				results = append(results, id)
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list entity IDs in '%s': %w", s.basePath, err)
	}

	return results, nil
}
