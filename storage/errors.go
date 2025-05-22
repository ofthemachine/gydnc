package storage

import (
	"errors"
)

// Common errors for storage backends
var (
	// ErrBackendNotFound is returned when a requested backend is not found in the registry
	ErrBackendNotFound = errors.New("backend not found")

	// ErrNoDefaultBackend is returned when no default backend is configured
	ErrNoDefaultBackend = errors.New("no default backend configured")

	// ErrEntityNotFound is returned when a requested entity is not found in a backend
	ErrEntityNotFound = errors.New("entity not found")

	// ErrReadOnlyBackend is returned when a write operation is attempted on a read-only backend
	ErrReadOnlyBackend = errors.New("backend is read-only")

	// ErrUnsupportedOperation is returned when an operation is not supported by a backend
	ErrUnsupportedOperation = errors.New("operation not supported by this backend")
)
