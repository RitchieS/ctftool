package ctf

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func setup() (client *Client, mux *http.ServeMux, cleanup func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)

	client = NewClient(nil)
	client.BaseURL, _ = url.Parse(server.URL + "/")

	return client, mux, server.Close
}

func copyTestFile(w io.Writer, filename string) error {
	f, err := os.Open("testdata/" + filename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(w, f)
	return err
}

func readFile(s string) (*html.Node, error) {
	b, err := ioutil.ReadFile(s)
	if err != nil {
		return nil, err
	}
	return html.Parse(strings.NewReader(string(b)))
}

func TestCheck(t *testing.T) {
	tests := []struct {
		description string
		html        string
		expected    bool
	}{
		{
			"ctfd instance",
			`<html><body><small class="text-muted">Powered by CTFd</small></body></html>`,
			true,
		},
	}

	client, mux, cleanup := setup()
	defer cleanup()

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s mux", test.description), func(t *testing.T) {
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(test.html))
				if err != nil {
					t.Errorf("Write() returned error: %v", err)
				}
			})
		})
		t.Run(fmt.Sprintf("%s check", test.description), func(t *testing.T) {
			err := client.Check()
			if err != nil {
				t.Errorf("Check() returned error: %v", err)
			}
		})
	}
}

// test Check for error
/* func TestCheckFail(t *testing.T) {
	tests := []struct {
		description  string
		responseBody string
		expected     bool
	}{
		{
			"check fail",
			`<html><body><small class="text-muted">Not Powered by CTFd</small></body></html>`,
			false,
		},
	}

	client, mux, cleanup := setup()
	defer cleanup()

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s mux", test.description), func(t *testing.T) {
			mux.HandleFunc("/nope", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(test.responseBody))
			})

			err := client.Check()
			if err == nil {
				t.Errorf("Check() returned no error")
			}

		})
	}
} */

// Test Check Failure
func TestCheckFailure(t *testing.T) {
	tests := []struct {
		description string
		html        string
		expected    bool
	}{
		{
			"not ctfd instance",
			`<html><body><small class="text-muted">Not A CTFd Instance</small></body></html>`,
			false,
		},
	}

	client, mux, cleanup := setup()
	defer cleanup()

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s mux", test.description), func(t *testing.T) {
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(test.html))
				if err != nil {
					t.Errorf("Write() returned error: %v", err)
				}
			})
		})

		t.Run(fmt.Sprintf("%s check", test.description), func(t *testing.T) {
			err := client.Check()
			if err == nil {
				t.Errorf("Check() returned no error")
			}
		})
	}
}

// Test joinPath
func TestJoinPath(t *testing.T) {
	tests := []struct {
		description string
		baseurl     string
		paths       []string
		expected    string
	}{
		{
			"join path",
			"http://localhost:1337",
			[]string{"login"},
			"http://localhost:1337/login",
		},
		{
			"join path with slash",
			"http://localhost:1337",
			[]string{"/", "login"},
			"http://localhost:1337/login",
		},
		{
			"join path with multiple paths",
			"http://localhost:1337",
			[]string{"/", "login", "register"},
			"http://localhost:1337/login/register",
		},
		{
			"join path with one element with path",
			"http://localhost:1337",
			[]string{"/api/v1/users"},
			"http://localhost:1337/api/v1/users",
		},
		{
			"join path with one element with path and without leading slash",
			"http://localhost:1337",
			[]string{"api/v1/users"},
			"http://localhost:1337/api/v1/users",
		},
		{
			"add nothing",
			"http://localhost:1337/api/v1/users",
			[]string{},
			"http://localhost:1337/api/v1/users",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actual, _ := joinPath(test.baseurl, test.paths...)
			if actual.String() != test.expected {
				t.Errorf("joinPath() returned %s, expected %s", actual, test.expected)
			}
		})
	}

}

