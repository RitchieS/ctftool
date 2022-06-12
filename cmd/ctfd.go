package cmd

import (
	"strings"

	"github.com/gosimple/slug"
	"github.com/spf13/cobra"
)

var (
	CTFDUrl          string // the URL of the CTFd instance
	CTFDUser         string // the username used to login to the CTFd instance
	CTFDPass         string // password used to login to the CTFd instance
	CTFDOutputFolder string // the output folder
	OutputOverwrite  bool   // overwrite existing files
	RateLimit        int    // rate limit (per second)
	SaveConfig       bool   // save config bool
	MaxFileSize      int64  // max file size in mb
)

// ctfdCmd represents the ctfd command
var ctfdCmd = &cobra.Command{
	Use:   "ctfd",
	Short: "Query CTFd instance",
	Long:  `Retrieve challenges and files from a CTFd instance.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctfdDownloadCmd.Run(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(ctfdCmd)
}

// cleanStr removes non-alphanumeric characters from a string and will
// lowercase the string if setLower is true.
func cleanStr(s string, setLower bool) string {
	slug.Lowercase = setLower
	s = slug.Make(s)

	if len(s) > 50 {
		tempCategory := strings.Split(s, "-")
		for i := range tempCategory {
			combined := strings.Join(tempCategory[:i+1], "-")
			if len(combined) > 50 {
				s = strings.Join(tempCategory[:i], "-")
			}
		}
		if len(s) > 50 {
			s = s[:50]
		}
	}

	return s
}
