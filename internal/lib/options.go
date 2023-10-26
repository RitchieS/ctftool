package lib

type Options struct {
	ConfigFile  string
	Debug       bool
	DebugFormat string
	RateLimit   int // rate limit per second
}

// NewOptions returns a new Options struct
func NewOptions() *Options {
	return &Options{}
}
