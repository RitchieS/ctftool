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
		CTFDUrl = viper.GetString("url")
		CTFDUser = viper.GetString("username")
		CTFDPass = viper.GetString("password")

		baseURL, err := url.Parse(CTFDUrl)
		if err != nil || baseURL.Host == "" {
			cmd.Help()
			log.Fatalf("Invalid or empty URL provided: %s", baseURL.String())
		}

		client.BaseURL = baseURL

		if CTFDSubmissionID == 0 {
			cmd.Help()
			log.Fatal("CTFD Submission ID is required")
		}

		if CTFDSubmission == "" {
			cmd.Help()
			log.Fatal("CTFD Submission is required")
		}

		if CTFDUser != "" && CTFDPass == "" {
			fmt.Print("Enter your password: ")
			var password string
			fmt.Scanln(&password)
			CTFDPass = strings.TrimSpace(password)
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

		submission := ctf.Submission{
			ID:   CTFDSubmissionID,
			Flag: strings.TrimSpace(CTFDSubmission),
		}

		if err := client.SubmitFlag(submission); err != nil {
			log.Fatal(err)
		}

		log.Infof("Successfully submitted flag %q for challenge %d", CTFDSubmission, CTFDSubmissionID)
	},
}

func init() {
	ctfdCmd.AddCommand(ctfdSubmitCmd)

	ctfdSubmitCmd.Flags().StringVarP(&CTFDUrl, "url", "", "", "CTFd URL")
	ctfdSubmitCmd.Flags().StringVarP(&CTFDUser, "username", "u", "", "CTFd Username")
	ctfdSubmitCmd.Flags().StringVarP(&CTFDPass, "password", "p", "", "CTFd Password")

	ctfdSubmitCmd.Flags().IntVarP(&CTFDSubmissionID, "id", "i", 0, "CTFd Submission ID")
	ctfdSubmitCmd.Flags().StringVarP(&CTFDSubmission, "submission", "s", "", "Submission")

	// viper
	viper.BindPFlag("url", ctfdSubmitCmd.Flags().Lookup("url"))
	viper.BindPFlag("username", ctfdSubmitCmd.Flags().Lookup("username"))
	viper.BindPFlag("password", ctfdSubmitCmd.Flags().Lookup("password"))
}
