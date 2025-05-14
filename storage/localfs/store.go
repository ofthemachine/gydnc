package localfs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gydnc/config"
	"gydnc/core/content"
	// We need to import the storage package to refer to storage.Backend if this store is to implement it.
	// However, to avoid import cycles if storage.Backend refers to specific store types (which it shouldn't),
	// it's better if storage.Backend is a pure interface definition.
	// Assuming storage.Backend is defined in a way that this is fine.
)

const g6eExtension = ".g6e"

// Store implements the storage.Backend interface for local filesystem.
type Store struct {
	config.LocalFSConfig        // Embed LocalFSConfig for easy access to Path
	basePath             string // Absolute path to the guidance directory
}

// NewStore creates a new instance of a local filesystem store.
// It expects the specific LocalFSConfig for this backend instance.
func NewStore(cfg config.LocalFSConfig) (*Store, error) {
	if cfg.Path == "" {
		return nil, fmt.Errorf("localfs store path cannot be empty")
	}
	absPath, err := filepath.Abs(cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for localfs store: %w", err)
	}
	return &Store{LocalFSConfig: cfg, basePath: absPath},
		nil
}

// Init initializes the local filesystem backend.
// For localfs, this primarily means ensuring the base path exists.
func (s *Store) Init(_ map[string]interface{}) error { // config map not used for this simple Init
	if s.basePath == "" {
		return fmt.Errorf("localfs store base path is not configured")
	}
	if _, err := os.Stat(s.basePath); os.IsNotExist(err) {
		// Create the directory if it doesn't exist.
		// This is more for `gydnc init`'s responsibility, but good to have robustness here.
		if err := os.MkdirAll(s.basePath, 0750); err != nil {
			return fmt.Errorf("failed to create localfs base directory %s: %w", s.basePath, err)
		}
		fmt.Printf("Created localfs guidance directory: %s\n", s.basePath)
	} else if err != nil {
		return fmt.Errorf("failed to check localfs base directory %s: %w", s.basePath, err)
	}
	return nil
}

// resolveAlias converts an alias to an absolute file path.
func (s *Store) resolveAlias(alias string) string {
	return filepath.Join(s.basePath, alias+g6eExtension)
}

// Read retrieves the raw content of a guidance entity by its alias.
func (s *Store) Read(alias string) ([]byte, map[string]interface{}, error) {
	filePath := s.resolveAlias(alias)

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, fmt.Errorf("guidance alias '%s' not found at %s", alias, filePath)
		}
		return nil, nil, fmt.Errorf("failed to read guidance file %s: %w", filePath, err)
	}

	gc, err := content.ParseG6E(fileContent)
	if err != nil {
		// As per plan: direct read of a malformed file should error.
		return nil, nil, fmt.Errorf("failed to parse guidance file %s (alias: %s): %w", filePath, alias, err)
	}

	// For MVP, metadata can be a simple map derived from GuidanceContent.
	// In a fuller implementation, this might be more structured or directly return *gc.
	metadata := map[string]interface{}{
		"title":       gc.Title,
		"description": gc.Description,
		"tags":        gc.Tags,
		// Body is not typically part of "metadata" but ParseG6E separates it.
	}

	return fileContent, metadata, nil // Returning full fileContent as per backend.go comment for now.
}

