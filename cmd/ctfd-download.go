package cmd

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ritchies/ctftool/internal/lib"
	"github.com/ritchies/ctftool/pkg/ctfd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ctfdDownloadCmd represents the download command
var ctfdDownloadCmd = &cobra.Command{
	Use:     "download",
	Aliases: []string{"d", "down"},
	Short:   "Download files and create writeups",
	Long:    `Download files and create writeups for each challenge.`,
	Run:     runDownload,
}

func runDownload(cmd *cobra.Command, args []string) {
	client := ctfd.NewClient()
	downloadOptions()
	opts.Output = setupOutputFolder()

	client.BaseURL = getBaseURL(cmd)
	client.Creds = getCredentials(cmd)
	client.MaxFileSize = options.MaxFileSize

	if !opts.SkipCTFDCheck {
		CheckErr(ctfd.Check())
	}

	if (opts.Username != "" || opts.Password != "") && opts.Token == "" {
		err := ctfd.Authenticate()
		CheckErr(err)
		log.Infof("Authenticated as %q", opts.Username)
	}

	processChallenges()

	if opts.SaveConfig {
		saveConfig()
	}

	if opts.Watch {
		watch(processChallenges)
	}
}

func processChallenges() {
	rl := GetRateLimit()
	var wg sync.WaitGroup

	// List challenges
	challenges, err := ctfd.ListChallenges()
	CheckErr(err)

	// Generate Index
	err = ctfd.GenerateIndex(challenges, opts.Output)
	CheckErr(err)

	// sort challenges so unsolved are first, otherwise sort by category
	sortFunc := func(i, j int) bool {
		if challenges[i].SolvedByMe != challenges[j].SolvedByMe {
			return !challenges[i].SolvedByMe
		}

		if challenges[i].Category != challenges[j].Category {
			return challenges[i].Category < challenges[j].Category
		}

		return challenges[i].Name < challenges[j].Name
	}

	sort.Slice(challenges, sortFunc)

	for _, challenge := range challenges {
		if opts.UnsolvedOnly && challenge.SolvedByMe {
			log.Debugf("Skipping %d : already solved", challenge.ID)
			continue
		}

		name := lib.CleanSlug(challenge.Name, false)
		category := strings.Split(challenge.Category, " ")[0]
		category = lib.CleanSlug(category, true)

		if len(category) < 1 || len(name) < 1 {
			log.Debugf("Skipping %d : invalid name or category", challenge.ID)
			continue
		}

		challengePath := path.Join(opts.Output, category, name)

		if _, statErr := os.Stat(challengePath); statErr == nil {
			if !opts.Overwrite {
				log.Debugf("Skipping %d : overwrite is false", challenge.ID)
				continue
			}
		}

		wg.Add(1)

		if options.RateLimit > 0 {
			rl.Take()
		}

		go func(challenge ctfd.ChallengesData) {
			log.WithField("challenge", fmt.Sprintf("%s/%s", category, name)).Infof("Downloading challenge %d", challenge.ID)

			chall, err := ctfd.Challenge(challenge.ID)
			CheckWarn(err)
			if err != nil {
				wg.Done()
				return
			}

			err = os.MkdirAll(challengePath, os.ModePerm)
			CheckErr(err)

			// get description
			err = ctfd.GetDescription(chall, challengePath)
			CheckErr(err)

			// download challenge files
			err = ctfd.DownloadFiles(chall.Files, challengePath)
			CheckWarn(err)

			wg.Done()
		}(challenge)
	}

	wg.Wait()
}

func watch(processFunc func()) {
	interval, err := time.ParseDuration(opts.WatchInterval.String())
	CheckErr(err)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		processFunc()
	}
}

func init() {
	ctfdCmd.AddCommand(ctfdDownloadCmd)

	ctfdDownloadCmd.Flags().StringVarP(&opts.URL, "url", "", "", "CTFd URL")
	ctfdDownloadCmd.Flags().StringVarP(&opts.Username, "username", "u", "", "CTFd Username")
	ctfdDownloadCmd.Flags().StringVarP(&opts.Password, "password", "p", "", "CTFd Password")
	ctfdDownloadCmd.Flags().StringVarP(&opts.Token, "token", "t", "", "CTFd Token")
	ctfdDownloadCmd.Flags().BoolVarP(&opts.Watch, "watch", "w", false, "Watch for new challenges")
	ctfdDownloadCmd.Flags().DurationVarP(&opts.WatchInterval, "watch-interval", "", 5*time.Minute, "Watch interval")
	ctfdDownloadCmd.Flags().BoolVarP(&opts.UnsolvedOnly, "unsolved", "", false, "Only download unsolved challenges")

	// viper
	err := viper.BindPFlag("url", ctfdDownloadCmd.Flags().Lookup("url"))
	CheckErr(err)

	err = viper.BindPFlag("username", ctfdDownloadCmd.Flags().Lookup("username"))
	CheckErr(err)

	err = viper.BindPFlag("password", ctfdDownloadCmd.Flags().Lookup("password"))
	CheckErr(err)

	err = viper.BindPFlag("token", ctfdDownloadCmd.Flags().Lookup("token"))
	CheckErr(err)

	err = viper.BindPFlag("watch", ctfdDownloadCmd.Flags().Lookup("watch"))
	CheckErr(err)

	err = viper.BindPFlag("watch-interval", ctfdDownloadCmd.Flags().Lookup("watch-interval"))
	CheckErr(err)

	err = viper.BindPFlag("unsolved", ctfdDownloadCmd.Flags().Lookup("unsolved"))
	CheckErr(err)
}
