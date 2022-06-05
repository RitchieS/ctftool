package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// The Build and Version are set to "dev" by default, or set by the Makefile
	Build   = "dev" // Build is the current build of the program
	Version = "dev" // Version is the current version of the program
	builtBy = "dev"
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
		fmt.Println("Built by:", builtBy)
	},
}
