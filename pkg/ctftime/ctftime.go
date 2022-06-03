package ctftime

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/publicsuffix"
)

// Struct for API Endpoint ctftime.org/api/v1/events/
type Event struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time

	Hidden bool

	ID            uint64    `json:"id"`
	CTFID         int       `json:"ctf_id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	URL           string    `json:"url"`
	Logo          string    `json:"logo"`
	Weight        float64   `json:"weight"`
	Onsite        bool      `json:"onsite"`
	Location      string    `json:"location"`
	Restrictions  string    `json:"restrictions"`
	Format        string    `json:"format"`
	FormatID      int       `json:"format_id"`
	Participants  int       `json:"participants"`
	CTFTimeURL    string    `json:"ctftime_url"`
	LiveFeed      string    `json:"live_feed"`
	IsVotableNow  bool      `json:"is_votable_now"`
	PublicVotable bool      `json:"public_votable"`
	Start         time.Time `json:"start"`
	Finish        time.Time `json:"finish"`
	/* Organizers    []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"organizers"` */
}

// Struct for API Endpoint ctftime.org/api/v1/teams/
type Team struct {
	ID           int      `json:"id"`
	Academic     bool     `json:"academic"`
	PrimaryAlias string   `json:"primary_alias"`
	Name         string   `json:"name"`
	Logo         string   `json:"logo"`
	Country      string   `json:"country"`
	Aliases      []string `json:"aliases"`
	Rating       map[string]struct {
		RatingPlace     int     `json:"rating_place"`
		OrganizerPoints float64 `json:"organizer_points"`
		RatingPoints    float64 `json:"rating_points"`
		CountryPlace    int     `json:"country_place"`
	} `json:"rating"`
}

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

	cookieJar, _ := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})

	// Set long timeout to avoid timeouts because CTFd is slow
	if transport == nil {
		transport = &http.Transport{
			MaxIdleConns:          10,
			IdleConnTimeout:       30 * time.Second,
			ResponseHeaderTimeout: time.Duration(30) * time.Second,
		}
	}

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

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request for %q: %v", urlStr, err)
	}

	// Set the User-Agent header
	req.Header.Set("User-Agent", "CTF Tool/1.0")

	resp, err := c.Client.Do(req)
	// resp, err := c.Client.Get(u.String())
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

// Return if a CTF is currently active
func IsCTFEventActive(event Event) bool {
	now := time.Now()

	return event.Start.Before(now) && event.Finish.After(now)
}

// Clean the description of a CTF event, removing \r\n
func CleanDescription(description string) string {
	re := regexp.MustCompile(`\r?\n`)
	description = re.ReplaceAllString(description, "\n")

	lines := strings.Split(description, "\n")

	var linesArray []string
	linesArray = append(linesArray, lines...)

	for len(strings.Join(linesArray, "")) > 1024 {
		linesArray = linesArray[1:]
	}

	description = strings.Join(linesArray, "\n")

	return description
}

// Clean CTF Events, return only 'Open' Jeopardy Style CTFs that are either active or upcoming
func CleanCTFEvents(events []Event) ([]Event, error) {
	for i := 0; i < len(events); i++ {
		events[i].Title = strings.TrimSpace(events[i].Title)
		events[i].Description = strings.TrimSpace(events[i].Description)
		events[i].Description = CleanDescription(events[i].Description)
	}

	// Remove events that are not "Open"
	/* 	for i := 0; i < len(events); i++ {
		if events[i].Restrictions != "Open" {
			events = append(events[:i], events[i+1:]...)
			i--
		}
	} */

	// Remove events that are onsite
	/* 	for i := 0; i < len(events); i++ {
		if events[i].Onsite {
			events = append(events[:i], events[i+1:]...)
			i--
		}
	} */

	// Remove events with format_id != 1 (Jeopardy Style)
	/* 	for i := 0; i < len(events); i++ {
		if events[i].FormatID != 1 {
			events = append(events[:i], events[i+1:]...)
			i--
		}
	} */

	// Remove events that have finished
	for i := 0; i < len(events); i++ {
		if events[i].Finish.Before(time.Now()) {
			events = append(events[:i], events[i+1:]...)
			i--
		}
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("no events found")
	}

	return events, nil
}

// Retrieve all active and upcoming CTF events from ctftime.org/api/v1/events/
func (c *Client) GetCTFEvents() ([]Event, error) {
	var events []Event

	now := time.Now()
	start := now.Add(-time.Hour * 24 * 60).Unix()
	end := now.Add(time.Hour * 24 * 180).Unix()

	params := url.Values{}
	params.Add("start", fmt.Sprintf("%d", start))
	params.Add("end", fmt.Sprintf("%d", end))
	params.Add("limit", "100")

	ctf_api := fmt.Sprintf("api/v1/events/?%s", params.Encode())

	goquerydoc, err := c.get(ctf_api)
	if err != nil {
		return nil, err
	}

	// get the json
	json_data := goquerydoc.Find("body").Text()

	// unmarshal the json
	err = json.Unmarshal([]byte(json_data), &events)
	if err != nil {
		return nil, err
	}

	events, err = CleanCTFEvents(events)
	if err != nil {
		return events, err
	}

	return events, nil
}

// Retrieve information about a specific CTF event on CTFTime
func (c *Client) GetCTFEvent(id int) (Event, error) {
	var event Event
	uri := fmt.Sprintf("api/v1/events/%d/", id)

	goquerydoc, err := c.get(uri)
	if err != nil {
		return event, err
	}

	// get the json
	json_data := goquerydoc.Find("body").Text()

	// unmarshal the json
	err = json.NewDecoder(strings.NewReader(json_data)).Decode(&event)
	if err != nil {
		return event, err
	}

	return event, nil
}

// Get information about a specific team on CTFTime
func (c *Client) GetCTFTeam(id int) (Team, error) {
	var team Team
	uri := fmt.Sprintf("api/v1/teams/%d/", id)

	goquerydoc, err := c.get(uri)
	if err != nil {
		return team, err
	}

	// get the json
	json_data := goquerydoc.Find("body").Text()

	// unmarshal the json
	err = json.NewDecoder(strings.NewReader(json_data)).Decode(&team)
	if err != nil {
		return team, err
	}

	return team, nil
}

type TopTeam struct {
	TeamName string  `json:"team_name"`
	Points   float64 `json:"points"`
	TeamID   int     `json:"team_id"`
}

type TopTeams struct {
	Teams []TopTeam `json:"2022"`
}

func (c *Client) GetTopTeams() ([]TopTeam, error) {
	var teams TopTeams
	var result []TopTeam

	currentYear := time.Now().Year()
	// https://ctftime.org/api/v1/top/2022/
	uri := fmt.Sprintf("api/v1/top/%d/", currentYear)

	goquerydoc, err := c.get(uri)
	if err != nil {
		return result, err
	}

	// get the json
	json_data := goquerydoc.Find("body").Text()

	// unmarshal the json
	err = json.NewDecoder(strings.NewReader(json_data)).Decode(&teams)
	if err != nil {
		return result, err
	}

	result = teams.Teams
	return result, nil
}
