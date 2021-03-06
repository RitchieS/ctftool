package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/ritchies/ctftool/pkg/ctf"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var CTFDSubmission string
var CTFDSubmissionID int

// ctfdSubmitCmd represents the download command
var ctfdSubmitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit a flag",
	Long:  `Submit a flag for a challenge.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := ctf.NewClient(nil)

		// check if flags are set using viper
		opts.URL = viper.GetString("url")
		opts.Username = viper.GetString("username")
		opts.Password = viper.GetString("password")

		baseURL, err := url.Parse(opts.URL)
		CheckErr(err)

		if baseURL.Host == "" {
			ShowHelp(cmd, fmt.Sprintf("Invalid or empty URL provided: %s", baseURL.String()))
		}

		client.BaseURL = baseURL

		if CTFDSubmissionID == 0 {
			ShowHelp(cmd, "CTFD Submission ID is required")
		}

		if CTFDSubmission == "" {
			ShowHelp(cmd, "CTFD Submission is required")
		}

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

		submission := ctf.Submission{
			ID:   CTFDSubmissionID,
			Flag: strings.TrimSpace(CTFDSubmission),
		}

		err = client.SubmitFlag(submission)
		CheckErr(err)

		log.Infof("Successfully submitted flag %q for challenge %d", CTFDSubmission, CTFDSubmissionID)
	},
}

func init() {
	ctfdCmd.AddCommand(ctfdSubmitCmd)

	ctfdSubmitCmd.Flags().StringVarP(&opts.URL, "url", "", "", "CTFd URL")
	ctfdSubmitCmd.Flags().StringVarP(&opts.Username, "username", "u", "", "CTFd Username")
	ctfdSubmitCmd.Flags().StringVarP(&opts.Password, "password", "p", "", "CTFd Password")

	ctfdSubmitCmd.Flags().IntVarP(&CTFDSubmissionID, "id", "i", 0, "CTFd Submission ID")
	ctfdSubmitCmd.Flags().StringVarP(&CTFDSubmission, "submission", "s", "", "Submission")

	// viper
	err := viper.BindPFlag("url", ctfdSubmitCmd.Flags().Lookup("url"))
	CheckErr(err)

	err = viper.BindPFlag("username", ctfdSubmitCmd.Flags().Lookup("username"))
	CheckErr(err)

	err = viper.BindPFlag("password", ctfdSubmitCmd.Flags().Lookup("password"))
	CheckErr(err)
}
