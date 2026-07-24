package stern

import "time"

// Options holds the configuration for a single stern invocation. Fields are
// populated from the cobra-bound flags before Run is invoked.
type Options struct {
	Container    string
	Selector     string
	Exclude      []string
	Include      []string
	Highlight    []string
	Since        time.Duration
	Tail         int64
	Timestamps   string
	Color        string
	Output       string
	NoFollow     bool
	AllNamespace bool
}
