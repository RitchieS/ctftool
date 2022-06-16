package lib

type Options struct {
	ConfigFile  string
	Debug       bool
	DebugFormat string
	RateLimit   int  // rate limit per second
	Interactive bool // Interactive is a flag to enable interactive mode
}

// NewOptions returns a new Options struct
func NewOptions() *Options {
	return &Options{}
}
