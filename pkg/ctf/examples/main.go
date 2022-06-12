package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/ritchies/ctftool/pkg/ctf"
	"github.com/sirupsen/logrus"
)

var (
	log = logrus.New()
)

var (
	username  = flag.String("username", "", "ctfd auth: username")
	password  = flag.String("password", "", "ctfd auth: password")
	ctfdURL   = flag.String("url", "", "ctfd auth: url")
	outputDir = flag.String("output", "output", "output directory (default: output)")
	overwrite = flag.Bool("overwrite", false, "overwrite existing files")
)

func main() {
	// Logrus options
	log.SetFormatter(&logrus.TextFormatter{
		DisableSorting:         false,
		DisableTimestamp:       true,
		DisableLevelTruncation: true,
		ForceColors:            true,
		ForceQuote:             true,
		PadLevelText:           true,
		QuoteEmptyFields:       true,
	})

	flag.Parse()

	client := ctf.NewClient(nil)

	baseURL, err := url.Parse(*ctfdURL)
	if err != nil {
		log.Fatal(err)
	}

	client.BaseURL = baseURL

	reader := bufio.NewReader(os.Stdin)
	if *password == "" && *username != "" {
		fmt.Print("password: ")
		*password, _ = reader.ReadString('\n')
		*password = strings.TrimSpace(*password)
	}

	// username and password are required
	if *username == "" || *password == "" {
		flag.Usage()
		log.Fatal("url, username and password are required")
	}

	credentials := ctf.Credentials{
		Username: *username,
		Password: *password,
	}

	client.Creds = &credentials

	if err := client.Authenticate(); err != nil {
		log.Fatal(err)
	}
	log.Infof("Authenticated as %q", *username)

	// List challenges
	challenges, err := client.ListChallenges()
	if err != nil {
		log.Fatal(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("error getting current working directory: %v", err)
	}

	for _, challenge := range challenges {
		outputFolder := path.Join(cwd, *outputDir)

		category := strings.Split(challenge.Category, " ")[0]
		category = strings.ToLower(category)

		name := strings.Replace(challenge.Name, " ", "-", -1)
		name = regexp.MustCompile("[^a-zA-Z0-9-.]+").ReplaceAllString(name, "_")
		name = regexp.MustCompile("-_|_-|-_-|_-_").ReplaceAllString(name, "-")
		name = regexp.MustCompile("_+").ReplaceAllString(name, "_")
		name = strings.Trim(name, "-_")

		if len(category) > 50 {
			category = category[:50]
		}
		if len(name) > 50 {
			name = name[:50]
		}

		// make sure name and category are more than 1 character and less than 50
		if len(category) < 1 || len(name) < 1 {
			log.Warnf("skipping challenge %q because it has an invalid name or category", challenge.Name)
			continue
		}

		filePath := path.Join(outputFolder, category, name)

		if _, statErr := os.Stat(filePath); statErr == nil {
			log.Warnf("Challenge %q already exists", challenge.Name)
			// continue if overwrite is false
			if !*overwrite {
				log.Debugf("skipping challenge %q because overwrite is false", challenge.Name)
				continue
			} else {
				log.Debugf("overwriting challenge %q", challenge.Name)
			}
		}

		if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
			log.Fatal(fmt.Errorf("error creating directory: %v", err))
		}

		chall, err := client.Challenge(challenge.ID)
		if err != nil {
			log.Fatal(fmt.Errorf("error getting challenge: %v", err))
		}

		// download challenge files
		if err := client.DownloadFiles(chall.ID, filePath); err != nil {
			log.Fatal(fmt.Errorf("error downloading files: %v", err))
		}

		// get description
		if err := client.GetDescription(chall, filePath); err != nil {
			log.Fatal(fmt.Errorf("error getting description: %v", err))
		}
	}
}
