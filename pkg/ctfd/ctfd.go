package ctfd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"

	"github.com/ritchies/ctftool/pkg/scraper"
)

var client = scraper.NewClient(nil)

func NewClient() *scraper.Client {
	return client
}

// Check will check if the instance is a CTFd instance.
func Check() error {
	// make a request to https://demo.ctfd.io/api/v1/challenges
	resp, err := client.GetJson(fmt.Sprintf("%s/api/v1/challenges", client.BaseURL.String()))
	if err != nil {
		return fmt.Errorf("cant reach CTFd instance: %s", err)
	}

	// Check if the response is not OK
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to check: %s", resp.Status)
	}

	// Read the response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Parse the response as JSON
	var response map[string]interface{}
	err = json.Unmarshal(data, &response)
	if err != nil {
		return err
	}

	// Check if the response contains the "success" field
	if _, ok := response["success"]; !ok {
		return errors.New("not a CTFd instance")
	}

	return nil
}

// Authenticate will attempt to authenticate the client with the provided
// username and password.
func Authenticate() error {
	setPassword := func(values url.Values) {
		values.Set("name", client.Creds.Username)
		values.Set("password", client.Creds.Password)
	}

	loginURL, err := joinPath(client.BaseURL.String(), "login")
	if err != nil {
		return err
	}

	resp, err := scraper.FetchAndSubmitForm(client.Client, loginURL.String(), setPassword)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to authenticate: %s", resp.Status)
	}

	html, err := io.ReadAll(resp.Body)
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
