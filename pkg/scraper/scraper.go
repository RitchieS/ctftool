package scraper

import (
	"fmt"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
	"go.uber.org/ratelimit"
	"golang.org/x/net/publicsuffix"
)

// Client struct stores the http client, base url and credentials used to communicate with the server
type Client struct {
	Client      *http.Client // http client used to make requests to the server
	BaseURL     *url.URL     // base url of the server
	Creds       *Credentials // credentials used for authentication
	MaxFileSize int64        // maximum file size allowed
}

// Credentials struct stores the username and password used for authentication
type Credentials struct {
	Username string
	Password string
	Token    string
}

// NewClient returns a new instance of the Client struct with a specified transport.
// If no transport is provided, a default transport with a long timeout will be used.
// The client also uses a cookie jar and sets a default max file size of 25MB.
//
//	transport := &http.Transport{}
//	client := NewClient(transport)
func NewClient(transport http.RoundTripper) *Client {
	// Create a new cookie jar using publicsuffix.List as the public suffix list
	cookieJar, _ := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})

	// Check if the provided transport is nil. If it is, create a new transport with custom timeout and connection settings.
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

	// Return a new client with the provided transport wrapped in a NewTransport, and the cookie jar set to the created cookie jar.
	// Also set the Creds field to a new Credentials struct, and the MaxFileSize to 25MB
	return &Client{
		Client: &http.Client{
			Transport: NewTransport(transport),
			Jar:       cookieJar,
		},
		Creds:       &Credentials{},
		MaxFileSize: int64(1024 * 1024 * 25),
	}
}

// GetDoc takes in a url string and an optional list of interfaces, formats the url and sends a GET request.
// The response body is then parsed into a goquery document and returned, along with any error that may have occurred.
//
//	doc, err := client.GetDoc("https://example.com/%v", "path")
//	if err != nil {
//		fmt.Println(err)
//	}
func (c *Client) GetDoc(urlStr string, a ...interface{}) (*goquery.Document, error) {
	// Create a new URL by parsing the provided URL string and any additional arguments using the fmt.Sprintf function
	// and the BaseURL field of c.
	u, err := c.BaseURL.Parse(fmt.Sprintf(urlStr, a...))
	if err != nil {
		return nil, err
	}

	// Create a new GET request using the new URL.
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Perform the request using the Client's DoRequest method.
	resp, err := c.DoRequest(req)
	if err != nil {
		return nil, err
	}

	// Use goquery to parse the response body into a Document.
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	// Return the parsed Document.
	return doc, nil
}

// GetJson takes in a url string and an optional list of interfaces, formats the url and sends a GET request.
// The response is returned, along with any error that may have occurred.
//
//	resp, err := client.GetJson("https://example.com/%v", "path")
//	if err != nil {
//		fmt.Println(err)
//	}
func (c *Client) GetJson(urlStr string, a ...interface{}) (*http.Response, error) {
	// Create a new URL by parsing the provided URL string and any additional arguments using the fmt.Sprintf function
	// and the BaseURL field of c.
	u, err := c.BaseURL.Parse(fmt.Sprintf(urlStr, a...))
	if err != nil {
		return nil, err
	}

	// Create a new GET request using the new URL.
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Set headers
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	// Set the request headers using the SetHeaders method.
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Perform the request using the Client's DoRequest method.
	resp, err := c.DoRequest(req)
	if err != nil {
		return nil, err
	}

	// Return the response.
	return resp, nil
}

// DoRequest takes in an http request and sends it to the specified client.
// If the response status code is not http.StatusOK, the request will be retried up to 5 times with a rate limit of 1 request per second.
// If the final response status code is between http.StatusBadRequest and http.StatusNetworkAuthenticationRequired, an error will be returned.
//
//	resp, err := client.DoRequest(req)
//	if err != nil {
//		fmt.Println(err)
//	}
func (c *Client) DoRequest(req *http.Request) (*http.Response, error) {
	// Create a new rate limiter with a limit of 1 request per second.
	rl := ratelimit.New(1)

	// Set Authorization header if token is not empty.
	if c.Creds.Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Token %s", c.Creds.Token))
	}

	// Perform the request and capture the response and error.
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	// If the response status code is not http.StatusOK, perform up to 5 retries.
	for i := 0; i < 5; i++ {
		if resp.StatusCode == http.StatusOK {
			break
		}
		resp, err = c.Client.Do(req)
		if err != nil {
			return nil, err
		}

		// Wait for the rate limiter to allow another request before retrying.
		rl.Take()
	}

	// If the response status code is between http.StatusBadRequest and http.StatusNetworkAuthenticationRequired, return an error with the status code and text.
	if resp.StatusCode >= http.StatusBadRequest &&
		resp.StatusCode <= http.StatusNetworkAuthenticationRequired {
		return nil, fmt.Errorf("received status code %d (%s)", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	// Return the response.
	return resp, nil
}
