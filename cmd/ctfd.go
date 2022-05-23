package cmd

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/ritchies/ctftool/pkg/ctfd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	CTFDUrl          string
	CTFDCookie       string
	CTFDOutputFolder string
	OutputOverwrite  bool
	CTFDUser         string
	CTFDPass         string
	CTFDEmail        string
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
		CTFDUser = viper.GetString("username")
		CTFDPass = viper.GetString("password")
		CTFDUrl = viper.GetString("url")

		baseURL, err := url.Parse(CTFDUrl)
		if err != nil {
			log.Fatal(err)
		}

		client.BaseURL = baseURL

		reader := bufio.NewReader(os.Stdin)
		if CTFDPass == "" && CTFDUser != "" {
			fmt.Print("password: ")
			CTFDPass, _ = reader.ReadString('\n')
			CTFDPass = strings.TrimSpace(CTFDPass)
		}

		// CTFDUser and password are required
		if CTFDUser == "" || CTFDPass == "" {
			cmd.Help()
			log.Fatal("CTFD URL, CTFD User and Password are required")
		}

		if err := client.Authenticate(CTFDUser, CTFDPass); err != nil {
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

		for _, challenge := range challenges {
			outputFolder := path.Join(cwd, CTFDOutputFolder)

			category := strings.Split(challenge.Category, " ")[0]
			category = strings.ToLower(category)

			name := strings.Replace(challenge.Name, " ", "-", -1)
			name = regexp.MustCompile("[^a-zA-Z0-9-.]+").ReplaceAllString(name, "_")
			name = regexp.MustCompile("-_|_-|-_-|_-_").ReplaceAllString(name, "-")
			name = regexp.MustCompile("_+").ReplaceAllString(name, "_")
			name = strings.Trim(name, "-_")

			if len(category) > 50 {
				category = category[:50]
			}
			if len(name) > 50 {
				name = name[:50]
			}

			// make sure name and category are more than 1 character and less than 50
			if len(category) < 1 || len(name) < 1 {
				log.Warnf("skipping challenge %q because it has an invalid name or category", challenge.Name)
				continue
			}

			filePath := path.Join(outputFolder, category, name)

			if _, statErr := os.Stat(filePath); statErr == nil {
				log.Warnf("Challenge %q already exists", challenge.Name)
				// continue if overwrite is false
				if !OutputOverwrite {
					log.Debugf("skipping challenge %q because overwrite is false", challenge.Name)
					continue
				} else {
					log.Debugf("overwriting challenge %q", challenge.Name)
				}
			}

			if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
				log.Fatal(fmt.Errorf("error creating directory: %v", err))
			}

			chall, err := client.Challenge(challenge.ID)
			if err != nil {
				log.Fatal(fmt.Errorf("error getting challenge: %v", err))
			}

			// download challenge files
			if err := client.DownloadFiles(chall.ID, filePath); err != nil {
				log.Fatal(fmt.Errorf("error downloading files: %v", err))
			}

			// get description
			if err := client.GetDescription(chall, filePath); err != nil {
				log.Fatal(fmt.Errorf("error getting description: %v", err))
			}
		}

		// values to config file
		viper.WriteConfigAs(path.Join(os.Getenv("HOME"), ".ctftool.yaml"))
	},
}

func init() {
	rootCmd.AddCommand(ctfdCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// ctfdCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// ctfdCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	// ctfdCmd.Flags().StringVarP(&CTFDUrl, "url", "u", "", "CTFd URL")
	// ctfdCmd.Flags().StringVarP(&CTFDCookie, "cookie", "c", "", "CTFd Cookie")
	// ctfdCmd.Flags().StringVarP(&CTFDOutputFolder, "output", "o", "", "CTFd Output Folder")

	// ctfdCmd.Flags().StringVarP(&CTFDUser, "user", "", "", "CTFd User")
	// ctfdCmd.Flags().StringVarP(&CTFDPass, "pass", "", "", "CTFd Password")

	// ctfdCmd.MarkFlagRequired("url")
	// ctfdCmd.MarkFlagRequired("output")

	ctfdCmd.PersistentFlags().StringVarP(&CTFDUrl, "url", "", "", "CTFd URL")
	ctfdCmd.PersistentFlags().StringVarP(&CTFDCookie, "cookie", "c", "", "CTFd Cookie")
	ctfdCmd.PersistentFlags().StringVarP(&CTFDOutputFolder, "output", "o", "", "CTFd Output Folder")

	// overwrite
	ctfdCmd.PersistentFlags().BoolVarP(&OutputOverwrite, "overwrite", "", false, "Overwrite existing files")

	ctfdCmd.PersistentFlags().StringVarP(&CTFDUser, "username", "u", "", "CTFd Username")
	ctfdCmd.PersistentFlags().StringVarP(&CTFDPass, "password", "p", "", "CTFd Password")

	// viper
	viper.BindPFlag("username", ctfdCmd.PersistentFlags().Lookup("username"))
	viper.BindPFlag("password", ctfdCmd.PersistentFlags().Lookup("password"))
	viper.BindPFlag("url", ctfdCmd.PersistentFlags().Lookup("url"))

	ctfdCmd.MarkPersistentFlagRequired("url")
	ctfdCmd.MarkPersistentFlagRequired("output")

}
