package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/ritchies/ctftool/internal/lib"
	"github.com/ritchies/ctftool/pkg/ctfd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ctfdWriteupCmd represents the writeups command
var ctfdWriteupCmd = &cobra.Command{
	Use:     "writeups",
	Aliases: []string{"w", "write"},
	Short:   "Only create and update writeups",
	Long:    `Create and update writeups for each challenge. Skips downloading files.`,
	Run:     runWriteups,
}

func runWriteups(cmd *cobra.Command, args []string) {
	client := ctfd.NewClient()
	downloadOptions()
	opts.Output = setupOutputFolder()

	client.BaseURL = getBaseURL(cmd)
	client.Creds = getCredentials(cmd)

	if !opts.SkipCTFDCheck {
		CheckErr(ctfd.Check())
	}

	if (opts.Username != "" || opts.Password != "") && opts.Token == "" {
		err := ctfd.Authenticate()
		CheckErr(err)
		log.Infof("Authenticated as %q", opts.Username)
	}

	processWriteups()
}

func processWriteups() {
	// Similar to processChallenges but specific to writeups
	rl := GetRateLimit()
	var wg sync.WaitGroup

	// List challenges
	challenges, err := ctfd.ListChallenges()
	CheckErr(err)

	for _, challenge := range SortChallenges(challenges) {
		wg.Add(1)

		if options.RateLimit > 0 {
			rl.Take()
		}

		go func(challenge ctfd.ChallengesData) {
			name := lib.CleanSlug(challenge.Name, false)
			category := strings.Split(challenge.Category, " ")[0]
			category = lib.CleanSlug(category, true)

			if len(category) < 1 || len(name) < 1 {
				log.Debugf("Skipping challenge %d : invalid category or name", challenge.ID)
				wg.Done()
				return
			}

			log.WithField("challenge", fmt.Sprintf("%s/%s", category, name)).Infof("Processing challenge %d", challenge.ID)

			challengePath := path.Join(opts.Output, category, name)

			chall, err := ctfd.Challenge(challenge.ID)
			CheckErr(err)

			err = os.MkdirAll(challengePath, os.ModePerm)
			CheckErr(err)

			// get description
			err = ctfd.GetDescription(chall, challengePath)
			CheckErr(err)

			wg.Done()
		}(challenge)
	}

	wg.Wait()
}

func init() {
	ctfdCmd.AddCommand(ctfdWriteupCmd)

	ctfdWriteupCmd.Flags().StringVarP(&opts.URL, "url", "", "", "CTFd URL")
	ctfdWriteupCmd.Flags().StringVarP(&opts.Username, "username", "u", "", "CTFd Username")
	ctfdWriteupCmd.Flags().StringVarP(&opts.Password, "password", "p", "", "CTFd Password")
	ctfdWriteupCmd.Flags().StringVarP(&opts.Token, "token", "t", "", "CTFd Token")

	// viper
	err := viper.BindPFlag("url", ctfdWriteupCmd.Flags().Lookup("url"))
	CheckErr(err)

	err = viper.BindPFlag("username", ctfdWriteupCmd.Flags().Lookup("username"))
	CheckErr(err)

	err = viper.BindPFlag("password", ctfdWriteupCmd.Flags().Lookup("password"))
	CheckErr(err)

	err = viper.BindPFlag("token", ctfdWriteupCmd.Flags().Lookup("token"))
	CheckErr(err)
}
