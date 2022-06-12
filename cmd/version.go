package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// The Commit and Version are set to "dev" by default, or set by the Makefile
	Commit    = "dev" // Commit is the commit hash of the current build
	Version   = "dev" // Version is the current version of the program
	BuildTime = "dev" // BuildTime is the time the program was built
	BuiltBy   = "dev" // BuiltBy is how the program was built (dev, goreleaser, etc)
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
			fmt.Printf("ctftool %s (%s) built by %q on %q\n", Version, Commit, BuiltBy, BuildTime)
		}
	},
}
