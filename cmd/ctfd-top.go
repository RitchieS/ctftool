package cmd

import (
	"fmt"
	"net/url"

	"github.com/ritchies/ctftool/pkg/ctfd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ctfdTopCmd represents the top command
var ctfdTopCmd = &cobra.Command{
	Use:     "top",
	Short:   "Displays top 10 teams",
	Long:    `Display the top 10 teams from CTFd`,
	Example: `  ctftool ctfd top --url https://demo.ctfd.io`,
	Run: func(cmd *cobra.Command, args []string) {
		client := ctfd.NewClient()
		uri := viper.GetString("url")

		baseURL, err := url.Parse(uri)
		CheckErr(err)

		if baseURL.Host == "" {
			ShowHelp(cmd, fmt.Sprintf("Invalid or empty URL provided: %q", baseURL.String()))
		}

		client.BaseURL = baseURL

		teamsData, err := ctfd.ScoreboardTop(10)
		CheckErr(err)

		for i := 1; i <= 10; i++ {
			team, err := teamsData.GetTeam(i)
			CheckErr(err)

			if team.ID == 0 {
				continue
			}

			teamSolves := team.Solves
			teamScore := 0
			for _, solve := range teamSolves {
				teamScore += solve.Value
			}

			log.WithField("score", teamScore).Info(
				fmt.Sprintf("%d. %s", i, team.Name),
			)
		}
	},
}

func init() {
	ctfdCmd.AddCommand(ctfdTopCmd)

	ctfdTopCmd.Flags().StringVarP(&opts.URL, "url", "u", "", "URL of the CTFd instance")

	err := viper.BindPFlag("url", ctfdTopCmd.Flags().Lookup("url"))
	CheckErr(err)
}
