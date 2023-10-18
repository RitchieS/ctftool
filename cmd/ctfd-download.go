package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/ritchies/ctftool/internal/lib"
	"github.com/ritchies/ctftool/pkg/ctfd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ChallengeNotifications struct {
	Total      int
	Categories map[string]int
}

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

	// Setup challenge notifications
	notifications := ChallengeNotifications{
		Categories: make(map[string]int),
	}

	for _, challenge := range SortChallenges(challenges) {
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

			// Add challenge to notifications, probably should make sure to lock?
			notifications.Total++
			notifications.Categories[category]++

			wg.Done()
		}(challenge)
	}

	wg.Wait()

	// Generate Index
	err = ctfd.GenerateIndex(challenges, opts.Output)
	CheckErr(err)

	if opts.Notify && notifications.Total > 0 {
		builder := strings.Builder{}
		builder.WriteString(fmt.Sprintf("Downloaded %d challenges\n", notifications.Total))

		for category, count := range notifications.Categories {
			builder.WriteString(fmt.Sprintf("\n%s: %d", category, count))
		}

		err := lib.SendNotification("CTFTool", builder.String())
		CheckWarn(err)
	}
}

func watch(processFunc func()) {
	interval, err := time.ParseDuration(opts.WatchInterval.String())
	CheckErr(err)

	log.Infof("Watching for new challenges every %s", interval.String())

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		log.Debugf("Checking for new challenges")
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
	ctfdDownloadCmd.Flags().BoolVarP(&opts.Notify, "notify", "", false, "Send desktop notifications")

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

	err = viper.BindPFlag("notify", ctfdDownloadCmd.Flags().Lookup("notify"))
	CheckErr(err)
}
