package lib

type Options struct {
	ConfigFile  string
	Debug       bool
	DebugFormat string
	Verbose     bool
}

// NewOptions returns a new Options struct
func NewOptions() *Options {
	return &Options{}
}
