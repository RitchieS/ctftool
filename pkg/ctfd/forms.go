package ctfd

import (
	"fmt"
	"net/http"
	"net/url"
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

// parseForm parses the form and returns all the form elements beneath the given node. Form values include all input and textarea elements within the form. The values of the radio and checkbox inputs are included only if they are checked.
func parseForms(node *html.Node) (forms []htmlForm) {
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

func fetchAndSubmitForm(client *http.Client, urlStr string, setValues func(values url.Values)) (*http.Response, error) {
	resp, err := client.Get(urlStr)
	if err != nil {
		return nil, fmt.Errorf("error fetching url: %q: %v", urlStr, err)
	}

	defer resp.Body.Close()
	root, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing response body: %v", err)
	}

	forms := parseForms(root)
	if len(forms) == 0 {
		return nil, fmt.Errorf("no forms found at url: %q", urlStr)
	}
	form := forms[0]

	actionURL, err := url.Parse(form.Action)
	if err != nil {
		return nil, fmt.Errorf("error parsing form action: %q: %v", form.Action, err)
	}
	actionURL = resp.Request.URL.ResolveReference(actionURL)

	// allow caller to fill out the form
	if setValues != nil {
		setValues(form.Values)
	}

	// Store cookies from the response into the cookie jar
	client.Jar.SetCookies(actionURL, resp.Cookies())

	resp, err = client.PostForm(actionURL.String(), form.Values)
	if err != nil {
		return nil, fmt.Errorf("error submitting form: %q: %v", actionURL, err)
	}

	return resp, nil
}
