package cmd

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/ritchies/ctftool/internal/lib"
	"github.com/ritchies/ctftool/pkg/ctfd"
	"github.com/ritchies/ctftool/pkg/scraper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ctfdDownloadCmd represents the download command
var ctfdDownloadCmd = &cobra.Command{
	Use:     "download",
	Aliases: []string{"d", "down"},
	Short:   "Download files and create writeups",
	Long:    `Download files and create writeups for each challenge.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := ctfd.NewClient()

		// check if flags are set using viper
		opts.URL = viper.GetString("url")
		opts.Username = viper.GetString("username")
		opts.Password = viper.GetString("password")
		opts.Output = viper.GetString("output")
		opts.Overwrite = viper.GetBool("overwrite")
		opts.SaveConfig = viper.GetBool("save-config")
		opts.SkipCTFDCheck = viper.GetBool("skip-check")
		opts.Watch = viper.GetBool("watch")
		opts.WatchInterval = viper.GetDuration("watch-interval")
		opts.UnsolvedOnly = viper.GetBool("unsolved")

		baseURL, err := url.Parse(opts.URL)
		CheckErr(err)

		if baseURL.Host == "" {
			ShowHelp(cmd, fmt.Sprintf("Invalid or empty URL provided: %q", baseURL.String()))
		}

		client.BaseURL = baseURL

		if opts.Username != "" && opts.Password == "" {
			fmt.Print("Enter your password: ")
			var password string
			fmt.Scanln(&password)
			opts.Password = strings.TrimSpace(password)
		}

		// opts.Username and password are required
		if opts.Username == "" || opts.Password == "" {
			ShowHelp(cmd, "CTFD User and Password are required")
		}

		credentials := scraper.Credentials{
			Username: opts.Username,
			Password: opts.Password,
		}

		client.Creds = &credentials
		client.MaxFileSize = options.MaxFileSize

		if !opts.SkipCTFDCheck {
			err = ctfd.Check()
			CheckErr(err)
		}

		err = ctfd.Authenticate()
		CheckErr(err)

		log.Infof("Authenticated as %q", opts.Username)

		cwd, err := os.Getwd()
		CheckErr(err)

		outputFolder := cwd

		// if using config file
		if viper.ConfigFileUsed() == "" && opts.Output != "" {
			outputFolder = path.Join(cwd, opts.Output)
		}

		processChallenges := func() {
			rl := GetRateLimit()
			var wg sync.WaitGroup

			// List challenges
			challenges, err := ctfd.ListChallenges()
			CheckErr(err)

			// Warn the user that they are about to overwrite files
			if opts.Overwrite {
				log.Warn("This action will overwrite existing files")
				log.Info("Writeups will be updated if they exist")
				log.Info("Press enter to continue or ctrl+c to cancel")

				// Ask the user if they want to continue (default is yes)
				fmt.Print("Do you want to continue? [Y/n]: ")
				var answer string
				fmt.Scanln(&answer)
				switch strings.ToLower(answer) {
				case "n", "no":
					log.Fatal("Aborting by user request")
				}
			}

			for _, challenge := range challenges {
				wg.Add(1)

				if options.RateLimit > 0 {
					rl.Take()
				}

				go func(challenge ctfd.ChallengesData) {
					name := lib.CleanSlug(challenge.Name, false)

					category := strings.Split(challenge.Category, " ")[0]
					category = lib.CleanSlug(category, true)

					// make sure name and category are more than 1 character and less than 50
					if len(category) < 1 || len(name) < 1 {
						log.Warnf("Skipping (%q/%q) : invalid name or category", challenge.Name, challenge.Category)
						wg.Done()
						return
					}

					if opts.UnsolvedOnly && challenge.SolvedByMe {
						log.Warnf("Skipping (%q/%q) : already solved", challenge.Name, challenge.Category)
						wg.Done()
						return
					}

					challengePath := path.Join(outputFolder, category, name)

					if _, statErr := os.Stat(challengePath); statErr == nil {
						if !opts.Overwrite {
							log.Warnf("Skipping %q : overwrite is false", name)
							wg.Done()
							return
						}
					}

					err := os.MkdirAll(challengePath, os.ModePerm)
					CheckErr(err)

					chall, err := ctfd.Challenge(challenge.ID)
					CheckErr(err)

					// download challenge files
					err = ctfd.DownloadFiles(chall.ID, challengePath)
					CheckWarn(err)

					// get description
					err = ctfd.GetDescription(chall, challengePath)
					CheckErr(err)

					if len(chall.Files) > 0 {
						log.WithFields(logrus.Fields{
							"category": category,
							"files":    len(chall.Files),
							"solves":   chall.Solves,
						}).Infof("Downloaded %q", name)
					} else {
						log.WithFields(logrus.Fields{
							"category": category,
							"solves":   chall.Solves,
						}).Infof("Created %q", name)
					}

					wg.Done()
				}(challenge)
			}
			wg.Wait()
		}

		processChallenges()

		// values to config file if --save-config is set
		if opts.SaveConfig {
			viper.Set("url", opts.URL)
			viper.Set("username", opts.Username)
			viper.Set("password", "")
			viper.Set("output", outputFolder)
			viper.Set("overwrite", true)
			err := viper.SafeWriteConfigAs(path.Join(outputFolder, ".ctftool.yaml"))
			CheckErr(err)

			log.WithField("file", path.Join(outputFolder, ".ctftool.yaml")).Info("Saved config file")
			log.Info("You can now run `ctftool` from the same directory without specifying the --url, --username and --password in that directory")
		}

		if opts.Watch {
			log.Infof("Watching for new challenges every %s", opts.WatchInterval.String())

			interval, err := time.ParseDuration(opts.WatchInterval.String())
			CheckErr(err)

			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			for range ticker.C {
				processChallenges()
			}
		}
	},
}

func init() {
	ctfdCmd.AddCommand(ctfdDownloadCmd)

	ctfdDownloadCmd.Flags().StringVarP(&opts.URL, "url", "", "", "CTFd URL")
	ctfdDownloadCmd.Flags().StringVarP(&opts.Username, "username", "u", "", "CTFd Username")
	ctfdDownloadCmd.Flags().StringVarP(&opts.Password, "password", "p", "", "CTFd Password")
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

	err = viper.BindPFlag("watch", ctfdDownloadCmd.Flags().Lookup("watch"))
	CheckErr(err)

	err = viper.BindPFlag("watch-interval", ctfdDownloadCmd.Flags().Lookup("watch-interval"))
	CheckErr(err)

	err = viper.BindPFlag("unsolved", ctfdDownloadCmd.Flags().Lookup("unsolved"))
	CheckErr(err)
}
