package ctf

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Solves struct {
	ChallengeID interface{} `json:"challenge_id"`
	AccountID   int         `json:"account_id"`
	TeamID      int         `json:"team_id"`
	UserID      int         `json:"user_id"`
	Value       int         `json:"value"`
	Date        time.Time   `json:"date"`
}
type Team struct {
	ID     int      `json:"id"`
	Name   string   `json:"name"`
	Solves []Solves `json:"solves"`
}

type TopTeamData struct {
	Num1  Team `json:"1"`
	Num2  Team `json:"2"`
	Num3  Team `json:"3"`
	Num4  Team `json:"4"`
	Num5  Team `json:"5"`
	Num6  Team `json:"6"`
	Num7  Team `json:"7"`
	Num8  Team `json:"8"`
	Num9  Team `json:"9"`
	Num10 Team `json:"10"`
}

// ScoreboardTop returns the top teams on the scoreboard
func (c *Client) ScoreboardTop(count int64) (TopTeamData, error) {
	response := new(struct {
		Data    TopTeamData `json:"data"`
		Success bool        `json:"success"`
	})

	resp, err := c.GetJson(fmt.Sprintf("api/v1/scoreboard/top/%d", count))
	if err != nil {
		return response.Data, fmt.Errorf("error fetching scoreboard from %q: %v", resp.Request.URL, err)
	}
	defer resp.Body.Close()

	// 5 retries to get the scoreboard if the status code is not http.StatusOK
	for i := 0; i < 5; i++ {
		if resp.StatusCode == http.StatusOK {
			break
		}
		resp, err = c.GetJson(fmt.Sprintf("api/v1/scoreboard/top/%d", count))
		if err != nil {
			return response.Data, fmt.Errorf("error fetching scoreboard from %q: %v", resp.Request.URL, err)
		}
		defer resp.Body.Close()

		time.Sleep(time.Second * 1)
	}

	if resp.StatusCode != http.StatusOK {
		return response.Data, fmt.Errorf("error fetching challenges from %q: %v", resp.Request.URL, resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return response.Data, fmt.Errorf("error unmarshalling scoreboard from %q: %v", resp.Request.URL, err)
	}

	if !response.Success {
		return response.Data, fmt.Errorf("failed to get scoreboard from %q", resp.Request.URL)
	}

	return response.Data, nil
}

// GetTeam returns the team information for a given team ID
// !TODO: This needs to be refactored to allow for a list of any number of teams
func (d *TopTeamData) GetTeam(number int) (*Team, error) {
	switch number {
	case 1:
		return &d.Num1, nil
	case 2:
		return &d.Num2, nil
	case 3:
		return &d.Num3, nil
	case 4:
		return &d.Num4, nil
	case 5:
		return &d.Num5, nil
	case 6:
		return &d.Num6, nil
	case 7:
		return &d.Num7, nil
	case 8:
		return &d.Num8, nil
	case 9:
		return &d.Num9, nil
	case 10:
		return &d.Num10, nil
	default:
		return nil, fmt.Errorf("invalid team number: %d", number)
	}
}
