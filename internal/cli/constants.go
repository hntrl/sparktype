package cli

import "time"

const (
	// DefaultConfigFileName is the default configuration file name looked for in the current directory.
	DefaultConfigFileName = "typegen.jsonc"

	// WatchPollInterval is how often watch mode checks for file changes.
	// Using polling because fsnotify has cross-platform inconsistencies.
	WatchPollInterval = 500 * time.Millisecond
)
