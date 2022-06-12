package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/ritchies/ctftool/pkg/ctfd"
	"github.com/spf13/cobra"
)

var TeamID int // TeamID is the ID of the team

// ctftimeTeamCmd represents the team command
var ctftimeTeamCmd = &cobra.Command{
	Use:   "team",
	Short: "Get information about a team",
	Long:  `Get information about a team on CTFTime.`,
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		client := ctfd.NewClient(nil)
		client.BaseURL, _ = url.Parse("https://ctftime.org/")

		// if args is not an integer, exit
		if len(args) > 0 {
			if _, err := fmt.Sscanf(args[0], "%d", &TeamID); err != nil {
				log.Errorf("%v", err)
				return
			}
		}

		team, err := client.GetCTFTeam(TeamID)
		if err != nil {
			log.Fatalf("Error getting team: %s", err)
		}

		// pretty print the json result
		json, err := json.MarshalIndent(team, "", "  ")
		if err != nil {
			log.Fatalf("Error marshalling team: %s", err)
		}
		fmt.Println(string(json))

	},
}

func init() {
	ctftimeCmd.AddCommand(ctftimeTeamCmd)

	ctftimeTeamCmd.Flags().IntVar(&TeamID, "id", 0, "The ID of the team")
}
