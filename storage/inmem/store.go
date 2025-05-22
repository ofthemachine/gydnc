package inmem

import (
	"io/fs"
	"strings"
	"sync"
)

// Store implements the storage.ReadOnlyBackend interface for in-memory storage.
// This is useful for testing and demonstration purposes.
type Store struct {
	name     string
	entities map[string]entity
	mu       sync.RWMutex
}

type entity struct {
	content  []byte
	metadata map[string]interface{}
}

// NewStore creates a new in-memory backend.
func NewStore(name string) *Store {
	return &Store{
		name:     name,
		entities: make(map[string]entity),
	}
}

// Init initializes the in-memory store.
func (s *Store) Init(metadata map[string]interface{}) error {
	if name, ok := metadata["name"].(string); ok && name != "" {
		s.name = name
	}
	return nil
}

// GetName returns the name of the backend instance.
func (s *Store) GetName() string {
	if s.name == "" {
		return "inmem"
	}
	return s.name
}

// Read retrieves the content of a guidance entity by its ID (alias).
func (s *Store) Read(id string) ([]byte, map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	e, ok := s.entities[id]
	if !ok {
		return nil, nil, fs.ErrNotExist
	}

	// Clone metadata to prevent mutation
	metadata := make(map[string]interface{}, len(e.metadata))
	for k, v := range e.metadata {
		metadata[k] = v
	}

	// Add backend info to metadata
	metadata["backend_name"] = s.GetName()
	metadata["path"] = id

	return e.content, metadata, nil
}

// Stat returns metadata about a guidance entity by its ID (alias).
func (s *Store) Stat(id string) (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	e, ok := s.entities[id]
	if !ok {
		return nil, fs.ErrNotExist
	}

	// Clone metadata to prevent mutation
	metadata := make(map[string]interface{}, len(e.metadata))
	for k, v := range e.metadata {
		metadata[k] = v
	}

	// Add backend info to metadata
	metadata["backend_name"] = s.GetName()
	metadata["path"] = id

	return metadata, nil
}

// List retrieves a list of guidance entity IDs (aliases) based on a prefix.
func (s *Store) List(prefix string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var ids []string
	for id := range s.entities {
		if strings.HasPrefix(id, prefix) {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

// IsWritable returns true if this backend supports write operations.
func (s *Store) IsWritable() bool {
	return false // This is a read-only backend
}

// Capabilities returns a map of capability names to boolean values.
func (s *Store) Capabilities() map[string]bool {
	return map[string]bool{
		"write":  false,
		"delete": false,
		"list":   true,
		"read":   true,
		"stat":   true,
	}
}

// LoadEntities loads a set of entities into the store for testing.
// This is not part of the backend interface but is useful for testing.
func (s *Store) LoadEntities(entities map[string][]byte, metadata map[string]map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, content := range entities {
		meta := map[string]interface{}{}
		if m, ok := metadata[id]; ok {
			for k, v := range m {
				meta[k] = v
			}
		}
		s.entities[id] = entity{
			content:  content,
			metadata: meta,
		}
	}
}
