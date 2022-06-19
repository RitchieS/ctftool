package ctf

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
)

// Check will check if the instance is a CTFd instance.
func (c *Client) Check() error {
	doc, err := c.GetDoc(c.BaseURL.String())
	if err != nil {
		return err
	}

	// Check for "<small class="text-muted">Powered by CTFd</small>"
	footerText := doc.Find("small.text-muted").Text()
	if !strings.Contains(footerText, "Powered by CTFd") {
		return errors.New("not a CTFd instance")
	}

	// check /login
	/* 	doc, err = c.get(c.BaseURL.String() + "/login")
	   	if err != nil {
	   		return err
	   	} */

	// Check if there are captcha's on the login page
	/* 	captchaURLs := []string{"https://www.google.com/recaptcha/api.js", "https://hcaptcha.com/1/api.js"}
	   	for _, captchaURL := range captchaURLs {
	   		if doc.Find("script[src]").FilterFunction(func(i int, s *goquery.Selection) bool {
	   			return strings.Contains(s.AttrOr("src", ""), captchaURL)
	   		}).Length() > 0 {
	   			return fmt.Errorf("captcha detected on login page")
	   		}
	   	} */

	return nil
}

// Authenticate will attempt to authenticate the client with the provided
// username and password.
func (c *Client) Authenticate() error {
	if err := c.Check(); err != nil {
		return err
	}

	setPassword := func(values url.Values) {
		values.Set("name", c.Creds.Username)
		values.Set("password", c.Creds.Password)
	}

	loginURL, err := joinPath(c.BaseURL.String(), "login")
	if err != nil {
		return err
	}

	resp, err := fetchAndSubmitForm(c.Client, loginURL.String(), setPassword)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to authenticate: %s", resp.Status)
	}

	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	errorRegex := regexp.MustCompile(`<div class="alert alert-danger alert-dismissible text-center" role="alert">\s*<span>([^<]+)</span>`)
	if errorRegex.MatchString(string(html)) {
		errorMessage := errorRegex.FindStringSubmatch(string(html))[1]
		return fmt.Errorf("failed to authenticate: %s", errorMessage)
	}

	return nil
}

// joinPath returns a URL string with the provided path elements joined to
// the base URL.
func joinPath(base string, elements ...string) (*url.URL, error) {
	u, err := url.Parse(base)
	if err != nil {
		return nil, err
	}

	if len(elements) > 0 {
		elements = append([]string{u.Path}, elements...)
		u.Path = path.Join(elements...)
	}

	return u, nil
}
