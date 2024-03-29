package ctftime

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/ritchies/ctftool/pkg/scraper"
)

const ctftimeURL = "https://ctftime.org/"

var client = scraper.NewClient(nil)

// Struct for API Endpoint ctftime.org/api/v1/events/
type Event struct {
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

// IsActive takes in an Event and returns whether the event is currently active or not
// based on the current time. It compares the start and finish time of the event with the current time.
//
//	event := Event{
//		Start: time.Now().Add(-1 * time.Hour),
//		Finish: time.Now().Add(1 * time.Hour),
//	}
//	fmt.Println(IsActive(event)) // Output: true
func IsActive(event Event) bool {
	now := time.Now()

	if event.Finish.Before(now) {
		return false
	}

	return event.Start.Before(now) && event.Finish.After(now)
}

// CleanDescription takes in a string and removes any unnecessary new lines and ensures that the string is no longer than 1024 characters.
//
//	desc := "This is a long\ndescription\nwith\nmultiple\nlines."
//	cleanDesc := CleanDescription(desc)
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

// CleanCTFEvents takes in a slice of Event structs and performs several clean up operations on the slice.
// Title and Description fields of each event are trimmed and cleaned.
// Events that have finished are removed from the slice.
// The remaining events are sorted into two slices: active events and upcoming events.
// Active events are sorted by finish time, and upcoming events are sorted by start time.
// The two slices are then combined and returned, along with any error that may have occurred.
//
//	ctfEvents, err := CleanCTFEvents(events)
//	if err != nil {
//		fmt.Println(err)
//	}
func CleanCTFEvents(events []Event) ([]Event, error) {
	for i := 0; i < len(events); i++ {
		events[i].Title = strings.TrimSpace(events[i].Title)
		events[i].Description = strings.TrimSpace(events[i].Description)
		events[i].Description = CleanDescription(events[i].Description)
	}

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

	// create a slice of active and upcoming events
	var ctfEvents, activeEvents, upcomingEvents []Event
	for i := 0; i < len(events); i++ {
		if IsActive(events[i]) {
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

// GetCTFEvents retrieves CTF events from the ctftime API within the next 180 days.
// The events are unmarshaled from json and cleaned before being returned.
//
//	events, err := GetCTFEvents()
//	if err != nil {
//		fmt.Println(err)
//	}
func GetCTFEvents() ([]Event, error) {
	client.BaseURL, _ = url.Parse(ctftimeURL)
	var events []Event

	now := time.Now()
	start := now.Add(-time.Hour * 24 * 60).Unix()
	end := now.Add(time.Hour * 24 * 180).Unix()

	params := url.Values{}
	params.Add("start", fmt.Sprintf("%d", start))
	params.Add("end", fmt.Sprintf("%d", end))
	params.Add("limit", "100")

	ctf_api := fmt.Sprintf("api/v1/events/?%s", params.Encode())

	doc, err := client.GetDoc(ctf_api)
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

	return events, nil
}

// GetCTFEvent takes in an integer 'id' and returns an Event struct, along with any error that may have occurred.
// The function uses an http client to send a GET request to the ctftime API with the provided id,
// and parses the response body into an Event struct.
//
//	event, err := GetCTFEvent(1)
//	if err != nil {
//		fmt.Println(err)
//	}
func GetCTFEvent(id int) (Event, error) {
	client.BaseURL, _ = url.Parse(ctftimeURL)

	var event Event
	uri := fmt.Sprintf("api/v1/events/%d/", id)

	doc, err := client.GetDoc(uri)
	if err != nil {
		return event, err
	}

	err = json.Unmarshal([]byte(doc.Text()), &event)
	if err != nil {
		return event, err
	}

	return event, nil
}

// GetCTFTeam takes in an id and returns a CTFTeam struct, along with any error that may have occurred.
// The function uses the ctftimeURL to make a GET request to the API and retrieve the team information by id.
// The response body is then parsed into a CTFTeam struct using json.Unmarshal.
//
//	team, err := GetCTFTeam(1)
//	if err != nil {
//		fmt.Println(err)
//	}
func GetCTFTeam(id int) (CTFTeam, error) {
	client.BaseURL, _ = url.Parse(ctftimeURL)

	var team CTFTeam
	uri := fmt.Sprintf("api/v1/teams/%d/", id)

	doc, err := client.GetDoc(uri)
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
func GetTopTeams() ([]TopTeam, error) {
	client.BaseURL, _ = url.Parse(ctftimeURL)

	var teams TopTeams
	var result []TopTeam

	currentYear := time.Now().Year()
	// https://ctftime.org/api/v1/top/2022/
	uri := fmt.Sprintf("api/v1/top/%d/", currentYear)

	doc, err := client.GetDoc(uri)
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
