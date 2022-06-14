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
	"go.uber.org/ratelimit"
	"golang.org/x/term"
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
		CTFDUrl = viper.GetString("url")
		CTFDUser = viper.GetString("username")
		CTFDPass = viper.GetString("password")
		CTFDOutputFolder = viper.GetString("output")
		OutputOverwrite = viper.GetBool("overwrite")

		baseURL, err := url.Parse(CTFDUrl)
		if err != nil || baseURL.Host == "" {
			cmd.Help()
			log.Fatalf("Invalid or empty URL provided: %s", baseURL.String())
		}

		client.BaseURL = baseURL

		if CTFDUser != "" && CTFDPass == "" {
			fmt.Print("Password: ")
			bytepwd, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				log.Fatalf("Error reading password: %s", err)
			}
			fmt.Printf("\n")
			CTFDPass = strings.TrimSpace(string(bytepwd))
		}

		// CTFDUser and password are required
		if CTFDUser == "" || CTFDPass == "" {
			cmd.Help()
			log.Fatal("CTFD User and Password are required")
		}

		credentials := ctf.Credentials{
			Username: CTFDUser,
			Password: CTFDPass,
		}

		client.Creds = &credentials

		if err := client.Authenticate(); err != nil {
			log.Fatal(err)
		}
		log.Infof("Authenticated as %q", CTFDUser)

		// List challenges
		challenges, err := client.ListChallenges()
		if err != nil {
			log.Fatal(err)
		}

		cwd, err := os.Getwd()
		if err != nil {
			log.Fatalf("error getting current working directory: %v", err)
		}

		outputFolder := cwd

		// if using config file
		if viper.ConfigFileUsed() == "" && CTFDOutputFolder != "" {
			outputFolder = path.Join(cwd, CTFDOutputFolder)
		}

		var wg sync.WaitGroup
		var rl ratelimit.Limiter

		if RateLimit > 0 && RateLimit < 100 {
			rl = ratelimit.New(RateLimit)
		} else {
			rl = ratelimit.New(100)
		}

		// Warn the user that they are about to overwrite files
		log.Warn("This action will overwrite existing files")
		log.Info("Writeups will be updated if they exist")
		log.Info("Press enter or ctrl+c to cancel")

		// Ask the user if they want to continue
		fmt.Print("Do you want to continue? [y/N]: ")
		var answer string
		fmt.Scanln(&answer)
		if strings.ToLower(answer) != "y" {
			log.Fatal("Aborting by user request")
		}

		for _, challenge := range challenges {
			wg.Add(1)

			if RateLimit > 0 {
				rl.Take()
			}

			go func(challenge ctf.ChallengeData) {
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

				if err := os.MkdirAll(challengePath, os.ModePerm); err != nil {
					log.Fatalf("error creating directory %q: %v", challengePath, err)
				}

				chall, err := client.Challenge(challenge.ID)
				if err != nil {
					log.Fatalf("error getting challenge %q: %v", name, err)
				}

				// get description
				if err := client.GetDescription(chall, challengePath); err != nil {
					log.Fatalf("error getting description for %q: %v", name, err)
				}

				log.WithField(
					"category", challenge.Category,
				).Infof("Created writeup for %q", name)

				wg.Done()
			}(challenge)
		}
		wg.Wait()

		// values to config file if --save-config is set
		if SaveConfig {
			viper.Set("url", CTFDUrl)
			viper.Set("username", CTFDUser)
			viper.Set("password", "")
			viper.Set("output", outputFolder)
			viper.Set("overwrite", true)
			viper.WriteConfigAs(path.Join(outputFolder, ".ctftool.yaml"))

			log.WithField("file", path.Join(outputFolder, ".ctftool.yaml")).Info("Saved config file")
			log.Info("You can now run `ctftool` from the same directory without specifying the --url, --username and --password in that directory")
		}
	},
}

func init() {
	ctfdCmd.AddCommand(ctfdWriteupCmd)

	ctfdWriteupCmd.Flags().StringVarP(&CTFDUrl, "url", "", "", "CTFd URL")
	ctfdWriteupCmd.Flags().StringVarP(&CTFDUser, "username", "u", "", "CTFd Username")
	ctfdWriteupCmd.Flags().StringVarP(&CTFDPass, "password", "p", "", "CTFd Password")

	// viper
	viper.BindPFlag("url", ctfdWriteupCmd.Flags().Lookup("url"))
	viper.BindPFlag("username", ctfdWriteupCmd.Flags().Lookup("username"))
	viper.BindPFlag("password", ctfdWriteupCmd.Flags().Lookup("password"))
}
