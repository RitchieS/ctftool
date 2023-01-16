package scraper

import (
	"crypto/tls"
	"net/http"
)

// roundTripper is a struct that wraps an http.RoundTripper and adds a custom user-agent header
// to the requests it sends.
type roundTripper struct {
	// tripper is the underlying http.RoundTripper that will handle the actual request.
	tripper http.RoundTripper
	// userAgent is the custom user-agent header value that will be added to the requests.
	userAgent string
}

// NewTransport returns a new http.RoundTripper that wraps the provided http.RoundTripper
// and sets the TLSClientConfig to the value returned by getTLSConfig().
// It also sets the userAgent to a specific value.
//
//	tripper := &http.Transport{}
//	rt := NewTransport(tripper)
//	req, _ := http.NewRequest("GET", "https://example.com", nil)
//	rt.RoundTrip(req)
func NewTransport(tripper http.RoundTripper) http.RoundTripper {
	// Check if the provided tripper is an *http.Transport, if so set its TLSClientConfig to the one returned by getTLSConfig().
	if transport, ok := tripper.(*http.Transport); ok {
		transport.TLSClientConfig = getTLSConfig()
	}

	// Return a new roundTripper, with the tripper field set to the provided tripper, and the userAgent field set to a default value.
	return &roundTripper{
		tripper:   tripper,
		userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36	",
	}
}

// RoundTrip is an implementation of the http.RoundTripper interface.
// It checks the request headers for "Accept-Language", "Accept", and "User-Agent"
// and sets default values if they are not set.
// It then calls the RoundTrip method of the embedded http.RoundTripper.
// If the embedded tripper is nil, it uses a new http.Transport with a set TLSClientConfig.
//
//	rt := &roundTripper{}
//	req, _ := http.NewRequest("GET", "https://example.com", nil)
//	rt.RoundTrip(req)
func (b *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Check if the request's Accept-Language header is set. If not, set it to "en-US,en;q=0.5".
	if req.Header.Get("Accept-Language") == "" {
		req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	}

	// Check if the request's Accept header is set. If not, set it to "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8".
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	}

	// Check if the request's User-Agent header is set. If not, set it to the value stored in b.userAgent.
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", b.userAgent)
	}

	// Check if the tripper field of b is nil. If it is, create a new Transport with the default TLSClientConfig and use it to perform the RoundTrip.
	if b.tripper == nil {
		return (&http.Transport{
			TLSClientConfig: getTLSConfig(),
		}).RoundTrip(req)
	}

	// If b.tripper is not nil, use it to perform the RoundTrip.
	return b.tripper.RoundTrip(req)
}

// getTLSConfig returns a new tls.Config with preferred cipher suites and curve preferences set.
func getTLSConfig() *tls.Config {
	// Returns a tls.Config with the following fields set:
	// 1. PreferServerCipherSuites set to false
	// 2. CurvePreferences set to an array containing tls.CurveP256, tls.CurveP384, tls.CurveP521 and tls.X25519.
	return &tls.Config{
		PreferServerCipherSuites: false,
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
			tls.CurveP384,
			tls.CurveP521,
			tls.X25519,
		},
	}
}
