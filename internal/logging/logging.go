package logging

import (
	"log/slog"
	"os"

	"gydnc/internal/config" // To use the Set/Get verbosity/quiet funcs
)

// SetupLogger initializes the global slog logger based on verbosity and quiet flags.
// It also updates the global verbosity/quiet state in the config package.
func SetupLogger(verbosityCount int, quiet bool) {
	level := slog.LevelInfo // Default level changed to INFO

	if quiet {
		level = slog.LevelError + 100 // Effectively silence most slog output
		config.SetQuiet(true)
		config.SetVerbosity(0)
	} else {
		config.SetQuiet(false)
		config.SetVerbosity(verbosityCount)
		switch verbosityCount {
		case 0: // Default, no -v flags
			level = slog.LevelInfo // Default to INFO level
		case 1: // -v
			level = slog.LevelInfo // -v still INFO (or could be DEBUG if preferred for -v)
		default: // -vv or more
			level = slog.LevelDebug
		}
	}

	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{} // Remove time attribute always
			}

			return a
		},
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, opts))
	slog.SetDefault(logger)

	slog.Debug("Logger initialized", "level", level.String(), "quiet_mode", quiet, "verbosity", verbosityCount)
}
