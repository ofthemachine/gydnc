package logging

import (
	"log/slog"
	"os"

	"gydnc/internal/config" // To use the Set/Get verbosity/quiet funcs
)

// SetupLogger initializes the global slog logger based on verbosity and quiet flags.
// It also updates the global verbosity/quiet state in the config package.
func SetupLogger(verbosityCount int, quiet bool) {
	level := slog.LevelWarn // Default level

	if quiet {
		level = slog.LevelError + 100 // Effectively silence most slog output
		config.SetQuiet(true)
		config.SetVerbosity(0)
	} else {
		config.SetQuiet(false)
		config.SetVerbosity(verbosityCount)
		switch verbosityCount {
		case 0: // Default, no -v flags
			level = slog.LevelWarn
		case 1: // -v
			level = slog.LevelInfo
		default: // -vv or more
			level = slog.LevelDebug
		}
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)

	slog.Debug("Logger initialized", "level", level.String(), "quiet_mode", quiet, "verbosity", verbosityCount)
}
