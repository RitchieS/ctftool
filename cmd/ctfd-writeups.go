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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ctfdWriteupCmd represents the writeups command
var ctfdWriteupCmd = &cobra.Command{
	Use:     "writeups",
	Aliases: []string{"w", "write"},
	Short:   "Only create and update writeups",
	Long:    `Create and update writeups for each challenge. Skips downloading files.`,
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

		credentials := ctf.Credentials{
			Username: opts.Username,
			Password: opts.Password,
		}

		client.Creds = &credentials

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

		// Warn the user that they are about to overwrite files
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

		rl := GetRateLimit()
		var wg sync.WaitGroup

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

				err := os.MkdirAll(challengePath, os.ModePerm)
				CheckErr(err)

				chall, err := client.Challenge(challenge.ID)
				CheckErr(err)

				// get description
				err = client.GetDescription(chall, challengePath)
				CheckErr(err)

				log.WithField(
					"category", challenge.Category,
				).Infof("Created writeup for %q", name)

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
			err := viper.SafeWriteConfigAs(path.Join(outputFolder, ".ctftool.yaml"))
			CheckErr(err)

			log.WithField("file", path.Join(outputFolder, ".ctftool.yaml")).Info("Saved config file")
			log.Info("You can now run `ctftool` from the same directory without specifying the --url, --username and --password in that directory")
		}
	},
}

func init() {
	ctfdCmd.AddCommand(ctfdWriteupCmd)

	ctfdWriteupCmd.Flags().StringVarP(&opts.URL, "url", "", "", "CTFd URL")
	ctfdWriteupCmd.Flags().StringVarP(&opts.Username, "username", "u", "", "CTFd Username")
	ctfdWriteupCmd.Flags().StringVarP(&opts.Password, "password", "p", "", "CTFd Password")

	// viper
	err := viper.BindPFlag("url", ctfdWriteupCmd.Flags().Lookup("url"))
	CheckErr(err)

	err = viper.BindPFlag("username", ctfdWriteupCmd.Flags().Lookup("username"))
	CheckErr(err)

	err = viper.BindPFlag("password", ctfdWriteupCmd.Flags().Lookup("password"))
	CheckErr(err)
}
