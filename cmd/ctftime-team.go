package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/ritchies/ctftool/pkg/ctftime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var TeamID int

// ctftimeTeamCmd represents the team command
var ctftimeTeamCmd = &cobra.Command{
	Use:   "team",
	Short: "Get information about a team",
	Long:  `Get information about a team on CTFTime.`,
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		// if args is not an integer, exit
		if len(args) > 0 {
			if _, err := fmt.Sscanf(args[0], "%d", &TeamID); err != nil {
				logrus.Errorf("%v", err)
				return
			}
		}

		team, err := ctftime.GetCTFTeam(TeamID)
		if err != nil {
			logrus.Fatalf("Error getting team: %s", err)
		}

		// pretty print the json result
		json, err := json.MarshalIndent(team, "", "  ")
		if err != nil {
			logrus.Fatalf("Error marshalling team: %s", err)
		}
		fmt.Println(string(json))

	},
}

func init() {
	ctftimeCmd.AddCommand(ctftimeTeamCmd)

	ctftimeTeamCmd.Flags().IntVar(&TeamID, "id", 0, "The ID of the team")
}
