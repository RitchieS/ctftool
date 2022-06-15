package ctf

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
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
func (c *Client) ListChallenges() ([]ChallengesData, error) {
	response := new(struct {
		Success bool             `json:"success"`
		Data    []ChallengesData `json:"data"`
	})

	challengeAPI, err := joinPath(c.BaseURL.String(), "api/v1/challenges")
	if err != nil {
		return nil, fmt.Errorf("error joining path: %v", err)
	}

	resp, err := c.Client.Get(challengeAPI.String())
	if err != nil {
		return nil, fmt.Errorf("error fetching challenges from %s: %v", challengeAPI.String(), err)
	}
	defer resp.Body.Close()

	// 5 retries to get the challenge if the status code is not http.StatusOK
	for i := 0; i < 5; i++ {
		if resp.StatusCode == http.StatusOK {
			break
		}
		resp, err = c.Client.Get(challengeAPI.String())
		if err != nil {
			return nil, fmt.Errorf("error fetching challenges from %s: %v", challengeAPI.String(), err)
		}
		defer resp.Body.Close()

		time.Sleep(time.Second * 1)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error fetching challenges: received %v status code from %q", resp.StatusCode, challengeAPI.String())
	}

	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling challenges from %q: %v", challengeAPI.String(), err)
	}

	if !response.Success {
		return nil, fmt.Errorf("failed to get challenges")
	}
	return response.Data, nil
}
