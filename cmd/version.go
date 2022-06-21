package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version = "dev"     // Version is the current version of the program
	Commit  = "none"    // Commit is the commit hash of the current build
	Date    = "unknown" // Date is the time the program was built
	BuiltBy = "unknown" // BuiltBy is how the program was built (unknown, goreleaser, etc)
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `All software has versions. This is mine.`,
	Run: func(cmd *cobra.Command, args []string) {
		if Commit == "dev" {
			fmt.Printf("You are running a development build of ctftool\n")
		} else {
			fmt.Printf("ctftool %s (%s) built by %s on %s\n", Version, Commit, BuiltBy, Date)
		}
	},
}
