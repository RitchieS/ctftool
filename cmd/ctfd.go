package cmd

import (
	"github.com/ritchies/ctftool/pkg/ctfd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	opts = ctfd.NewOpts()
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

	ctfdCmd.Flags().StringVarP(&opts.URL, "url", "", "", "CTFd URL")
	ctfdCmd.Flags().StringVarP(&opts.Username, "username", "u", "", "CTFd Username")
	ctfdCmd.Flags().StringVarP(&opts.Password, "password", "p", "", "CTFd Password")
	ctfdCmd.PersistentFlags().StringVarP(&opts.Output, "output", "o", "", "CTFd Output Folder (defaults to current directory)")

	ctfdCmd.PersistentFlags().BoolVarP(&opts.Overwrite, "overwrite", "", false, "Overwrite existing files")
	ctfdCmd.PersistentFlags().BoolVarP(&opts.SaveConfig, "save-config", "", false, "Save config to (default is $OUTDIR/.ctftool.yaml)")
	ctfdCmd.PersistentFlags().BoolVarP(&opts.SkipCTFDCheck, "skip-check", "", false, "Skip CTFd instance check")

	// viper
	err := viper.BindPFlag("url", ctfdCmd.Flags().Lookup("url"))
	CheckErr(err)

	err = viper.BindPFlag("username", ctfdCmd.Flags().Lookup("username"))
	CheckErr(err)

	err = viper.BindPFlag("password", ctfdCmd.Flags().Lookup("password"))
	CheckErr(err)

	err = viper.BindPFlag("output", ctfdCmd.PersistentFlags().Lookup("output"))
	CheckErr(err)

	err = viper.BindPFlag("overwrite", ctfdCmd.PersistentFlags().Lookup("overwrite"))
	CheckErr(err)

	err = viper.BindPFlag("skip-check", ctfdCmd.PersistentFlags().Lookup("skip-check"))
	CheckErr(err)
}
