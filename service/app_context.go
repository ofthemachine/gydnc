package service

import (
	"log/slog"

	"gydnc/model"
	"gydnc/storage"
)

// AppContext holds application-wide dependencies and configuration.
// It is passed to service functions rather than relying on global state.
type AppContext struct {
	Config *model.Config
	Logger *slog.Logger
}

// NewAppContext creates a new AppContext with the provided configuration and logger.
// If logger is nil, a default logger will be created.
func NewAppContext(cfg *model.Config, logger *slog.Logger) *AppContext {
	if logger == nil {
		logger = slog.Default()
	}

	return &AppContext{
		Config: cfg,
		Logger: logger,
	}
}

// GetBackend returns the backend specified by name.
// If the backend does not exist in the registry, it will attempt to initialize it
// from the configuration.
func (ctx *AppContext) GetBackend(name string) (storage.ReadOnlyBackend, error) {
	// Check if backend is already registered
	backend := storage.GetBackend(name)
	if backend != nil {
		return backend, nil
	}

	// If not, try to initialize it from config
	backendCfg, ok := ctx.Config.StorageBackends[name]
	if !ok {
		return nil, storage.ErrBackendNotFound
	}

	return storage.NewBackendFromConfig(name, backendCfg)
}

// GetDefaultBackend returns the default backend as specified in the configuration.
func (ctx *AppContext) GetDefaultBackend() (storage.ReadOnlyBackend, error) {
	defaultBackendName := ctx.Config.DefaultBackend
	if defaultBackendName == "" {
		return nil, storage.ErrNoDefaultBackend
	}

	return ctx.GetBackend(defaultBackendName)
}

// GetAllBackends returns all configured backends.
// It ensures all backends in the config are initialized.
func (ctx *AppContext) GetAllBackends() (map[string]storage.ReadOnlyBackend, map[string]error) {
	backends := make(map[string]storage.ReadOnlyBackend)
	errors := make(map[string]error)

	// Initialize all backends if not already done
	for name := range ctx.Config.StorageBackends {
		backend, err := ctx.GetBackend(name)
		if err != nil {
			errors[name] = err
		} else {
			backends[name] = backend
		}
	}

	return backends, errors
}
