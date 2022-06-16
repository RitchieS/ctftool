package ctf

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestListChallenges(t *testing.T) {
	challengex := new(struct {
		Data    []ChallengesData `json:"data"`
		Success bool             `json:"success"`
	})

	challengex.Success = true
	challengex.Data = []ChallengesData{
		{
			ID:         1,
			Type:       "file",
			Name:       "test challenge",
			Value:      42,
			Solves:     1337,
			SolvedByMe: false,
			Category:   "file",
		},
	}

	// setup mux
	client, mux, cleanup := setup()
	defer cleanup()

	// mock response
	mux.HandleFunc("/api/v1/challenges", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(challengex)
	})

	// test
	challenges, err := client.ListChallenges()
	if err != nil {
		t.Errorf("error listing challenges: %v", err)
		return
	}

	if len(challenges) != 1 {
		t.Errorf("expected 1 challenge, got %d", len(challenges))
		return
	}

	if challenges[0].ID != 1 {
		t.Errorf("expected challenge ID 1, got %d", challenges[0].ID)
		return
	}

	if challenges[0].Type != "file" {
		t.Errorf("expected challenge type 'file', got %s", challenges[0].Type)
		return
	}

	if challenges[0].Name != "test challenge" {
		t.Errorf("expected challenge name 'test challenge', got %s", challenges[0].Name)
		return
	}

	if challenges[0].Value != 42 {
		t.Errorf("expected challenge value 42, got %d", challenges[0].Value)
		return
	}

	if challenges[0].Solves != 1337 {
		t.Errorf("expected challenge solves 1337, got %d", challenges[0].Solves)
		return
	}

}

func TestListChallenges_Error(t *testing.T) {
	tests := []struct {
		name       string
		want       error
		challenges []ChallengesData
	}{
		{
			name:       "no challenges",
			want:       fmt.Errorf("failed to get challenges"),
			challenges: []ChallengesData{},
		},
	}

	// setup mux
	client, mux, cleanup := setup()
	defer cleanup()

	// mock response
	mux.HandleFunc("/api/v1/challenges", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			Data    []ChallengesData `json:"data"`
			Success bool             `json:"success"`
		}{
			Data:    []ChallengesData{},
			Success: false,
		})
	})

	// test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.ListChallenges()
			if err == nil {
				t.Errorf("expected error, got nil")
				return
			}
		})
	}
}

func TestListChallenges_InvalidJSON(t *testing.T) {
	// setup mux
	client, mux, cleanup := setup()
	defer cleanup()

	// mock response
	mux.HandleFunc("/api/v1/challenges", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	})

	// test
	_, err := client.ListChallenges()
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
}

/* func TestListChallenges_InvalidResponse(t *testing.T) {
	// setup mux
	client, mux, cleanup := setup()
	defer cleanup()

	// mock response
	mux.HandleFunc("/api/v1/challenges/nope", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{}"))
	})

	// test
	_, err := client.ListChallenges()
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
} */
