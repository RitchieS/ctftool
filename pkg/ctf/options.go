package ctf

type CTFOpts struct {
	URL         string
	Username    string
	Password    string
	Output      string
	Overwrite   bool
	SaveConfig  bool
	MaxFileSize int64
}

// NewOptions returns a new Options struct
func NewOpts() *CTFOpts {
	return &CTFOpts{}
}
