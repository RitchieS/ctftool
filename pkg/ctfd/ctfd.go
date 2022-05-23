package ctfd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/publicsuffix"
)

type Client struct {
	Client  *http.Client
	BaseURL *url.URL
	Log     *logrus.Logger
}

// NewClient constructs a new Client. If transport is nil, a default transport is used.
func NewClient(transport http.RoundTripper) *Client {
	log := logrus.New()

	log.SetFormatter(&logrus.TextFormatter{
		DisableSorting:         false,
		DisableTimestamp:       true,
		DisableLevelTruncation: true,
		ForceColors:            true,
		ForceQuote:             true,
		PadLevelText:           true,
		QuoteEmptyFields:       true,
	})

	// set log level to debug
	log.SetLevel(logrus.DebugLevel)

	cookieJar, _ := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})

	return &Client{
		Client: &http.Client{
			Transport: transport,
			Jar:       cookieJar,
		},
		Log: log,
	}
}

// get fetches a urlStr (URL relative to the client's BaseURL) and returns the parsed response document.
func (c *Client) get(urlStr string, a ...interface{}) (*goquery.Document, error) {
	u, err := c.BaseURL.Parse(fmt.Sprintf(urlStr, a...))
	if err != nil {
		return nil, fmt.Errorf("error parsing url %q: %v", urlStr, err)
	}

	resp, err := c.Client.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("error fetching url %q: %v", u, err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("received %v status code for url %q", resp.StatusCode, u)
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing response body: %v", err)
	}

	return doc, nil
}

// Check will check if the instance is a CTFd instance.
func (c *Client) Check() error {
	doc, err := c.get(c.BaseURL.String())
	if err != nil {
		return err
	}

	// Check for "<small class="text-muted">Powered by CTFd</small>"
	if doc.Find("small.text-muted").Text() != "Powered by CTFd" {
		return fmt.Errorf("instance is not a CTFd instance")
	}

	// check /login
	doc, err = c.get(c.BaseURL.String() + "/login")
	if err != nil {
		return err
	}

	// Check if there are captcha's on the login page
	captchaURLs := []string{"https://www.google.com/recaptcha/api.js", "https://hcaptcha.com/1/api.js"}
	for _, captchaURL := range captchaURLs {
		if doc.Find("script[src]").FilterFunction(func(i int, s *goquery.Selection) bool {
			return strings.Contains(s.AttrOr("src", ""), captchaURL)
		}).Length() > 0 {
			return fmt.Errorf("captcha detected on login page")
		}
	}

	return nil
}

// Authenticate client to the CTFd instance with the given username, password.
func (c *Client) Authenticate(username, password string) error {
	log := c.Log

	if err := c.Check(); err != nil {
		return err
	}

	setPassword := func(values url.Values) {
		values.Set("name", username)
		values.Set("password", password)
	}

	loginURL, err := joinPath(c.BaseURL.String(), "login")
	if err != nil {
		return fmt.Errorf("error joining path: %v", err)
	}

	log.WithField("url", loginURL).Debug("fetching login page")

	log.WithFields(logrus.Fields{
		"username": username,
		"password": password,
	}).Debug("Authenticating")

	resp, err := fetchAndSubmitForm(c.Client, loginURL.String(), setPassword)
	if err != nil {
		return fmt.Errorf("error authenticating: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error authenticating: received %v status code", resp.StatusCode)
	}

	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err)
	}

	errorRegex := regexp.MustCompile(`<div class="alert alert-danger alert-dismissible text-center" role="alert">\s*<span>([^<]+)</span>`)
	if errorRegex.MatchString(string(html)) {
		errorMessage := errorRegex.FindStringSubmatch(string(html))[1]
		return fmt.Errorf("error authenticating: %s", errorMessage)
	}

	return nil
}

// joinPath returns a URL string with the provided path elements joined to
// the existing path of base and the resulting path cleaned of any ./ or ../ elements.
func joinPath(base string, elements ...string) (*url.URL, error) {
	u, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("error parsing base url: %v", err)
	}

	if len(elements) > 0 {
		elements = append([]string{u.Path}, elements...)
		u.Path = path.Join(elements...)
	}

	return u, nil
}
