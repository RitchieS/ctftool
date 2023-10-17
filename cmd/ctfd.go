package cmd

import (
	"path"

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
		cmd.Usage()
	},
}

func init() {
	rootCmd.AddCommand(ctfdCmd)

	ctfdCmd.PersistentFlags().StringVarP(&opts.Output, "output", "o", "", "CTFd Output Folder (defaults to current directory)")
	ctfdCmd.PersistentFlags().BoolVarP(&opts.Overwrite, "overwrite", "", false, "Overwrite existing files")
	ctfdCmd.PersistentFlags().BoolVarP(&opts.SaveConfig, "save-config", "", false, "Save config to (default is $OUTDIR/.ctftool.yaml)")
	ctfdCmd.PersistentFlags().BoolVarP(&opts.SkipCTFDCheck, "skip-check", "", false, "Skip CTFd instance check")

	// viper
	err := viper.BindPFlag("output", ctfdCmd.PersistentFlags().Lookup("output"))
	CheckErr(err)

	err = viper.BindPFlag("overwrite", ctfdCmd.PersistentFlags().Lookup("overwrite"))
	CheckErr(err)

	err = viper.BindPFlag("skip-check", ctfdCmd.PersistentFlags().Lookup("skip-check"))
	CheckErr(err)

	err = viper.BindPFlag("save-config", ctfdCmd.PersistentFlags().Lookup("save-config"))
	CheckErr(err)
}

func saveConfig() {
	viper.Reset()
	viper.Set("url", opts.URL)
	viper.Set("username", opts.Username)
	viper.Set("password", opts.Password)
	viper.Set("token", opts.Token)
	viper.Set("output", opts.Output)
	viper.Set("overwrite", opts.Overwrite)
	viper.Set("skip-check", opts.SkipCTFDCheck)
	viper.Set("unsolved", opts.UnsolvedOnly)
	viper.Set("watch", opts.Watch)
	viper.Set("watch-interval", opts.WatchInterval)
	err := viper.SafeWriteConfigAs(path.Join(opts.Output, ".ctftool.yaml"))
	CheckErr(err)

	log.WithField("file", path.Join(opts.Output, ".ctftool.yaml")).Info("Saved config file")
	log.Info("You can now run `ctftool` from the same directory without specifying the --url, --username and --password in that directory")
}
