package cmd

import (
	"github.com/spf13/cobra"
)

// ctftimeCmd represents the ctftime command
var ctftimeCmd = &cobra.Command{
	Use:     "ctftime",
	Aliases: []string{"time"},
	Short:   "Query CTFTime",
	Long:    `Retrieve information about upcoming CTF events and teams from CTFTime.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctftimeEventsCmd.Run(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(ctftimeCmd)

	ctftimeCmd.Flags().IntVarP(&limit, "limit", "l", 10, "Limit the number of events to display")
}
