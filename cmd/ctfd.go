package cmd

import (
	"github.com/ritchies/ctftool/pkg/ctf"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	opts = ctf.NewOpts()
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

	// viper
	viper.BindPFlag("url", ctfdCmd.Flags().Lookup("url"))
	viper.BindPFlag("username", ctfdCmd.Flags().Lookup("username"))
	viper.BindPFlag("password", ctfdCmd.Flags().Lookup("password"))
	viper.BindPFlag("output", ctfdCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("overwrite", ctfdCmd.PersistentFlags().Lookup("overwrite"))
}