// Write creates or updates a guidance entity.
// For Phase 1, this is a stub to satisfy the interface.
func (s *Store) Write(alias string, data []byte, commitMsgDetails map[string]string) error {
	filePath := s.resolveAlias(alias)

	// Phase 2 will handle Git integration here based on `commitMsgDetails`.
	// For now, just write the file.
	// Ensure directory for the alias exists if it's nested, e.g. "core/new_rule"
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if mkErr := os.MkdirAll(dir, 0750); mkErr != nil {
			return fmt.Errorf("failed to create directory %s for alias %s: %w", dir, alias, mkErr)
		}
	}

	err := os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write guidance file %s (alias: %s): %w", filePath, alias, err)
	}

	// Basic Git operation (if repo exists and git is in PATH) - best effort for MVP
	// This is a simplified placeholder. Real Git logic would be more robust.
	// In a real scenario, use go-git or similar, or ensure git commands are safe.
	// For MVP, this is highly simplified and might not be robust enough for all environments.
	// fmt.Printf("Attempting Git operations for %s...\n", alias) // Debug
	// cmdAdd := exec.Command("git", "-C", s.basePath, "add", filePath)
	// if err := cmdAdd.Run(); err != nil { fmt.Fprintf(os.Stderr, "Warning: git add failed for %s: %v\n", alias, err) }
	// commitMsg := fmt.Sprintf("Update guidance: %s", alias)
	// if reason, ok := commitMsgDetails["reason"]; ok && reason != "" {
	// 	commitMsg = fmt.Sprintf("Update guidance: %s\n\nReason: %s", alias, reason)
	// } else if opType, ok := commitMsgDetails["operationType"]; ok && opType == "create" {
	// 	commitMsg = fmt.Sprintf("Create guidance: %s", alias)
	// }
	// cmdCommit := exec.Command("git", "-C", s.basePath, "commit", "-m", commitMsg)
	// if err := cmdCommit.Run(); err != nil { fmt.Fprintf(os.Stderr, "Warning: git commit failed for %s: %v\n", alias, err) }

	return nil
}

// List returns a list of aliases available in the backend, potentially filtered.
func (s *Store) List(filterQuery string) (aliases []string, err error) {
	var foundAliases []string

	// Basic tag filter parsing for MVP: "tags:value", "tags:value1 AND tags:value2", "NOT tags:value"
	// This is a placeholder for a more robust query parser.
	requiredTags := make(map[string]bool)
	negatedTags := make(map[string]bool)

	if filterQuery != "" {
		// Example parsing - extremely simplified for MVP
		parts := strings.Fields(filterQuery)
		for i := 0; i < len(parts); i++ {
			part := parts[i]
			if strings.HasPrefix(part, "tags:") {
				requiredTags[strings.TrimPrefix(part, "tags:")] = true
			} else if strings.HasPrefix(part, "NOT") && (i+1 < len(parts)) && strings.HasPrefix(parts[i+1], "tags:") {
				negatedTags[strings.TrimPrefix(parts[i+1], "tags:")] = true
				i++ // consume next part
			}
			// "AND" is implicit, other operators not supported in this MVP parse
		}
	}

	err = filepath.Walk(s.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), g6eExtension) {
			relPath, err := filepath.Rel(s.basePath, path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not get relative path for %s: %v\n", path, err)
				return nil // Skip this file
			}
			alias := strings.TrimSuffix(relPath, g6eExtension)

			// If filtering, read and parse frontmatter
			if len(requiredTags) > 0 || len(negatedTags) > 0 {
				fileContent, readErr := os.ReadFile(path)
				if readErr != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to read file %s for filtering: %v\n", path, readErr)
					return nil // Skip
				}
				gc, parseErr := content.ParseG6E(fileContent)
				if parseErr != nil {
					// As per plan: list skips malformed and warns.
					fmt.Fprintf(os.Stderr, "Warning: skipping malformed guidance file %s (alias: %s): %v\n", path, alias, parseErr)
					return nil // Skip
				}

				// Apply filters
				match := true
				for tag := range requiredTags {
					found := false
					for _, t := range gc.Tags {
						if t == tag {
							found = true
							break
						}
					}
					if !found {
						match = false
						break
					}
				}
				if match {
					for tag := range negatedTags {
						for _, t := range gc.Tags {
							if t == tag {
								match = false
								break
							}
						}
						if !match {
							break
						}
					}
				}

				if !match {
					return nil // Skip if filters don't match
				}
			}
			foundAliases = append(foundAliases, alias)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking guidance directory %s: %w", s.basePath, err)
	}
	return foundAliases, nil
}

// GetName returns the name of the backend.
func (s *Store) GetName() string {
	return "localfs"
}
