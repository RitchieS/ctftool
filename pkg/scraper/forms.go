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

type htmlForm struct {
	// Action is the URL where the form will be submitted to.
	Action string
	// Method is the HTTP method used to submit the form.
	Method string
	// Values are the form values.
	Values url.Values
}

// parseForm parses the form and returns all the form elements beneath the
// given node. Form values include all input and textarea elements within
// the form. The values of the radio and checkbox inputs are included
// only if they are checked.
func ParseForms(node *html.Node) (forms []htmlForm) {
	if node == nil {
		return []htmlForm{}
	}

	doc := goquery.NewDocumentFromNode(node)
	doc.Find("form").Each(func(_ int, s *goquery.Selection) {
		form := htmlForm{
			Values: url.Values{},
		}
		form.Action, _ = s.Attr("action")
		form.Method, _ = s.Attr("method")

		s.Find("input").Each(func(_ int, s *goquery.Selection) {
			name, _ := s.Attr("name")
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

		s.Find("textarea").Each(func(_ int, s *goquery.Selection) {
			name, _ := s.Attr("name")
			if name == "" {
				return
			}

			value := s.Text()
			form.Values.Add(name, value)
		})
		forms = append(forms, form)
	})

	return forms
}

// FetchAndSubmitForm fetches the given URL and submits the form with the given
// values. The form is parsed and the values are set on the form. The form is
// then submitted and the response is returned.
func FetchAndSubmitForm(client *http.Client, urlStr string, setValues func(values url.Values)) (*http.Response, error) {
	resp, err := client.Get(urlStr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	root, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	forms := ParseForms(root)
	if len(forms) == 0 {
		return nil, errors.New("no forms found")
	}
	form := forms[0]

	actionURL, err := url.Parse(form.Action)
	if err != nil {
		return nil, err
	}
	actionURL = resp.Request.URL.ResolveReference(actionURL)

	// allow caller to fill out the form
	if setValues != nil {
		setValues(form.Values)
	}

	// Store cookies from the response into the cookie jar
	client.Jar.SetCookies(actionURL, resp.Cookies())

	req, err := http.NewRequest("POST", actionURL.String(), strings.NewReader(form.Values.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Set CSRF token
	if csrf := ExtractCSRF(resp); csrf != "" {
		req.Header.Set("Csrf-Token", csrf)
	}

	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func ExtractCSRF(resp *http.Response) string {
	root, err := html.Parse(resp.Body)
	if err != nil {
		return ""
	}

	var csrf string

	doc := goquery.NewDocumentFromNode(root)

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

	return csrf
}
