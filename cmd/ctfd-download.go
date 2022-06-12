package cmd

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/ritchies/ctftool/pkg/ctf"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/ratelimit"
	"golang.org/x/term"
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
		client.MaxFileSize = MaxFileSize

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

		for _, challenge := range challenges {
			wg.Add(1)

			if RateLimit > 0 {
				rl.Take()
			}

			go func(challenge ctf.ChallengeData) {
				name := cleanStr(challenge.Name, false)

				category := strings.Split(challenge.Category, " ")[0]
				category = cleanStr(category, true)

				// make sure name and category are more than 1 character and less than 50
				if len(category) < 1 || len(name) < 1 {
					log.Warnf("Skipping (%q/%q) : invalid name or category", challenge.Name, challenge.Category)
					wg.Done()
					return
				}

				challengePath := path.Join(outputFolder, category, name)

				if _, statErr := os.Stat(challengePath); statErr == nil {
					if !OutputOverwrite {
						log.Warnf("Skipping %q : overwrite is false", name)
						wg.Done()
						return
					}
				}

				if err := os.MkdirAll(challengePath, os.ModePerm); err != nil {
					log.Fatalf("error creating directory %q: %v", challengePath, err)
				}

				chall, err := client.Challenge(challenge.ID)
				if err != nil {
					log.Fatalf("error getting challenge %q: %v", name, err)
				}

				// download challenge files
				if err := client.DownloadFiles(chall.ID, challengePath); err != nil {
					log.Errorf("error downloading files for %q: %v", name, err)
				}

				// get description
				if err := client.GetDescription(chall, challengePath); err != nil {
					log.Fatalf("error getting description for %q: %v", name, err)
				}

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
		if SaveConfig {
			viper.Set("url", CTFDUrl)
			viper.Set("username", CTFDUser)
			viper.Set("output", outputFolder)
			viper.Set("overwrite", true)
			viper.WriteConfigAs(path.Join(outputFolder, ".ctftool.yaml"))

			log.WithField("file", path.Join(outputFolder, ".ctftool.yaml")).Info("Saved config file")
			log.Info("You can now run `ctftool` from the same directory without specifying the --url, --username and --password in that directory")
		}
	},
}

func init() {
	ctfdCmd.AddCommand(ctfdDownloadCmd)

	ctfdDownloadCmd.Flags().StringVarP(&CTFDUrl, "url", "", "", "CTFd URL")
	ctfdDownloadCmd.Flags().StringVarP(&CTFDUser, "username", "u", "", "CTFd Username")
	ctfdDownloadCmd.Flags().StringVarP(&CTFDPass, "password", "p", "", "CTFd Password")
	ctfdDownloadCmd.Flags().StringVarP(&CTFDOutputFolder, "output", "o", "", "CTFd Output Folder (defaults to current directory)")

	ctfdDownloadCmd.Flags().BoolVarP(&OutputOverwrite, "overwrite", "", false, "Overwrite existing files")
	ctfdDownloadCmd.Flags().BoolVarP(&SaveConfig, "save-config", "", false, "Save config to (default is $OUTDIR/.ctftool.yaml)")

	// TODO: proper threads
	ctfdDownloadCmd.Flags().IntVarP(&RateLimit, "rate-limit", "", 1, "Rate limit (per second)")

	ctfdDownloadCmd.Flags().Int64VarP(&MaxFileSize, "max-file-size", "", 25, "Max file size in mb")

	// viper
	viper.BindPFlag("url", ctfdDownloadCmd.Flags().Lookup("url"))
	viper.BindPFlag("username", ctfdDownloadCmd.Flags().Lookup("username"))
	viper.BindPFlag("password", ctfdDownloadCmd.Flags().Lookup("password"))
	viper.BindPFlag("output", ctfdDownloadCmd.Flags().Lookup("output"))
	viper.BindPFlag("overwrite", ctfdDownloadCmd.Flags().Lookup("overwrite"))

}
