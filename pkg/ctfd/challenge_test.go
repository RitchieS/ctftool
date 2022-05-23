package ctfd

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestGetChallenge(t *testing.T) {

	// return success
	challengex := new(struct {
		Success bool      `json:"success"`
		Data    Challenge `json:"data"`
	})
	challengex.Success = true
	challengex.Data = Challenge{
		ID:             1,
		Name:           "test challenge",
		Description:    "test description",
		ConnectionInfo: "test connection info",
		Attempts:       0,
		MaxAttempts:    1,
		Value:          42,
		Category:       "file",
		Type:           "file",
		State:          "active",
		Solves:         1337,
		SolvedByMe:     false,
		Files:          []string{"test.txt"},
		Hints: []Hint{
			{
				ID:      1,
				Cost:    1,
				Content: "test hint",
			},
		},
		Tags: []interface{}{"test tag"},
	}

	// setup mux
	client, mux, cleanup := setup()
	defer cleanup()

	// mock request
	mux.HandleFunc("/api/v1/challenges/1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(challengex)
	})

	// test
	challenge, err := client.Challenge(1)
	if err != nil {
		t.Errorf("error getting challenge: %v", err)
		return
	}

	if challenge.ID != 1 {
		t.Errorf("expected challenge ID 1, got %d", challenge.ID)
		return
	}

	if challenge.Name != "test challenge" {
		t.Errorf("expected challenge name 'test challenge', got %s", challenge.Name)
		return
	}

	if challenge.Description != "test description" {
		t.Errorf("expected challenge description 'test challenge description', got %s", challenge.Description)
		return
	}

	if challenge.Category != "file" {
		t.Errorf("expected challenge category 'file', got %s", challenge.Category)
		return
	}

	if challenge.Value != 42 {
		t.Errorf("expected challenge value 42, got %d", challenge.Value)
		return
	}

	if challenge.Solves != 1337 {
		t.Errorf("expected challenge solves 1337, got %d", challenge.Solves)
		return
	}

	if challenge.SolvedByMe != false {
		t.Errorf("expected challenge solved by me false, got %t", challenge.SolvedByMe)
		return
	}

	if challenge.Type != "file" {
		t.Errorf("expected challenge type 'file', got %s", challenge.Type)
		return
	}

	// skip typedata tests

	if len(challenge.Files) != 1 {
		t.Errorf("expected 1 file, got %d", len(challenge.Files))
		return
	}

	if challenge.Files[0] != "test.txt" {
		t.Errorf("expected file 'test.txt', got %s", challenge.Files[0])
		return
	}

	if len(challenge.Hints) != 1 {
		t.Errorf("expected 1 hint, got %d", len(challenge.Hints))
		return
	}

	if challenge.Hints[0].ID != 1 {
		t.Errorf("expected hint ID 1, got %d", challenge.Hints[0].ID)
		return
	}
}

// fail tests
func TestGetChallengeFail(t *testing.T) {

	// return success
	challengex := new(struct {
		Success bool      `json:"success"`
		Data    Challenge `json:"data"`
	})

	challengex.Success = false
	challengex.Data = Challenge{}

	// setup mux
	client, mux, cleanup := setup()
	defer cleanup()

	// mock request
	mux.HandleFunc("/api/v1/challenges/1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(challengex)
	})

	// test
	_, err := client.Challenge(1)
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	if err.Error() != "failed to get challenge" {
		t.Errorf("expected error 'challenge not found', got %s", err.Error())
		return
	}

	//// json fail for challenge 2
	challengex.Success = true
	challengex.Data = Challenge{
		ID:             2,
		Name:           "test challenge",
		Description:    "test description",
		ConnectionInfo: "test connection info",
		Attempts:       0,
		MaxAttempts:    1,
		Value:          42,
		Category:       "file",
		Type:           "file",
		State:          "active",
		Solves:         1337,
		SolvedByMe:     false,
		Files:          []string{"test.txt"},
		Hints: []Hint{
			{
				ID:      1,
				Cost:    1,
				Content: "test hint",
			},
		},
		Tags: []interface{}{"test tag"},
	}

	// mock request
	mux.HandleFunc("/api/v1/challenges/2", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		//json.NewEncoder(w).Encode(challengex)
		w.Write([]byte(`{"success":false,"data":{derp}}`))
	})

	// test
	_, err = client.Challenge(2)
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	//// give teapot status code
	challengex.Success = true

	// mock request
	mux.HandleFunc("/api/v1/challenges/3", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte(`{"success":false,"data":{}}`))
	})

	// test
	_, err = client.Challenge(3)
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	if err.Error() != "error fetching challenge: received 418 status code" {
		t.Errorf("expected 'error fetching challenge: received 418 status code', got %q", err.Error())
		return
	}

	//// challenge 4 does not exist :d
	challengex.Success = true
	challengex.Data = Challenge{}

	// mock request

	// test
	_, err = client.Challenge(4)
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

}

// Test getFileName(challengeFileURL string)
func TestGetFileName(t *testing.T) {
	exampleUrls := [][]string{
		{"https://www.google.com/test.txt", "test.txt"},
		{"https://www.google.com/test.txt?foo=bar", "test.txt"},
	}

	for _, example := range exampleUrls {
		if getFileName(example[0]) != example[1] {
			t.Errorf("expected %s, got %s", example[1], getFileName(example[0]))
			return
		}
	}
}
