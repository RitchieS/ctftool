package ctfd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
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
	URLIsCTFD     bool      `json:"url_is_ctfd"`
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
type CTFTeam struct {
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

// IsCTFEventActive returns true if the CTF event is active based on the
// start and finish times
func IsCTFEventActive(event Event) bool {
	now := time.Now()

	if event.Finish.Before(now) {
		return false
	}

	return event.Start.Before(now) && event.Finish.After(now)
}

// Clean the description of a CTF event, removing \r\n and limiting the length
// of the description
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

// CleanCTFEvents will clean the CTF events, removing any events that are
// not "Open", are on-site, are not of jeopardy style or that have finished
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

// Retrieve information about all CTF events on CTFTime
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

	doc, err := c.getDoc(ctf_api)
	if err != nil {
		return nil, fmt.Errorf("failed to get CTF events: %v", err)
	}

	// unmarshal the json
	err = json.Unmarshal([]byte(doc.Text()), &events)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal CTF events: %v", err)
	}

	events, err = CleanCTFEvents(events)
	if err != nil {
		return events, fmt.Errorf("failed to clean CTF events: %v", err)
	}

	/* 	var wg sync.WaitGroup
	   	for i := 0; i < len(events); i++ {
	   		wg.Add(1)
	   		go func(e Event) {
	   			c.BaseURL, err = url.Parse(e.URL)
	   			if err != nil {
	   				return
	   			}

	   			err = c.Check()
	   			if err == nil {
	   				e.URLIsCTFD = true
	   			}

	   			wg.Done()
	   		}(events[i])
	   	}

	   	wg.Wait() */

	return events, nil
}

// Retrieve information about a specific CTF event on CTFTime
func (c *Client) GetCTFEvent(id int) (Event, error) {
	var event Event
	uri := fmt.Sprintf("api/v1/events/%d/", id)

	doc, err := c.getDoc(uri)
	if err != nil {
		return event, err
	}

	err = json.Unmarshal([]byte(doc.Text()), &event)
	if err != nil {
		return event, err
	}

	return event, nil
}

// Get information about a specific team on CTFTime
func (c *Client) GetCTFTeam(id int) (CTFTeam, error) {
	var team CTFTeam
	uri := fmt.Sprintf("api/v1/teams/%d/", id)

	doc, err := c.getDoc(uri)
	if err != nil {
		return team, err
	}

	err = json.Unmarshal([]byte(doc.Text()), &team)
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

// Get the top teams on CTFTime
func (c *Client) GetTopTeams() ([]TopTeam, error) {
	var teams TopTeams
	var result []TopTeam

	currentYear := time.Now().Year()
	// https://ctftime.org/api/v1/top/2022/
	uri := fmt.Sprintf("api/v1/top/%d/", currentYear)

	doc, err := c.getDoc(uri)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal([]byte(doc.Text()), &teams)
	if err != nil {
		return result, err
	}

	result = teams.Teams
	return result, nil
}
