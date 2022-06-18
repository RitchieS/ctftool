package cmd

import (
	"fmt"
	"net/url"

	"github.com/ritchies/ctftool/internal/lib"
	"github.com/ritchies/ctftool/pkg/ctf"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// ctftimeTopCmd represents the top command
var ctftimeTopCmd = &cobra.Command{
	Use:   "top",
	Short: "Displays top 10 teams",
	Long:  `Display the top 10 teams from CTFTime`,
	Run: func(cmd *cobra.Command, args []string) {
		client := ctf.NewClient(nil)
		client.BaseURL, _ = url.Parse(ctftimeURL)

		teams, err := client.GetTopTeams()
		CheckErr(err)

		for i, team := range teams {
			teamID := team.TeamID
			teamName := team.TeamName
			teamPoints := team.Points

			log.WithFields(logrus.Fields{
				"id":     teamID,
				"name":   teamName,
				"points": lib.FtoaWithDigits(teamPoints, 2),
			}).Info(fmt.Sprintf("%d. %s", i+1, teamName))
		}
	},
}

func init() {
	ctftimeCmd.AddCommand(ctftimeTopCmd)
}
