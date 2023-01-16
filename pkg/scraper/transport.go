package scraper

import (
	"crypto/tls"
	"net/http"
)

type roundTripper struct {
	tripper   http.RoundTripper
	userAgent string
}

func NewTransport(tripper http.RoundTripper) http.RoundTripper {
	if transport, ok := tripper.(*http.Transport); ok {
		transport.TLSClientConfig = getTLSConfig()
	}

	return &roundTripper{
		tripper:   tripper,
		userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.102 Safari/537.36",
	}
}

func (b *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Header.Get("Accept-Language") == "" {
		req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	}

	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	}

	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", b.userAgent)
	}

	if b.tripper == nil {
		return (&http.Transport{
			TLSClientConfig: getTLSConfig(),
		}).RoundTrip(req)
	}

	return b.tripper.RoundTrip(req)
}

func getTLSConfig() *tls.Config {
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
