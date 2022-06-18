package cmd

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/ritchies/ctftool/internal/lib"
	"github.com/ritchies/ctftool/pkg/ctf"
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
		client := ctf.NewClient(nil)

		// check if flags are set using viper
		opts.URL = viper.GetString("url")
		opts.Username = viper.GetString("username")
		opts.Password = viper.GetString("password")
		opts.Output = viper.GetString("output")
		opts.Overwrite = viper.GetBool("overwrite")

		baseURL, err := url.Parse(opts.URL)
		CheckErr(err)

		if baseURL.Host == "" {
			cmd.Help()
			log.Fatalf("Invalid or empty URL provided: %s", baseURL.String())
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
			cmd.Help()
			log.Fatal("CTFD User and Password are required")
		}

		credentials := ctf.Credentials{
			Username: opts.Username,
			Password: opts.Password,
		}

		client.Creds = &credentials
		client.MaxFileSize = options.MaxFileSize

		err = client.Authenticate()
		CheckErr(err)

		log.Infof("Authenticated as %q", opts.Username)

		// List challenges
		challenges, err := client.ListChallenges()
		CheckErr(err)

		cwd, err := os.Getwd()
		CheckErr(err)

		outputFolder := cwd

		// if using config file
		if viper.ConfigFileUsed() == "" && opts.Output != "" {
			outputFolder = path.Join(cwd, opts.Output)
		}

		var wg sync.WaitGroup

		rl := GetRateLimit()

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

			go func(challenge ctf.ChallengesData) {
				name := lib.CleanSlug(challenge.Name, false)

				category := strings.Split(challenge.Category, " ")[0]
				category = lib.CleanSlug(category, true)

				// make sure name and category are more than 1 character and less than 50
				if len(category) < 1 || len(name) < 1 {
					log.Warnf("Skipping (%q/%q) : invalid name or category", challenge.Name, challenge.Category)
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

				chall, err := client.Challenge(challenge.ID)
				CheckErr(err)

				// download challenge files
				err = client.DownloadFiles(chall.ID, challengePath)
				CheckErr(err)

				// get description
				err = client.GetDescription(chall, challengePath)
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

		// values to config file if --save-config is set
		if opts.SaveConfig {
			viper.Set("url", opts.URL)
			viper.Set("username", opts.Username)
			viper.Set("password", "")
			viper.Set("output", outputFolder)
			viper.Set("overwrite", true)
			viper.SafeWriteConfigAs(path.Join(outputFolder, ".ctftool.yaml"))

			log.WithField("file", path.Join(outputFolder, ".ctftool.yaml")).Info("Saved config file")
			log.Info("You can now run `ctftool` from the same directory without specifying the --url, --username and --password in that directory")
		}
	},
}

func init() {
	ctfdCmd.AddCommand(ctfdDownloadCmd)

	ctfdDownloadCmd.Flags().StringVarP(&opts.URL, "url", "", "", "CTFd URL")
	ctfdDownloadCmd.Flags().StringVarP(&opts.Username, "username", "u", "", "CTFd Username")
	ctfdDownloadCmd.Flags().StringVarP(&opts.Password, "password", "p", "", "CTFd Password")

	// viper
	viper.BindPFlag("url", ctfdDownloadCmd.Flags().Lookup("url"))
	viper.BindPFlag("username", ctfdDownloadCmd.Flags().Lookup("username"))
	viper.BindPFlag("password", ctfdDownloadCmd.Flags().Lookup("password"))
}
