package cmd

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/ritchies/ctftool/pkg/ctfd"
	"github.com/ritchies/ctftool/pkg/scraper"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	opts = ctfd.NewOpts()
)

// ctfdCmd represents the ctfd command
var ctfdCmd = &cobra.Command{
	Use:   "ctfd",
	Short: "Query CTFd instance",
	Long:  `Retrieve challenges and files from a CTFd instance.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		CheckErr(err)
	},
}

func init() {
	rootCmd.AddCommand(ctfdCmd)

	ctfdCmd.PersistentFlags().BoolVarP(&opts.SaveConfig, "save-config", "", false, "Save config to (default is $OUTDIR/.ctftool.yaml)")

	// viper
	err := viper.BindPFlag("save-config", ctfdCmd.PersistentFlags().Lookup("save-config"))
	CheckErr(err)
}

func ctfdOptions() {
	opts.URL = viper.GetString("url")
	opts.Username = viper.GetString("username")
	opts.Password = viper.GetString("password")
	opts.Token = viper.GetString("token")
	opts.Output = viper.GetString("output")
	opts.Overwrite = viper.GetBool("overwrite")
	opts.SaveConfig = viper.GetBool("save-config")
	opts.SkipCTFDCheck = viper.GetBool("skip-check")
	opts.Watch = viper.GetBool("watch")
	opts.WatchInterval = viper.GetDuration("watch-interval")
	opts.UnsolvedOnly = viper.GetBool("unsolved")
	opts.Notify = viper.GetBool("notify")
	opts.MaxFileSize = viper.GetInt64("max-file-size")
	options.RateLimit = viper.GetInt("rate-limit")
}

func getBaseURL(cmd *cobra.Command) *url.URL {
	baseURL, err := url.Parse(opts.URL)
	CheckErr(err)

	if baseURL.Host == "" {
		ShowHelp(cmd, fmt.Sprintf("Invalid or empty URL provided: %q", baseURL.String()))
	}

	return baseURL
}

func getCredentials(cmd *cobra.Command) *scraper.Credentials {
	if opts.Username != "" && opts.Password == "" {
		fmt.Print("Enter your password: ")
		var password string
		fmt.Scanln(&password)
		opts.Password = strings.TrimSpace(password)
	}

	if (opts.Username == "" || opts.Password == "") && opts.Token == "" {
		ShowHelp(cmd, "Either CTFD Username and Password or a Token are required")
	}

	return &scraper.Credentials{
		Username: opts.Username,
		Password: opts.Password,
		Token:    opts.Token,
	}
}

func setupOutputFolder() string {
	cwd, err := os.Getwd()
	CheckErr(err)
	outputFolder := cwd

	if viper.ConfigFileUsed() == "" && opts.Output != "" {
		outputFolder = path.Join(cwd, opts.Output)
	}
	return outputFolder
}

func saveConfig() {
	viper.Reset()

	if opts.URL != "" {
		viper.Set("url", opts.URL)
	}
	if opts.Username != "" {
		viper.Set("username", opts.Username)
	}
	if opts.Password != "" {
		viper.Set("password", opts.Password)
	}
	if opts.Token != "" {
		viper.Set("token", opts.Token)
	}
	if opts.Output != "" {
		viper.Set("output", opts.Output)
	}
	if opts.Overwrite {
		viper.Set("overwrite", opts.Overwrite)
	}
	if opts.SkipCTFDCheck {
		viper.Set("skip-check", opts.SkipCTFDCheck)
	}
	if opts.UnsolvedOnly {
		viper.Set("unsolved", opts.UnsolvedOnly)
	}
	if opts.Watch {
		viper.Set("watch", opts.Watch)
	}
	if opts.WatchInterval != 0 {
		viper.Set("watch-interval", opts.WatchInterval)
	}
	if opts.MaxFileSize != 0 {
		viper.Set("max-file-size", opts.MaxFileSize)
	}
	if options.RateLimit != 0 {
		viper.Set("rate-limit", options.RateLimit)
	}

	err := viper.WriteConfigAs(path.Join(opts.Output, ".ctftool.yaml"))
	CheckErr(err)

	log.WithField("file", path.Join(opts.Output, ".ctftool.yaml")).Info("Saved config file")
	log.Info("You can now run ctftool without any arguments")
}

// sort challenges modifies the order of the challenges slice
func SortChallenges(challenges []ctfd.ChallengesData) []ctfd.ChallengesData {
	// sort priority:
	// - challenges with zero solves (and that reward points)
	// - challenges with solves, not solved by me
	// - challenges with solves, not solved by me, sorted by solves (lowest first)
	// - challenges with solves, solved by me, sorted by solves (lowest first)

	sortFunc := func(i, j int) bool {
		// Challenges with zero solves come first
		if challenges[i].Solves == 0 && challenges[j].Solves == 0 {
			return challenges[i].ID < challenges[j].ID
		}

		// If one has zero solves and the other doesn't, the one with zero solves comes first
		if challenges[i].Solves == 0 && challenges[j].Solves > 0 {
			return true
		}

		if challenges[i].Solves > 0 && challenges[j].Solves == 0 {
			return false
		}

		// Both have more than zero solves
		if challenges[i].Solves > 0 && challenges[j].Solves > 0 {
			// Both are solved by me, sort by number of solves (lowest first)
			if challenges[i].SolvedByMe && challenges[j].SolvedByMe {
				return challenges[i].Solves < challenges[j].Solves
			}

			// Neither are solved by me, sort by number of solves (lowest first)
			if !challenges[i].SolvedByMe && !challenges[j].SolvedByMe {
				return challenges[i].Solves < challenges[j].Solves
			}

			// One is solved by me and the other is not, the one not solved by me comes first
			if challenges[i].SolvedByMe && !challenges[j].SolvedByMe {
				return false
			}

			if !challenges[i].SolvedByMe && challenges[j].SolvedByMe {
				return true
			}
		}

		// Fallback to sorting by ID
		return challenges[i].ID < challenges[j].ID
	}

	sort.Slice(challenges, sortFunc)

	return challenges
}
