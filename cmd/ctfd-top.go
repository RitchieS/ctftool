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
	Use:   "top",
	Short: "Displays top 10 teams",
	Long:  `Display the top 10 teams from CTFd`,
	Run: func(cmd *cobra.Command, args []string) {
		client := ctfd.NewClient(nil)
		log := client.Log

		uri := viper.GetString("url")

		baseURL, err := url.Parse(uri)
		if err != nil {
			log.Errorf("Invalid or empty URL provided %q: %s", uri, err)
			return
		}

		client.BaseURL = baseURL

		if client.BaseURL.Host == "" {
			log.Errorf("Invalid or empty URL provided %q: %s", uri, err)
			return
		}

		teamsData, err := client.ScoreboardTop(10)
		if err != nil {
			log.Fatalf("Failed to retrieve scoreboard: %s", err)
		}

		for i := 1; i <= 10; i++ {
			team, err := teamsData.GetTeam(i)
			if err != nil {
				log.Fatalf("Failed to retrieve team %d: %s", i, err)
			}

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
}
