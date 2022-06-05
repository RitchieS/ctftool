package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var (
	// The Commit and Version are set to "dev" by default, or set by the Makefile
	Commit    = "dev" // Build is the current build of the program
	Version   = "dev" // Version is the current version of the program
	BuildTime = "dev"
	BuiltBy   = "dev"
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
			fmt.Println("You are running a development build of ctftool")
		} else {
			fmt.Printf("ctftool %s (%s) built by %q on %q\n", Version, Commit, BuiltBy, BuildTime)
		}

		// Display license information
		fmt.Printf("Copyright © %d RitchieS\n", time.Now().Year())
		fmt.Println("All rights reserved.")
	},
}
