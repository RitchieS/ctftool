package cmd

import (
	"net/url"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/gosimple/slug"
	"github.com/ritchies/ctftool/pkg/ctfd"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/ratelimit"
)

var (
	CTFDUrl          string
	CTFDUser         string
	CTFDPass         string
	CTFDOutputFolder string
	OutputOverwrite  bool
	RateLimit        int
	SaveConfig       bool
	MaxFileSize      int64
)

// ctfdCmd represents the ctfd command
var ctfdCmd = &cobra.Command{
	Use:     "ctfd",
	Aliases: []string{"d", "download"},
	Short:   "Query CTFd instance",
	Long:    `Retrieve challenges and files from a CTFd instance.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := ctfd.NewClient(nil)

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

		// CTFDUser and password are required
		if CTFDUser == "" || CTFDPass == "" {
			cmd.Help()
			log.Fatal("CTFD User and Password are required")
		}

		credentials := ctfd.Credentials{
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
		rl := ratelimit.New(RateLimit)

		for _, challenge := range challenges {
			wg.Add(1)
			rl.Take()
			go func(challenge ctfd.ChallengeData) {
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
					if OutputOverwrite {
						log.Warnf("Overwriting %q : already exists", name)
					} else {
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
			viper.Set("password", CTFDPass)
			viper.Set("output", outputFolder)
			viper.Set("overwrite", true)
			viper.WriteConfigAs(path.Join(outputFolder, ".ctftool.yaml"))

			log.WithField("file", path.Join(outputFolder, ".ctftool.yaml")).Info("Saved config file")
			log.Info("You can now run `ctftool` from the same directory without specifying the --url, --username and --password in that directory")
		}
	},
}

func cleanStr(s string, setLowercase bool) string {
	slug.Lowercase = setLowercase
	s = slug.Make(s)

	if len(s) > 50 {
		tempCategory := strings.Split(s, "-")
		for i := range tempCategory {
			combined := strings.Join(tempCategory[:i+1], "-")
			if len(combined) > 50 {
				s = strings.Join(tempCategory[:i], "-")
			}
		}
		if len(s) > 50 {
			s = s[:50]
		}
	}

	return s
}

func init() {
	rootCmd.AddCommand(ctfdCmd)

	ctfdCmd.Flags().StringVarP(&CTFDUrl, "url", "", "", "CTFd URL")
	ctfdCmd.Flags().StringVarP(&CTFDUser, "username", "u", "", "CTFd Username")
	ctfdCmd.Flags().StringVarP(&CTFDPass, "password", "p", "", "CTFd Password")
	ctfdCmd.Flags().StringVarP(&CTFDOutputFolder, "output", "o", "", "CTFd Output Folder (defaults to current directory)")

	ctfdCmd.Flags().BoolVarP(&OutputOverwrite, "overwrite", "", false, "Overwrite existing files")
	ctfdCmd.Flags().BoolVarP(&SaveConfig, "save-config", "", false, "Save config to (default is $OUTDIR/.ctftool.yaml)")

	// TODO: proper threads

	ctfdCmd.Flags().Int64VarP(&MaxFileSize, "max-file-size", "", 25, "Max file size in mb")

	// viper
	viper.BindPFlag("url", ctfdCmd.Flags().Lookup("url"))
	viper.BindPFlag("username", ctfdCmd.Flags().Lookup("username"))
	viper.BindPFlag("password", ctfdCmd.Flags().Lookup("password"))
	viper.BindPFlag("output", ctfdCmd.Flags().Lookup("output"))
	viper.BindPFlag("overwrite", ctfdCmd.Flags().Lookup("overwrite"))
}
