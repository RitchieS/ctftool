package ctfd

type CTFOpts struct {
	URL           string
	Username      string
	Password      string
	Output        string
	Overwrite     bool
	SaveConfig    bool
	SkipCTFDCheck bool
}

// NewOptions returns a new Options struct
func NewOpts() *CTFOpts {
	return &CTFOpts{}
}
