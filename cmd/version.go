package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `All software has versions. This is mine.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Build:", Build)
		fmt.Println("Version:", Version)
	},
}
