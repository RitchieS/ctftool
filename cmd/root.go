package cmd

import (
	"os"

	"github.com/ritchies/ctftool/internal/lib"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	Build   = "dev" // Build is the current build of the program
	Version = "dev" // Version is the current version of the program
)

var (
	options = lib.NewOptions()
	log     = logrus.New()
	// db      = storage.NewDb()
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ctftool",
	Short: "ctftool is a tool for interacting with CTF instances",
	Long: `ctftool is a tool for interacting with CTF instances.

It can interact with the CTFTime.org API to retrieve the latest upcoming CTFs,
and can interact with CTFd API to retrieve the challenges and files.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {

		// if viper config file is in current directory;
		if viper.ConfigFileUsed() != "" {
			ctfdCmd.Run(cmd, args)
		} else {
			ctftimeCmd.Run(cmd, args)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.PersistentFlags().StringVar(&options.ConfigFile, "config", "", "config file (default is $HOME/.ctftool.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&options.Verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().BoolVarP(&options.Debug, "debug", "d", false, "Debug output")
	rootCmd.PersistentFlags().StringVar(&options.DebugFormat, "debug-format", "text", "Debug output format (text|json)")

	// rootCmd.PersistentFlags().StringVar(&db.Path, "db-path", "ctftool.sqlite", "Path to the database file")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if options.ConfigFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(options.ConfigFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		cwd, err := os.Getwd()
		cobra.CheckErr(err)

		// Search config in home/cwd directory with name ".ctftool" (without extension).
		viper.AddConfigPath(cwd)
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".ctftool")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// Logrus options

	log.SetFormatter(&logrus.TextFormatter{
		DisableSorting:         false,
		DisableTimestamp:       true,
		DisableLevelTruncation: true,
		ForceColors:            true,
		ForceQuote:             true,
		PadLevelText:           true,
		QuoteEmptyFields:       true,
	})

	// set log level to debug
	if options.Debug || options.Verbose {
		log.SetLevel(logrus.DebugLevel)
	}

	// set text or json output
	switch options.DebugFormat {
	case "json":
		log.SetFormatter(&logrus.JSONFormatter{})
	}

	// Set log output
	log.Out = os.Stdout

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.WithField("config", viper.ConfigFileUsed()).Debug("Using config file")
	}
}
