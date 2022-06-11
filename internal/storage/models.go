package storage

import (
	"encoding/json"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type Event struct {
	gorm.Model

	Hidden bool

	ID            uint64
	CTFID         int
	Title         string
	Description   string
	URL           string
	URLIsCTFD     bool
	Logo          string
	Weight        float64
	Onsite        bool
	Location      string
	Restrictions  string
	Format        string
	FormatID      int
	Participants  int
	CTFTimeURL    string
	LiveFeed      string
	IsVotableNow  bool
	PublicVotable bool
	Start         time.Time
	Finish        time.Time
}

type EventCustomTitle struct {
	gorm.Model

	ID    uint64
	Title string
}

type EventCustomDescription struct {
	gorm.Model

	ID          uint64
	Description string
}

type EventCustomDate struct {
	gorm.Model

	ID     uint64
	Start  time.Time
	Finish time.Time
}

type EventCustomURL struct {
	gorm.Model

	ID  uint64
	URL string
}

// MarshallCSV returns values as a slice
func (ctf *Event) MarshallCSV() (res []string) {
	return []string{strconv.FormatUint(ctf.ID, 10),
		strconv.Itoa(ctf.CTFID),
		ctf.Title,
		ctf.Description,
		ctf.URL,
		ctf.Logo,
		strconv.FormatFloat(ctf.Weight, 'f', -1, 64),
		strconv.FormatBool(ctf.Onsite),
		ctf.Location,
		ctf.Restrictions,
		ctf.Format,
		strconv.Itoa(ctf.FormatID),
		strconv.Itoa(ctf.Participants),
		ctf.CTFTimeURL,
		ctf.LiveFeed,
		strconv.FormatBool(ctf.IsVotableNow),
		strconv.FormatBool(ctf.PublicVotable),
		ctf.Start.Format(time.RFC3339),
		ctf.Finish.Format(time.RFC3339),
	}
}

// MarshallJSON returns values as a slice
func (ctf *Event) MarshallJSON() ([]byte, error) {
	var tmp struct {
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
	}

	tmp.ID = ctf.ID
	tmp.CTFID = ctf.CTFID
	tmp.Title = ctf.Title
	tmp.Description = ctf.Description
	tmp.URL = ctf.URL
	tmp.URLIsCTFD = ctf.URLIsCTFD
	tmp.Logo = ctf.Logo
	tmp.Weight = ctf.Weight
	tmp.Onsite = ctf.Onsite
	tmp.Location = ctf.Location
	tmp.Restrictions = ctf.Restrictions
	tmp.Format = ctf.Format
	tmp.FormatID = ctf.FormatID
	tmp.Participants = ctf.Participants
	tmp.CTFTimeURL = ctf.CTFTimeURL
	tmp.LiveFeed = ctf.LiveFeed
	tmp.IsVotableNow = ctf.IsVotableNow
	tmp.PublicVotable = ctf.PublicVotable
	tmp.Start = ctf.Start
	tmp.Finish = ctf.Finish

	return json.Marshal(&tmp)
}
