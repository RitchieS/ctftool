package ctfd

import "time"

type CTFOpts struct {
	URL           string
	Username      string
	Password      string
	Output        string
	Overwrite     bool
	SaveConfig    bool
	SkipCTFDCheck bool
	Watch         bool
	WatchInterval time.Duration
}

// NewOptions returns a new Options struct
func NewOpts() *CTFOpts {
	return &CTFOpts{}
}
