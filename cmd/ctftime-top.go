package cmd

import (
	"fmt"

	"github.com/ritchies/ctftool/pkg/ctftime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// ctftimeTopCmd represents the top command
var ctftimeTopCmd = &cobra.Command{
	Use:   "top",
	Short: "Displays top 10 teams",
	Long:  `Display the top 10 teams from CTFTime`,
	Run: func(cmd *cobra.Command, args []string) {
		teams, err := ctftime.GetTopTeams()
		if err != nil {
			logrus.Fatalf("Error getting teams: %s", err)
		}

		for i, team := range teams {
			teamID := team.TeamID
			teamName := team.TeamName
			teamPoints := team.Points

			fmt.Printf("%d. %.2f \t%s (%d)\n", i+1, teamPoints, teamName, teamID)
		}
	},
}

func init() {
	ctftimeCmd.AddCommand(ctftimeTopCmd)
}
