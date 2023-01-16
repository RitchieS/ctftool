package scraper

import (
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

// htmlForm struct represents an HTML form
type htmlForm struct {
	// Action is the URL where the form will be submitted to.
	Action string
	// Method is the HTTP method used to submit the form.
	Method string
	// Values are the form values.
	Values url.Values
}

// ParseForms takes in an html node and returns a slice of htmlForm structs.
// Each struct represents an HTML form found in the node, including its action, method, and input values.
//
//	forms := ParseForms(node)
//	for _, form := range forms {
//		fmt.Println(form.Action)
//		fmt.Println(form.Method)
//		fmt.Println(form.Values)
//	}
func ParseForms(node *html.Node) (forms []htmlForm) {
	// If the provided node is nil, return an empty slice
	if node == nil {
		return []htmlForm{}
	}

	// Create a new goquery document from the provided node
	doc := goquery.NewDocumentFromNode(node)
	// Find all forms in the document
	doc.Find("form").Each(func(_ int, s *goquery.Selection) {
		form := htmlForm{
			Values: url.Values{},
		}
		// Get the form's action attribute
		form.Action, _ = s.Attr("action")
		// Get the form's method attribute
		form.Method, _ = s.Attr("method")

		// Find all input elements within the form
		s.Find("input").Each(func(_ int, s *goquery.Selection) {
			name, _ := s.Attr("name")
			// If the input element does not have a name, skip it
			if name == "" {
				return
			}

			typ, _ := s.Attr("type")
			typ = strings.ToLower(typ)
			_, checked := s.Attr("checked")
			if (typ == "radio" || typ == "checkbox") && !checked {
				return
			}

			value, _ := s.Attr("value")
			form.Values.Add(name, value)
		})

		// Find all textarea elements within the form
		s.Find("textarea").Each(func(_ int, s *goquery.Selection) {
			name, _ := s.Attr("name")
			// If the textarea element does not have a name, skip it
			if name == "" {
				return
			}

			value := s.Text()
			form.Values.Add(name, value)
		})
		// Append the parsed form to the forms slice
		forms = append(forms, form)
	})

	return forms
}

// FetchAndSubmitForm takes in an http client, a url string, and a function that sets values for the form.
// The function fetches the form from the given url, parses it, and allows the setValues function to fill out the form.
// The form is then submitted and the response is returned, along with any error that may have occurred.
//
//	resp, err := FetchAndSubmitForm(client, "https://example.com/form", func(values url.Values) {
//		values.Set("username", "john")
//		values.Set("password", "password123")
//	})
//	if err != nil {
//		fmt.Println(err)
//	}
func FetchAndSubmitForm(client *http.Client, urlStr string, setValues func(values url.Values)) (*http.Response, error) {
	// Get the response from the provided url
	resp, err := client.Get(urlStr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse the HTML from the response body
	root, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	// Get the forms from the parsed HTML
	forms := ParseForms(root)
	if len(forms) == 0 {
		return nil, errors.New("no forms found")
	}
	form := forms[0]

	// Resolve the action URL for the form
	actionURL, err := url.Parse(form.Action)
	if err != nil {
		return nil, err
	}
	actionURL = resp.Request.URL.ResolveReference(actionURL)

	// Allow the caller to fill out the form
	if setValues != nil {
		setValues(form.Values)
	}

	// Store cookies from the response into the cookie jar
	client.Jar.SetCookies(actionURL, resp.Cookies())

	// Create a new request to submit the form
	req, err := http.NewRequest("POST", actionURL.String(), strings.NewReader(form.Values.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Set the CSRF token in the request header
	if csrf := ExtractCSRF(resp); csrf != "" {
		req.Header.Set("Csrf-Token", csrf)
	}

	// Submit the form
	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// ExtractCSRF takes in an http response, extracts the CSRF token from the body of the response and returns it as a string.
// The token is extracted using regular expressions to match specific patterns in the HTML body of the response.
//
//	csrf := ExtractCSRF(resp)
//	fmt.Println(csrf)
func ExtractCSRF(resp *http.Response) (csrf string) {
	// Parse the HTML from the response body
	root, err := html.Parse(resp.Body)
	if err != nil {
		return
	}

	// Create a goquery document from the parsed HTML
	doc := goquery.NewDocumentFromNode(root)

	// Compile the regex for extracting the CSRF token from the document's text
	// 'csrfNonce': "[a-zA-Z0-9]{64}",
	initRegex := regexp.MustCompile(`'csrfNonce': "([a-zA-Z0-9]{64})"`)
	initToken := initRegex.FindStringSubmatch(doc.Text())
	if len(initToken) == 2 {
		csrf = initToken[1]
	}

	// <input id="nonce" name="nonce" type="hidden" value="[a-zA-Z0-9]{64}">
	inputRegex := regexp.MustCompile(`<input id="nonce" name="nonce" type="hidden" value="([a-zA-Z0-9]{64})">`)
	inputToken := inputRegex.FindStringSubmatch(doc.Text())
	if len(inputToken) == 2 {
		csrf = inputToken[1]
	}

	return
}
