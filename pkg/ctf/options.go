package ctf

import "fmt"

type CTFOpts struct {
	URL        string
	Username   string
	Password   string
	Output     string
	Overwrite  bool
	SaveConfig bool
}

// NewOptions returns a new Options struct
func NewOpts() *CTFOpts {
	return &CTFOpts{}
}

// Error reports an error and the operation and URL that caused it.
type Error struct {
	Op  string
	URL string
	Err error
}

func (e *Error) Unwrap() error { return e.Err }
func (e *Error) Error() string { return fmt.Sprintf("%s %q: %s", e.Op, e.URL, e.Err) }
