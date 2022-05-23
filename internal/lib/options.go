package lib

type Credentials struct {
	Username string
	Password string
	Email    string
}

type Options struct {
	ConfigFile  string
	Debug       bool
	DebugFormat string
	Verbose     bool
	Credential  Credentials
}

// NewOptions returns a new Options struct
func NewOptions() *Options {
	return &Options{}
}