// test joinPath error
func TestJoinPathError(t *testing.T) {
	tests := []struct {
		description string
		baseurl     string
		paths       []string
		expected    string
		expectedErr error
	}{
		/* 		{
		   			"invalid http",
		   			"http:/localhost:1337",
		   			[]string{"login"},
		   			"",
		   			fmt.Errorf(""),
		   		},
		   		{
		   			"empty url",
		   			"",
		   			[]string{""},
		   			"",
		   			fmt.Errorf(""),
		   		}, */
		{
			"url with CTLByte",
			// use 0x7f
			"http://localhost:1337/\x7f",
			[]string{"login"},
			"",
			// invalid control character in URL
			fmt.Errorf("net/url: invalid control character in URL"),
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			_, err := joinPath(test.baseurl, test.paths...)

			if err == nil {
				t.Errorf("joinPath() returned no error")
			}
		})
	}
}

// Test Authenticate
func TestAuthenticate(t *testing.T) {
	tests := []struct {
		description string
		htmlFile    string
		expected    bool
	}{
		{
			"authenticate",
			"ctfd_login_full.html",
			true,
		},
	}

	client, mux, cleanup := setup()
	defer cleanup()

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s mux", test.description), func(t *testing.T) {
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				if err := copyTestFile(w, test.htmlFile); err != nil {
					t.Errorf("copyTestFile() returned error: %v", err)
				}
			})
		})

		t.Run(fmt.Sprintf("%s check", test.description), func(t *testing.T) {
			client.Creds = &Credentials{
				Username: "admin",
				Password: "password",
			}
			err := client.Authenticate()
			if err != nil {
				t.Errorf("Authenticate() returned error: %v", err)
			}
		})
	}
}

// Test AuthenticateFails
func TestAuthenticateFails(t *testing.T) {
	tests := []struct {
		description string
		htmlFile    string
		expected    bool
	}{
		{
			"authenticate fail non ctfd",
			"example.html",
			false,
		},
	}

	client, mux, cleanup := setup()
	defer cleanup()

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s mux", test.description), func(t *testing.T) {
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				if err := copyTestFile(w, test.htmlFile); err != nil {
					t.Errorf("copyTestFile() returned error: %v", err)
				}
			})
		})

		t.Run(fmt.Sprintf("%s check", test.description), func(t *testing.T) {
			client.Creds = &Credentials{
				Username: "admin",
				Password: "password",
			}
			err := client.Authenticate()
			if err == nil {
				t.Errorf("Authenticate() returned error: %v", err)
			}
		})
	}
}

// test authenticate with empty baseurl
func TestAuthenticateEmptyBaseurl(t *testing.T) {
	client, _, cleanup := setup()
	defer cleanup()

	fakeBaseurl, _ := url.Parse("")
	client.BaseURL = fakeBaseurl

	client.Creds = &Credentials{
		Username: "admin",
		Password: "password",
	}

	err := client.Authenticate()
	if err == nil {
		t.Errorf("Authenticate() returned error: %v", err)
		return
	}
}

// test authenticate with broken ctfd login page
func TestAuthenticateBrokenLoginPage(t *testing.T) {
	tests := []struct {
		description string
		htmlFile    string
		expected    bool
	}{
		{
			"authenticate fail non ctfd",
			"ctfd_nologin.html",
			false,
		},
	}

	client, mux, cleanup := setup()
	defer cleanup()

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s mux", test.description), func(t *testing.T) {
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				if err := copyTestFile(w, test.htmlFile); err != nil {
					t.Errorf("copyTestFile() returned error: %v", err)
				}
			})
		})

		t.Run(fmt.Sprintf("%s check", test.description), func(t *testing.T) {
			client.Creds = &Credentials{
				Username: "admin",
				Password: "password",
			}
			err := client.Authenticate()
			if err == nil {
				t.Errorf("Authenticate() returned error: %v", err)
			}
		})

	}
}

// check error regex
func TestCheckErrorRegex(t *testing.T) {
	tests := []struct {
		description string
		htmlFile    string
		expected    bool
	}{
		{
			"check error regex",
			"ctfd_login_failed.html",
			false,
		},
	}

	client, mux, cleanup := setup()
	defer cleanup()

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s mux", test.description), func(t *testing.T) {
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				if err := copyTestFile(w, test.htmlFile); err != nil {
					t.Errorf("copyTestFile() returned error: %v", err)
				}
			})
		})

		t.Run(fmt.Sprintf("%s check", test.description), func(t *testing.T) {
			client.Creds = &Credentials{
				Username: "admin",
				Password: "password",
			}
			err := client.Authenticate()
			if err == nil {
				t.Errorf("Authenticate() returned error: %v", err)
			}
		})

	}
}
