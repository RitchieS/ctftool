package ctfd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Response struct {
	Success bool        `json:"success"`
	Data    TopTeamData `json:"data"`
}
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

func (c *Client) ScoreboardTop(count int64) (TopTeamData, error) {
	scoreboard := new(Response)

	scoreboardAPI, err := joinPath(c.BaseURL.String(), "api/v1/scoreboard/top", fmt.Sprintf("%d", count))
	if err != nil {
		return scoreboard.Data, fmt.Errorf("error joining path: %v", err)
	}

	resp, err := c.Client.Get(scoreboardAPI.String())
	if err != nil {
		return scoreboard.Data, fmt.Errorf("error fetching scoreboard from %q: %v", scoreboardAPI.String(), err)
	}
	defer resp.Body.Close()

	// 5 retries to get the scoreboard if the status code is not http.StatusOK
	for i := 0; i < 5; i++ {
		if resp.StatusCode == http.StatusOK {
			break
		}
		resp, err = c.Client.Get(scoreboardAPI.String())
		if err != nil {
			return scoreboard.Data, fmt.Errorf("error fetching scoreboard from %q: %v", scoreboardAPI.String(), err)
		}
		defer resp.Body.Close()

		time.Sleep(time.Second * 1)
	}

	if resp.StatusCode != http.StatusOK {
		return scoreboard.Data, fmt.Errorf("error fetching scoreboard: received %v status code", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(scoreboard)
	if err != nil {
		return scoreboard.Data, fmt.Errorf("error unmarshalling scoreboard from %q: %v", scoreboardAPI.String(), err)
	}

	if !scoreboard.Success {
		return scoreboard.Data, fmt.Errorf("failed to get scoreboard")
	}

	return scoreboard.Data, nil
}

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
