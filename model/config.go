package model

// IMPORTANT: Configuration Backward Compatibility Notice
//
// The YAML config format (config.yml) is considered "extend only" for backward compatibility.
// New fields may be added, but existing fields must not be renamed or removed.
// The structure must remain compatible with older versions of the application.

// LocalFSConfig defines settings specific to the local filesystem backend.
// For the MVP, Git integration settings are omitted and considered a future enhancement.
//
// @stable: This structure must not be renamed or have fields removed
type LocalFSConfig struct {
	Path string `yaml:"path" json:"path"` // @stable: Required field
}

// StorageConfig defines the configuration for a storage backend.
// Only one backend type (e.g., LocalFS) should be configured at a time per named backend instance.
//
// @stable: This structure must not be renamed or have fields removed
// @extendable: New backend types can be added
type StorageConfig struct {
	Type    string         `yaml:"type" json:"type"`                 // e.g., "localfs" @stable: Required field
	LocalFS *LocalFSConfig `yaml:"localfs,omitempty" json:"localfs"` // Pointer to allow omitempty @stable
	// Other backend types like S3Config, DBConfig etc. would go here
}

// Config defines the structure of the gydnc.conf file.
// It supports multiple named storage backends.
//
// @stable: This structure must not be renamed or have fields removed
// @extendable: New fields can be added
type Config struct {
	DefaultBackend  string                    `yaml:"default_backend" json:"default_backend"`
	StorageBackends map[string]*StorageConfig `yaml:"storage_backends" json:"storage_backends"`
	// Future global settings can go here, e.g., relating to canonicalization or hashing defaults
	// Canonicalization struct {
	// 	 HashAlgorithm string   `yaml:"hash_algorithm"`
	// 	 IncludeFields []string `yaml:"include_fields"`
	// } `yaml:"canonicalization"`
}
