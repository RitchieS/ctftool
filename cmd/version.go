package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var (
	Version = "dev"     // Version is the current version of the program
	Commit  = "none"    // Commit is the commit hash of the current build
	Date    = "unknown" // Date is the time the program was built
	BuiltBy = "unknown" // BuiltBy is how the program was built (unknown, goreleaser, etc)
)

const (
	user = "ritchies"
	repo = "ctftool"
)

type Response struct {
	Name    string `json:"name"`
	Zipball string `json:"zipball_url"`
	Tarball string `json:"tarball_url"`
	Commit  struct {
		Sha string `json:"sha"`
		URL string `json:"url"`
	} `json:"commit"`
	NodeID string `json:"node_id"`
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `All software has versions. This is mine.`,
	Run: func(cmd *cobra.Command, args []string) {
		if Version == "dev" {
			fmt.Printf("You are running a development build of ctftool\n")
		} else {
			fmt.Printf("ctftool %s (%s) built by %s on %s\n", Version, Commit, BuiltBy, Date)

			// get the latest tag and hash
			latest, hash := getLatestVersion()

			// compare commit hashes
			if strings.Contains(hash, Commit) {
				fmt.Printf("You are on the latest version\n")
				return
			}

			// compare the versions
			if compareVersions(Version, latest) {
				fmt.Printf("You are running an older version of ctftool\n")
				fmt.Println()
				fmt.Printf("Latest version: %s\n", latest)
				fmt.Printf("Update using: go install -v github.com/ritchies/ctftool@latest\n")
				// go install using the latest tag
				fmt.Printf("Update using: go install -v github.com/ritchies/ctftool@%s\n", latest)
			}

			// compare versions if newer
			if compareVersions(latest, Version) {
				fmt.Printf("You are running a newer version of ctftool\n")
				fmt.Println()
				fmt.Printf("%s is the latest version and you are running %s\n", latest, Version)
			}
		}
	},
}

func compareVersions(a, b string) bool {
	majorA, minorA, patchA := parseVersion(a)
	majorB, minorB, patchB := parseVersion(b)

	// sort by semver
	if majorA > majorB {
		return true
	}
	if majorA < majorB {
		return false
	}

	if minorA > minorB {
		return true
	}

	if minorA < minorB {
		return false
	}

	if patchA > patchB {
		return true
	}

	if patchA < patchB {
		return false
	}

	return false
}

func parseVersion(version string) (int, int, int) {

	// remove the v from the version
	version = strings.TrimPrefix(version, "v")

	// split the version on - and grab the first part
	if strings.Contains(version, "-") {
		version = strings.Split(version, "-")[0]
	}

	// split the version on .
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return 0, 0, 0
	}
	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	patch, _ := strconv.Atoi(parts[2])
	return major, minor, patch
}

func getLatestVersion() (string, string) {
	uri := fmt.Sprintf("https://api.github.com/repos/%s/%s/tags", user, repo)
	resp, err := http.Get(uri)
	if err != nil {
		log.Fatalf("Error checking for updates: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalf("Error checking for updates: %s", resp.Status)
	}

	// Check for updates (if we're not on the latest version)
	var response []Response
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Fatalf("Error checking for updates: %s", err)
	}

	var tags []string

	for _, r := range response {
		tags = append(tags, r.Name)
	}

	// get the tags
	// sort by semver
	sort.Slice(tags, func(i, j int) bool {
		return compareVersions(tags[i], tags[j])
	})

	// get the latest version
	return tags[0], response[0].Commit.Sha
}
