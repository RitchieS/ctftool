package ctfd

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/net/html"
)

func Test_ParseForms(t *testing.T) {
	tests := []struct {
		description string
		html        string
		forms       []htmlForm
	}{
		{"no forms", `<html><body></body></html>`, nil},
		{"empty form", `<html><body><form></form></body></html>`, []htmlForm{{Values: url.Values{}}}},
		{
			"single form with one value",
			`<html><body><form action="a" method="m"><input name="n1" value="v1"/></form></body></html>`,
			[]htmlForm{{
				Action: "a",
				Method: "m",
				Values: url.Values{
					"n1": {"v1"},
				},
			}},
		},
		{
			"two forms",
			`<html>
				<body>
					<form action="a1" method="m1"><input name="n1" value="v1"/></form>
					<form action="a2" method="m2"><input name="n2" value="v2"/></form>
				</body>
			</html>`,
			[]htmlForm{
				{
					Action: "a1",
					Method: "m1",
					Values: url.Values{
						"n1": {"v1"},
					},
				},
				{
					Action: "a2",
					Method: "m2",
					Values: url.Values{
						"n2": {"v2"},
					},
				},
			},
		},

		{
			"form with radio buttons (none checked)",
			`<html><form>
				<input type="radio" name="n1" value="v1">
				<input type="radio" name="n1" value="v2">
				<input type="radio" name="n1" value="v3">
			</form></html>`,
			[]htmlForm{{Values: url.Values{}}},
		},
		{
			"form with radio buttons",
			`<html><form>
				<input type="radio" name="n1" value="v1">
				<input type="radio" name="n1" value="v2">
				<input type="radio" name="n1" value="v3" checked>
			</form></html>`,
			[]htmlForm{{Values: url.Values{"n1": {"v3"}}}},
		},
		{
			"form with checkboxes",
			`<html><form>
				<input type="checkbox" name="n1" value="v1" checked>
				<input type="checkbox" name="n2" value="v2">
				<input type="checkbox" name="n3" value="v3" checked>
			</form></html>`,
			[]htmlForm{{Values: url.Values{"n1": {"v1"}, "n3": {"v3"}}}},
		},
		{
			"single form with textarea",
			`<html><form><textarea name="n1">v1</textarea></form></html>`,
			[]htmlForm{{Values: url.Values{"n1": {"v1"}}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			node, err := html.Parse(strings.NewReader(tt.html))
			if err != nil {
				t.Errorf("error parsing html: %v", err)
			}
			if got, want := parseForms(node), tt.forms; !cmp.Equal(got, want) {
				t.Errorf("parseForms(%q) returned %+v, want %+v", tt.html, got, want)
			}
		})
	}
}

// Test using the testdata/ctfd_login.html file.
func Test_ParseForms_Login(t *testing.T) {
	node, err := readFile("testdata/ctfd_login_full.html")
	if err != nil {
		t.Errorf("error reading file: %v", err)
	}

	forms := parseForms(node)
	if len(forms) != 1 {
		t.Errorf("expected 1 form, got %d", len(forms))
		return
	}

	want := url.Values{
		"name":     {""},
		"password": {""},
		"_submit":  {"Submit"},
		"nonce":    {"4a38d931755087a5512c817955dbb646c04adf71d36049c2d820854ffe17f7af"},
	}
	if got := forms[0].Values; !cmp.Equal(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func Test_FetchAndSubmitForm(t *testing.T) {
	client, mux, cleanup := setup()
	defer cleanup()

	var submitted bool

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `<html><form action="/login">
			<input type="text" name="name" value="test">
			<input type="password" name="password" value="test">
			<input type="submit" name="_submit" value="Submit">
			<input type="hidden" name="nonce" value="4a38d931755087a5512c817955dbb646c04adf71d36049c2d820854ffe17f7af">
		</form></html>`)
	})

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		want := url.Values{
			"name":     {"test"},
			"password": {"test"},
			"_submit":  {"Submit"},
			"nonce":    {"4a38d931755087a5512c817955dbb646c04adf71d36049c2d820854ffe17f7af"},
		}
		if got := r.Form; !cmp.Equal(got, want) {
			t.Errorf("got %+v, want %+v", got, want)
		}
		submitted = true
	})

	setValues := func(values url.Values) { values.Set("name", "test") }
	fetchAndSubmitForm(client.Client, client.BaseURL.String()+"/", setValues)
	if !submitted {
		t.Errorf("expected form to be submitted")
	}
}

// test errors like no name for input and textarea
func Test_ParseForms_Errors(t *testing.T) {
	tests := []struct {
		description string
		html        string
		returned    []htmlForm
		err         bool
	}{
		{
			"no name for input",
			`<html><form><input type="text" value="test"></form></html>`,
			// [{Action: Method: Values:map[]}]
			[]htmlForm{
				{
					Action: "",
					Method: "",
					Values: url.Values{},
				},
			},
			false,
		},
		{
			"no name for textarea",
			`<html><form><textarea>test</textarea></form></html>`,
			[]htmlForm{
				{
					Action: "",
					Method: "",
					Values: url.Values{},
				},
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			node, err := html.Parse(strings.NewReader(tt.html))
			if err != nil {
				t.Errorf("error parsing html: %v", err)
			}
			if got, want := parseForms(node), tt.returned; !cmp.Equal(got, want) {
				t.Errorf("parseForms(%q) returned %+v, want %+v", tt.html, got, want)
			}
		})
	}
}

// test node == nil for parseForms
func Test_ParseForms_Nil(t *testing.T) {
	tests := []struct {
		description string
		html        string
		returned    []htmlForm
		err         bool
	}{
		{
			"nil node",
			"",
			[]htmlForm{},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			if got, want := parseForms(nil), tt.returned; !cmp.Equal(got, want) {
				t.Errorf("parseForms(%q) returned %+v, want %+v", tt.html, got, want)
			}
		})
	}
}
