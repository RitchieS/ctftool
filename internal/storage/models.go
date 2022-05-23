package storage

import (
	"encoding/json"
	"strconv"

	"gorm.io/gorm"
)

type CTFEvent struct {
	gorm.Model

	ID          int
	Title       string
	Description string
	Start       string
	Finish      string
	Weight      float64
}

// MarshallCSV returns values as a slice
func (ctf *CTFEvent) MarshallCSV() (res []string) {
	return []string{strconv.Itoa(ctf.ID),
		ctf.Title,
		ctf.Description,
		ctf.Start,
		ctf.Finish,
		strconv.FormatFloat(ctf.Weight, 'f', 2, 64)}
}

// MarshallJSON returns values as a slice
func (ctf *CTFEvent) MarshallJSON() ([]byte, error) {
	var tmp struct {
		ID          int     `json:"id"`
		Title       string  `json:"title"`
		Description string  `json:"description"`
		Start       string  `json:"start"`
		Finish      string  `json:"finish"`
		Weight      float64 `json:"weight"`
	}

	tmp.ID = ctf.ID
	tmp.Title = ctf.Title
	tmp.Description = ctf.Description
	tmp.Start = ctf.Start
	tmp.Finish = ctf.Finish
	tmp.Weight = ctf.Weight

	return json.Marshal(&tmp)
}
