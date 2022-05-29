package lib

type Options struct {
	ConfigFile  string
	Debug       bool
	DebugFormat string
}

// NewOptions returns a new Options struct
func NewOptions() *Options {
	return &Options{}
}
