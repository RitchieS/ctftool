package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/ritchies/ctftool/pkg/ctftime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var TeamID int

// teamCmd represents the team command
var teamCmd = &cobra.Command{
	Use:   "team",
	Short: "Get information about a team",
	Long:  `Get information about a team on CTFTime.`,
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Info("team called")

		// print args
		logrus.Debugf("args: %v", args)

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
	ctftimeCmd.AddCommand(teamCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// teamCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// teamCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	teamCmd.Flags().IntVar(&TeamID, "id", 0, "The ID of the team")

}
