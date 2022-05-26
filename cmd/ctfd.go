package cmd

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/gosimple/slug"
	"github.com/ritchies/ctftool/pkg/ctfd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	CTFDUrl          string
	CTFDUser         string
	CTFDPass         string
	CTFDOutputFolder string
	OutputOverwrite  bool
	SaveConfig       bool
)

// ctfdCmd represents the ctfd command
var ctfdCmd = &cobra.Command{
	Use:     "ctfd",
	Aliases: []string{"d", "download"},
	Short:   "Query CTFd instance",
	Long:    `Retrieve challenges and files from a CTFd instance.`,
	Run: func(cmd *cobra.Command, args []string) {

		client := ctfd.NewClient(nil)
		log := client.Log
		log.Info("ctfd called")

		// check if username or password are set using viper
		CTFDUrl = viper.GetString("url")
		CTFDUser = viper.GetString("username")
		CTFDPass = viper.GetString("password")
		CTFDOutputFolder = viper.GetString("output")

		baseURL, err := url.Parse(CTFDUrl)
		if err != nil {
			log.Fatal(err)
		}

		if baseURL.Host == "" {
			log.Fatal("Invalid URL")
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

		for _, challenge := range challenges {
			name := cleanStr(challenge.Name, false)

			category := strings.Split(challenge.Category, " ")[0]
			category = cleanStr(category, true)

			// make sure name and category are more than 1 character and less than 50
			if len(category) < 1 || len(name) < 1 {
				log.Warnf("Skipping (%q/%q) : invalid name or category", challenge.Category, challenge.Name)
				continue
			}

			challengePath := path.Join(outputFolder, category, name)

			if _, statErr := os.Stat(challengePath); statErr == nil {
				if OutputOverwrite {
					log.Warnf("Overwriting %q : already exists", challenge.Name)
				} else {
					log.Warnf("Skipping %q : overwrite is false", challenge.Name)
					continue
				}
			}

			if err := os.MkdirAll(challengePath, os.ModePerm); err != nil {
				log.Fatal(fmt.Errorf("error creating directory: %v", err))
			}

			chall, err := client.Challenge(challenge.ID)
			if err != nil {
				log.Fatal(fmt.Errorf("error getting challenge: %v", err))
			}

			// download challenge files
			if err := client.DownloadFiles(chall.ID, challengePath); err != nil {
				log.Fatal(fmt.Errorf("error downloading files: %v", err))
			}

			// get description
			if err := client.GetDescription(chall, challengePath); err != nil {
				log.Fatal(fmt.Errorf("error getting description: %v", err))
			}
		}

		// values to config file if --save-config is set
		if SaveConfig {
			viper.Set("url", CTFDUrl)
			viper.Set("username", CTFDUser)
			viper.Set("password", CTFDPass)
			viper.Set("output", outputFolder)
			viper.Set("overwrite", true)
			viper.WriteConfigAs(path.Join(outputFolder, ".ctftool.yaml"))

			log.WithField("file", path.Join(outputFolder, ".ctftool.yaml")).Info("Saved config file")
			log.Info("You can now run `ctftool` from the same directory without specifying the --url, --username, --password and --output flags")
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

	ctfdCmd.PersistentFlags().StringVarP(&CTFDUrl, "url", "", "", "CTFd URL")
	ctfdCmd.PersistentFlags().StringVarP(&CTFDUser, "username", "u", "", "CTFd Username")
	ctfdCmd.PersistentFlags().StringVarP(&CTFDPass, "password", "p", "", "CTFd Password")
	ctfdCmd.PersistentFlags().StringVarP(&CTFDOutputFolder, "output", "o", "", "CTFd Output Folder (defaults to current directory)")

	ctfdCmd.PersistentFlags().BoolVarP(&OutputOverwrite, "overwrite", "", false, "Overwrite existing files")
	ctfdCmd.PersistentFlags().BoolVarP(&SaveConfig, "save-config", "", false, "Save config to (default is $OUTDIR/.ctftool.yaml)")

	// viper
	viper.BindPFlag("url", ctfdCmd.PersistentFlags().Lookup("url"))
	viper.BindPFlag("username", ctfdCmd.PersistentFlags().Lookup("username"))
	viper.BindPFlag("password", ctfdCmd.PersistentFlags().Lookup("password"))
	viper.BindPFlag("output", ctfdCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("overwrite", ctfdCmd.PersistentFlags().Lookup("overwrite"))
}
