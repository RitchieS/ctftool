package lib

type Options struct {
	ConfigFile  string
	Debug       bool
	DebugFormat string
	RateLimit   int   // rate limit per second
	MaxFileSize int64 // max file size in mb
}

// NewOptions returns a new Options struct
func NewOptions() *Options {
	return &Options{}
}
