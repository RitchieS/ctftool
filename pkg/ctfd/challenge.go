package ctfd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

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

type Challenge struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	ConnectionInfo string `json:"connection_info"`
	// NextID       int64  `json:"next_id"`
	Attempts    int64  `json:"attempts"`
	MaxAttempts int64  `json:"max_attempts"`
	Value       int64  `json:"value"`
	Category    string `json:"category"`
	Type        string `json:"type"`
	TypeData    struct {
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
	State string `json:"state"`
	// Requirements []struct {
	// 	ID   int64  `json:"id"`
	// 	Name string `json:"name"`
	// } `json:"requirements"`
	Solves     int64         `json:"solves"`
	SolvedByMe bool          `json:"solved_by_me"`
	Files      []string      `json:"files"`
	Hints      []Hint        `json:"hints"`
	Tags       []interface{} `json:"tags"`
	// View
}

func (c *Client) Challenge(id int64) (*Challenge, error) {
	challenge := new(struct {
		Success bool      `json:"success"`
		Data    Challenge `json:"data"`
	})

	challengeAPI, err := joinPath(c.BaseURL.String(), "api/v1/challenges", fmt.Sprintf("%d", id))
	if err != nil {
		return nil, fmt.Errorf("error joining path: %v", err)
	}

	doc, err := c.Client.Get(challengeAPI.String())
	if err != nil {
		return nil, fmt.Errorf("error fetching challenge from %q: %v", challengeAPI.String(), err)
	}
	defer doc.Body.Close()

	// 5 retries to get the challenge if the status code is not http.StatusOK
	for i := 0; i < 5; i++ {
		if doc.StatusCode == http.StatusOK {
			break
		}
		doc, err = c.Client.Get(challengeAPI.String())
		if err != nil {
			return nil, fmt.Errorf("error fetching challenge from %q: %v", challengeAPI.String(), err)
		}
		defer doc.Body.Close()

		time.Sleep(time.Second * 1)
	}

	if doc.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error fetching challenge: received %v status code", doc.StatusCode)
	}

	err = json.NewDecoder(doc.Body).Decode(challenge)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling challenge from %q: %v", challengeAPI.String(), err)
	}

	if !challenge.Success {
		return nil, fmt.Errorf("failed to get challenge")
	}

	return &challenge.Data, nil
}

func (c *Client) DownloadFiles(id int64, outputPath string) error {
	challenge, err := c.Challenge(id)
	if err != nil {
		return fmt.Errorf("error getting challenge: %v", err)
	}

	files := make([]string, len(challenge.Files))
	copy(files, challenge.Files)

	// if no files, return
	if len(files) == 0 {
		return nil
	}

	for _, file := range files {
		challengeFileURL, _ := c.BaseURL.Parse(file)
		fileName := getFileName(challengeFileURL.String())

		resp, err := c.Client.Get(challengeFileURL.String())
		if err != nil {
			return fmt.Errorf("error getting challenge file: %v", err)
		}
		defer resp.Body.Close()

		// 5 retries to get the challenge if the status code is not http.StatusOK
		for i := 0; i < 5; i++ {
			if resp.StatusCode == http.StatusOK {
				break
			}
			resp, err = c.Client.Get(challengeFileURL.String())
			if err != nil {
				return fmt.Errorf("error getting challenge file: %v", err)
			}
			defer resp.Body.Close()

			time.Sleep(time.Second * 1)
		}

		file, err := os.Create(path.Join(outputPath, fileName))
		if err != nil {
			return fmt.Errorf("error creating file: %v", err)
		}
		defer file.Close()

		if resp.ContentLength > TwentyFiveMB || resp.ContentLength < 0 {
			sizeInMegaBytes := resp.ContentLength / OneMB
			return fmt.Errorf("file size is too big : %vmb", sizeInMegaBytes)
		}

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			return fmt.Errorf("error copying file: %v", err)
		}
	}

	time.Sleep(time.Millisecond * 333)

	return nil
}

// GetDescription retrieves a challenge and returns a writeup template of the challenge
func (c *Client) GetDescription(challenge *Challenge, challengePath string) error {
	challengePath = path.Join(challengePath, "README.md")

	var oldWriteupText []string
	// Check the file exists, if it does, then we need to extract everything after "## Writeup\n"
	if _, err := os.Stat(challengePath); err == nil {
		oldChallenge, err := os.Open(challengePath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("error opening challenge file: %v", err)
		}
		defer oldChallenge.Close()

		oldChallengeString, err := ioutil.ReadAll(oldChallenge)
		if err != nil && err != io.EOF {
			return fmt.Errorf("error reading challenge file: %v", err)
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

	file.WriteString(fmt.Sprintf("# %s %s - %s\n\n", solved, strings.ToUpper(challenge.Category), challenge.Name))

	// tags (if available)
	if len(challenge.Tags) > 0 {
		file.WriteString("## Tags\n\n")
		for _, tag := range challenge.Tags {
			file.WriteString(fmt.Sprintf("- %s\n", tag))
		}

		file.WriteString("\n")
	}

	// connection info (if available)
	if challenge.ConnectionInfo != "" {
		file.WriteString(fmt.Sprintf("Connection Info: %s\n\n", challenge.ConnectionInfo))
	}

	// files (if available)
	if len(challenge.Files) > 0 {
		file.WriteString("Files:\n\n")
		for _, challengeFile := range challenge.Files {
			fileURL, _ := c.BaseURL.Parse(challengeFile)
			file.WriteString(fmt.Sprintf("- [%s](%s)\n", getFileName(challengeFile), fileURL.String()))
		}

		file.WriteString("\n")
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

	file.WriteString("## Description\n\n")

	// trip leading and trailing newlines
	description := strings.TrimSpace(cleanDescription)

	// replace multiple newlines and \r\n
	newlineRegex := regexp.MustCompile(`\n{2,}|\r\n`)
	description = newlineRegex.ReplaceAllString(description, "\n\n")

	// remove html entities
	description = html.UnescapeString(description)

	file.WriteString(fmt.Sprintf("%s\n", description))

	// hints (if available)
	if len(challenge.Hints) > 0 && challenge.Hints[0].Content != "" {
		file.WriteString("## Hints\n")
		for _, hint := range challenge.Hints {
			if hint.Content != "" {
				file.WriteString(fmt.Sprintf("- %s\n", hint.Content))
			}
		}

		file.WriteString("\n")
	}

	// writeup
	file.WriteString("\n## Writeup\n")

	if len(oldWriteupText) > 1 {
		file.WriteString(oldWriteupText[1])
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
