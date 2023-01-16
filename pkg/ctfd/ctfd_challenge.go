package ctfd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/ritchies/ctftool/pkg/scraper"
	"golang.org/x/net/html"
)

const (
	// Constants for max file sizes (1mb, 5mb, 25mb, 100mb, 0)
	NoFileSizeLimit    = 0
	OneMB              = 1000000
	FiveMB             = 5000000
	TwentyFiveMB       = 25000000
	OneHundredMB       = 100000000
	TwoHhundredFiftyMB = 250000000
)

type Hint struct {
	ID      int64  `json:"id"`
	Cost    int64  `json:"cost"`
	Content string `json:"content"`
}

type ChallengeData struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	ConnectionInfo string `json:"connection_info"`
	NextID         int64  `json:"next_id"`
	Attempts       int64  `json:"attempts"`
	MaxAttempts    int64  `json:"max_attempts"`
	Value          int64  `json:"value"`
	Category       string `json:"category"`
	Type           string `json:"type"`
	TypeData       struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		Templates struct {
			Create string `json:"create"`
			Update string `json:"update"`
			View   string `json:"view"`
		} `json:"templates"`
		Scripts struct {
			Create string `json:"create"`
			Update string `json:"update"`
			View   string `json:"view"`
		} `json:"scripts"`
	} `json:"type_data"`
	State      string        `json:"state"`
	Solves     int64         `json:"solves"`
	SolvedByMe bool          `json:"solved_by_me"`
	Files      []string      `json:"files"`
	Hints      []Hint        `json:"hints"`
	Tags       []interface{} `json:"tags"`
}

// Challenge returns a challenge by ID
func Challenge(id int64) (*ChallengeData, error) {
	response := new(struct {
		Success bool          `json:"success"`
		Data    ChallengeData `json:"data"`
	})

	resp, err := c.GetJson(fmt.Sprintf("api/v1/challenges/%d", id))
	if err != nil {
		return nil, fmt.Errorf("failed to get challenge: %v", err)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return nil, fmt.Errorf("failed to decode challenge: %v", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("failed to get challenge from %q", resp.Request.URL)
	}

	return &response.Data, nil
}

// DownloadFiles will download all the files of a challenge by ID and save
// them to the given directory
func DownloadFiles(id int64, outputPath string) error {
	challenge, err := Challenge(id)
	if err != nil {
		return err
	}

	files := make([]string, len(challenge.Files))
	copy(files, challenge.Files)

	// if no files, return
	if len(files) == 0 {
		return nil
	}

	for _, file := range files {
		resp, err := c.GetJson(file)
		if err != nil {
			return fmt.Errorf("failed to get file: %v", err)
		}
		defer resp.Body.Close()

		fileName := getFileName(resp.Request.URL.String())

		// 5 retries to get the challenge if the status code is not http.StatusOK
		for i := 0; i < 5; i++ {
			if resp.StatusCode == http.StatusOK {
				break
			}
			resp, err = c.GetJson(file)
			if err != nil {
				return fmt.Errorf("failed to get file: %v", err)
			}
			defer resp.Body.Close()

			time.Sleep(time.Second * 1)
		}

		if resp.ContentLength > (c.MaxFileSize*OneMB) || resp.ContentLength <= 0 {
			sizeInMegaBytes := resp.ContentLength / OneMB
			return fmt.Errorf("file %q is too large (%d/%d MB)", fileName, sizeInMegaBytes, c.MaxFileSize)
		}

		file, err := os.Create(path.Join(outputPath, fileName))
		if err != nil {
			return fmt.Errorf("failed to create file: %v", err)
		}
		defer file.Close()

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			return fmt.Errorf("failed to copy file: %v", err)
		}
	}

	return nil
}

