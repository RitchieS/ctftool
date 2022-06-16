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

		if err := client.Authenticate(); err != nil {
			log.Fatal(err)
		}
		log.Infof("Authenticated as %q", opts.Username)

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

	ctfdSubmitCmd.Flags().StringVarP(&opts.URL, "url", "", "", "CTFd URL")
	ctfdSubmitCmd.Flags().StringVarP(&opts.Username, "username", "u", "", "CTFd Username")
	ctfdSubmitCmd.Flags().StringVarP(&opts.Password, "password", "p", "", "CTFd Password")

	ctfdSubmitCmd.Flags().IntVarP(&CTFDSubmissionID, "id", "i", 0, "CTFd Submission ID")
	ctfdSubmitCmd.Flags().StringVarP(&CTFDSubmission, "submission", "s", "", "Submission")

	// viper
	viper.BindPFlag("url", ctfdSubmitCmd.Flags().Lookup("url"))
	viper.BindPFlag("username", ctfdSubmitCmd.Flags().Lookup("username"))
	viper.BindPFlag("password", ctfdSubmitCmd.Flags().Lookup("password"))
}
