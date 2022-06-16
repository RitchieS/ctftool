package ctf

import (
	"fmt"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/ritchies/ctftool/internal/lib"
	"golang.org/x/net/publicsuffix"
)

type Client struct {
	Client      *http.Client
	BaseURL     *url.URL
	Creds       *Credentials
	MaxFileSize int64
}

type Credentials struct {
	Username string
	Password string
}

// NewClient constructs a new Client. If transport is nil, a default transport is used.
func NewClient(transport http.RoundTripper) *Client {
	cookieJar, _ := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})

	// Set long timeout to avoid timeouts because CTFd is slow
	if transport == nil {
		transport = &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   time.Duration(3 * time.Minute),
				KeepAlive: time.Duration(15 * time.Second),
				DualStack: true,
			}).DialContext,
			MaxConnsPerHost:       0,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   100,
			Proxy:                 http.ProxyFromEnvironment,
			ExpectContinueTimeout: time.Duration(1 * time.Second),
			TLSHandshakeTimeout:   time.Duration(10 * time.Second),
			IdleConnTimeout:       time.Duration(90 * time.Second),
			ResponseHeaderTimeout: time.Duration(2 * time.Minute),
		}
	}

	return &Client{
		Client: &http.Client{
			Transport: lib.Bypass(transport),
			Jar:       cookieJar,
		},
		Creds:       &Credentials{},
		MaxFileSize: int64(1024 * 1024 * 25),
	}
}

// GetDoc fetches a urlStr (URL relative to the client's BaseURL) and returns the parsed response document.
func (c *Client) GetDoc(urlStr string, a ...interface{}) (*goquery.Document, error) {
	u, err := c.BaseURL.Parse(fmt.Sprintf(urlStr, a...))
	if err != nil {
		return nil, fmt.Errorf("error parsing url %q: %v", urlStr, err)
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching url %q: %v", u, err)
	}
	defer resp.Body.Close()

	// 5 retries to get the challenge if the status code is not http.StatusOK
	for i := 0; i < 5; i++ {
		if resp.StatusCode == http.StatusOK {
			break
		}
		resp, err = c.Client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error fetching url %q: %v", u, err)
		}
		defer resp.Body.Close()

		time.Sleep(time.Second * 1)
	}

	if resp.StatusCode == (http.StatusUnauthorized | http.StatusForbidden) {
		return nil, fmt.Errorf("received %v status code for url %q", resp.StatusCode, u)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing response body: %v", err)
	}

	return doc, nil
}

// GetJson fetches a urlStr (URL relative to the client's BaseURL) and returns the parsed response body.
func (c *Client) GetJson(urlStr string, a ...interface{}) (*http.Response, error) {
	u, err := c.BaseURL.Parse(fmt.Sprintf(urlStr, a...))
	if err != nil {
		return nil, fmt.Errorf("error parsing url %q: %v", urlStr, err)
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching url %q: %v", u, err)
	}
	defer resp.Body.Close()

	// 5 retries to get the challenge if the status code is not http.StatusOK
	for i := 0; i < 5; i++ {
		if resp.StatusCode == http.StatusOK {
			break
		}
		resp, err = c.Client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error fetching url %q: %v", u, err)
		}
		defer resp.Body.Close()

		time.Sleep(time.Second * 1)
	}

	if resp.StatusCode == (http.StatusUnauthorized | http.StatusForbidden) {
		return nil, fmt.Errorf("received %v status code for url %q", resp.StatusCode, u)
	}

	return resp, nil
}