// GetDescription retrieves a challenge and returns a writeup template of the challenge
func GetDescription(challenge *ChallengeData, challengePath string) error {
	challengePath = path.Join(challengePath, "README.md")

	var oldWriteupText []string
	// Check the file exists, if it does, then we need to extract everything after "## Writeup\n"
	if _, err := os.Stat(challengePath); err == nil {
		oldChallenge, err := os.Open(challengePath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to open challenge file: %v", err)
		}
		defer oldChallenge.Close()

		oldChallengeString, err := io.ReadAll(oldChallenge)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read challenge file: %v", err)
		}

		oldWriteupText = strings.Split(string(oldChallengeString), "## Writeup\n")
	}

	file, err := os.Create(challengePath)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	var solved string
	// Check if solved by me
	if challenge.SolvedByMe {
		solved = "✅"
	} else {
		solved = "❌"
	}

	_, err = file.WriteString(fmt.Sprintf("# %s %s - %s\n\n", solved, strings.ToUpper(challenge.Category), challenge.Name))
	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	// tags (if available)
	if len(challenge.Tags) > 0 {
		_, err = file.WriteString("## Tags\n\n")
		if err != nil {
			return fmt.Errorf("error writing to file: %v", err)
		}
		for _, tag := range challenge.Tags {
			_, err = file.WriteString(fmt.Sprintf("- %s\n", tag))
			if err != nil {
				return fmt.Errorf("error writing to file: %v", err)
			}
		}

		_, err = file.WriteString("\n")
		if err != nil {
			return fmt.Errorf("error writing to file: %v", err)
		}
	}

	// connection info (if available)
	if challenge.ConnectionInfo != "" {
		_, err := file.WriteString(fmt.Sprintf("Connection Info: %s\n\n", challenge.ConnectionInfo))
		if err != nil {
			return fmt.Errorf("error writing to file: %v", err)
		}
	}

	// files (if available)
	if len(challenge.Files) > 0 {
		_, err := file.WriteString("Files:\n\n")
		if err != nil {
			return fmt.Errorf("error writing to file: %v", err)
		}
		for _, challengeFile := range challenge.Files {
			fileURL, _ := c.BaseURL.Parse(challengeFile)
			_, err := file.WriteString(fmt.Sprintf("- [%s](%s)\n", getFileName(challengeFile), fileURL.String()))
			if err != nil {
				return fmt.Errorf("error writing to file: %v", err)
			}
		}

		_, err = file.WriteString("\n")
		if err != nil {
			return fmt.Errorf("error writing to file: %v", err)
		}
	}

	// description
	cleanDescription := func(desc string) string {
		parser := strings.NewReader(desc)
		decoder := html.NewTokenizer(parser)
		var text string
		for {
			tt := decoder.Next()
			switch tt {
			case html.ErrorToken:
				return text

			// img tags
			case html.SelfClosingTagToken:
				token := decoder.Token()
				if token.Data == "img" {
					text += fmt.Sprintf("![%s](%s)\n", token.String(), token.Attr[0].Val)
				}

			case html.StartTagToken:
				token := decoder.Token()

				// img tags
				if token.Data == "img" {
					fileName := getFileName(token.Attr[0].Val)
					text += fmt.Sprintf("![%s](%s)\n", fileName, token.Attr[0].Val)
				}

				// code block
				if token.Data == "code" {
					text += "`"
				}
			case html.EndTagToken:
				token := decoder.Token()

				// code block
				if token.Data == "code" {
					text += "`"
				}

			case html.TextToken:
				text += decoder.Token().String()
			}
		}

	}(challenge.Description)

	_, err = file.WriteString("## Description\n\n")
	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	// trip leading and trailing newlines
	description := strings.TrimSpace(cleanDescription)

	// replace multiple newlines and \r\n
	newlineRegex := regexp.MustCompile(`\n{2,}|\r\n`)
	description = newlineRegex.ReplaceAllString(description, "\n\n")

	// remove html entities
	description = html.UnescapeString(description)

	_, err = file.WriteString(fmt.Sprintf("%s\n", description))
	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	// hints (if available)
	if len(challenge.Hints) > 0 && challenge.Hints[0].Content != "" {
		_, err = file.WriteString("## Hints\n")
		if err != nil {
			return fmt.Errorf("error writing to file: %v", err)
		}
		for _, hint := range challenge.Hints {
			if hint.Content != "" {
				_, err = file.WriteString(fmt.Sprintf("- %s\n", hint.Content))
				if err != nil {
					return fmt.Errorf("error writing to file: %v", err)
				}
			}
		}

		_, err = file.WriteString("\n")
		if err != nil {
			return fmt.Errorf("error writing to file: %v", err)
		}
	}

	// writeup
	_, err = file.WriteString("\n## Writeup\n")
	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	if len(oldWriteupText) > 1 {
		_, err = file.WriteString(oldWriteupText[1])
		if err != nil {
			return fmt.Errorf("error writing to file: %v", err)
		}
	}

	return nil
}

type Submission struct {
	ID   int    `json:"challenge_id"`
	Flag string `json:"submission"`
}

func SubmitFlag(submission Submission) error {
	resp, err := c.GetJson("challenges")
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	nonce := scraper.ExtractCSRF(resp)

	c.BaseURL, _ = c.BaseURL.Parse("api/v1/challenges/attempt")

	data, err := json.Marshal(submission)
	if err != nil {
		return fmt.Errorf("failed to marshal submission: %v", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL.String(), bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Csrf-Token", nonce)
	req.Header.Set("Content-Type", "application/json")

	resp, err = c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("received status code %d (%s)", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	response := new(struct {
		Success bool `json:"success"`
		Data    struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		} `json:"data"`
	})

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	err = json.Unmarshal(data, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if !response.Success {
		return fmt.Errorf("failed to submit flag: %s", response.Data.Message)
	}

	c.BaseURL.Path = ""

	// check the challenge id and if we actually solved it
	challenge, err := Challenge(int64(submission.ID))
	if err != nil {
		return fmt.Errorf("failed to get challenge: %v", err)
	}

	if !challenge.SolvedByMe {
		return fmt.Errorf("failed to submit flag: %v", response.Data.Message)
	}

	return nil
}

// getFileName returns the file name from a URL path like
// /files/challenge.zip?token=12345
func getFileName(challengeFileURL string) string {
	directories := strings.Split(challengeFileURL, "/")
	challengeFile := strings.Split(directories[len(directories)-1], "?")

	fileName := challengeFile[0]
	return fileName
}
