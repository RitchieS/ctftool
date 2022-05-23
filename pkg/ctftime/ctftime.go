package ctftime

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Struct for API Endpoint ctftime.org/api/v1/events/
type Event struct {
	ID            int       `json:"id"`
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
	Organizers    []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"organizers"`
	Duration struct {
		Hours int `json:"hours"`
		Days  int `json:"days"`
	} `json:"duration"`
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

	// Remove events that are onsite
	for i := 0; i < len(events); i++ {
		if events[i].Onsite {
			events = append(events[:i], events[i+1:]...)
			i--
		}
	}

	// Remove events with format_id != 1 (Jeopardy Style)
	for i := 0; i < len(events); i++ {
		if events[i].FormatID != 1 {
			events = append(events[:i], events[i+1:]...)
			i--
		}
	}

	// Remove events that have finished
	for i := 0; i < len(events); i++ {
		if events[i].Finish.Before(time.Now()) {
			events = append(events[:i], events[i+1:]...)
			i--
		}
	}

	// Remove events that are not "Open"
	for i := 0; i < len(events); i++ {
		if events[i].Restrictions != "Open" {
			events = append(events[:i], events[i+1:]...)
			i--
		}
	}

	now := time.Now()
	maxTime := now.AddDate(0, 0, 90)

	// Remove events that are not active
	for i := 0; i < len(events); i++ {
		if events[i].Start.After(maxTime) {
			events = append(events[:i], events[i+1:]...)
			i--
		}
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("no events found")
	}

	// create a slice of active and upcoming events
	var ctfEvents, activeEvents, upcomingEvents []Event
	for i := 0; i < len(events); i++ {
		if IsCTFEventActive(events[i]) {
			activeEvents = append(activeEvents, events[i])
		} else {
			upcomingEvents = append(upcomingEvents, events[i])
		}
	}

	// Sort the active events by finish time
	sort.Slice(activeEvents, func(i, j int) bool {
		return activeEvents[i].Finish.Before(activeEvents[j].Finish)
	})

	// Sort upcoming events by start time
	sort.Slice(upcomingEvents, func(i, j int) bool {
		return upcomingEvents[i].Start.Before(upcomingEvents[j].Start)
	})

	// Combine the active and upcoming events
	ctfEvents = append(ctfEvents, activeEvents...)
	ctfEvents = append(ctfEvents, upcomingEvents...)

	return ctfEvents, nil
}

// Retrieve all active and upcoming CTF events from ctftime.org/api/v1/events/
func GetCTFEvents() ([]Event, error) {
	var events []Event

	now := time.Now()
	start := now.Add(-time.Hour * 24 * 14).Unix()
	end := now.Add(time.Hour * 24 * 60).Unix()

	params := url.Values{}
	params.Add("start", fmt.Sprintf("%d", start))
	params.Add("end", fmt.Sprintf("%d", end))
	params.Add("limit", "100")

	ctf_api := fmt.Sprintf("https://ctftime.org/api/v1/events/?%s", params.Encode())

	req, err := http.NewRequest("GET", ctf_api, nil)
	if err != nil {
		return events, err
	}

	req.Header.Set("User-Agent", "Go CTFTime API Client/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return events, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return events, fmt.Errorf("error getting events: %s", resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(&events)
	if err != nil {
		return events, err
	}

	events, err = CleanCTFEvents(events)
	if err != nil {
		return events, err
	}

	return events, nil
}

// Retrieve information about a specific CTF event on CTFTime
func GetCTFEvent(id int) (Event, error) {
	var event Event
	url := fmt.Sprintf("https://ctftime.org/api/v1/events/%d/", id)

	// build a new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return event, err
	}

	// set the header
	req.Header.Set("User-Agent", "Go CTFTime API Client/1.0")

	// do the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return event, err
	}

	// close the response body
	defer resp.Body.Close()

	// unmarshal the response
	err = json.NewDecoder(resp.Body).Decode(&event)
	if err != nil {
		return event, err
	}

	return event, nil
}

// Get information about a specific team on CTFTime
func GetCTFTeam(id int) (Team, error) {
	var team Team
	url := fmt.Sprintf("https://ctftime.org/api/v1/teams/%d/", id)

	// build a new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return team, err
	}

	// set the header
	req.Header.Set("User-Agent", "Go CTFTime API Client/1.0")

	// do the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return team, err
	}

	// close the response body
	defer resp.Body.Close()

	// if the response is not 200, return an error
	if resp.StatusCode != 200 {
		return team, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// unmarshal the response
	err = json.NewDecoder(resp.Body).Decode(&team)
	if err != nil {
		return team, err
	}

	return team, nil
}
