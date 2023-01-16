package ctfd

import (
	"encoding/json"
	"fmt"
)

type ChallengesData struct {
	ID         int64         `json:"id"`
	Type       string        `json:"type"`
	Name       string        `json:"name"`
	Value      int64         `json:"value"`
	Solves     int64         `json:"solves"`
	SolvedByMe bool          `json:"solved_by_me"`
	Category   string        `json:"category"`
	Tags       []interface{} `json:"tags"`
}

// ListChallenges returns a list of challenges
func ListChallenges() ([]ChallengesData, error) {
	response := new(struct {
		Success bool             `json:"success"`
		Data    []ChallengesData `json:"data"`
	})

	resp, err := c.GetJson("api/v1/challenges")
	if err != nil {
		return nil, fmt.Errorf("failed to get challenges: %v", err)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return nil, fmt.Errorf("failed to decode challenges: %v", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("failed to get challenges from %q", resp.Request.URL)
	}
	return response.Data, nil
}
