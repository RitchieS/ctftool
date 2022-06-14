package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

	ctfdCmd.Flags().StringVarP(&CTFDUrl, "url", "", "", "CTFd URL")
	ctfdCmd.Flags().StringVarP(&CTFDUser, "username", "u", "", "CTFd Username")
	ctfdCmd.Flags().StringVarP(&CTFDPass, "password", "p", "", "CTFd Password")
	ctfdCmd.PersistentFlags().StringVarP(&CTFDOutputFolder, "output", "o", "", "CTFd Output Folder (defaults to current directory)")

	ctfdCmd.PersistentFlags().BoolVarP(&OutputOverwrite, "overwrite", "", false, "Overwrite existing files")
	ctfdCmd.PersistentFlags().BoolVarP(&SaveConfig, "save-config", "", false, "Save config to (default is $OUTDIR/.ctftool.yaml)")

	// TODO: proper threads
	ctfdCmd.PersistentFlags().IntVarP(&RateLimit, "rate-limit", "", 3, "Rate limit (per second)")
	ctfdCmd.PersistentFlags().Int64VarP(&MaxFileSize, "max-file-size", "", 25, "Max file size in mb")

	// viper
	viper.BindPFlag("url", ctfdCmd.Flags().Lookup("url"))
	viper.BindPFlag("username", ctfdCmd.Flags().Lookup("username"))
	viper.BindPFlag("password", ctfdCmd.Flags().Lookup("password"))
	viper.BindPFlag("output", ctfdCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("overwrite", ctfdCmd.PersistentFlags().Lookup("overwrite"))

}
