package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// topCmd represents the top command
var topCmd = &cobra.Command{
	Use:   "top",
	Short: "Displays top 10 teams",
	Long:  `Display the top 10 teams from CTFTime`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Info("top called")
	},
}

func init() {
	ctftimeCmd.AddCommand(topCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// topCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// topCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
