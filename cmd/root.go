package cmd

import (
	"os"
	"path"

	"github.com/ritchies/ctftool/internal/lib"
	"github.com/ritchies/ctftool/internal/storage"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/ratelimit"
)

var (
	options = lib.NewOptions() // global options
	log     = logrus.New()     // global logger
	dB      = storage.NewDb()  // global database
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ctftool",
	Short: "ctftool is a tool for interacting with CTF instances",
	Long: `ctftool is a tool for interacting with CTF instances.

It can interact with the CTFTime.org API to retrieve the latest upcoming CTFs,
and can interact with CTFd API to retrieve the challenges and files.`,
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.Flags().Lookup("version").Value.String() == "true" {
			versionCmd.Run(cmd, args)
		} else {
			if viper.ConfigFileUsed() != "" {
				ctfdCmd.Run(cmd, args)
			} else {
				ctftimeCmd.Run(cmd, args)
			}
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

	rootCmd.PersistentFlags().BoolVar(&options.Interactive, "interactive", false, "Interactive mode")

	rootCmd.PersistentFlags().IntVarP(&options.RateLimit, "rate-limit", "", 3, "Rate limit (per second)")
	rootCmd.PersistentFlags().Int64VarP(&options.MaxFileSize, "max-file-size", "", 25, "Max file size in mb")

	rootCmd.PersistentFlags().StringVar(&options.ConfigFile, "config", "", "Config file (default is .ctftool.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&options.Debug, "debug", "d", false, "Verbose logging")
	rootCmd.PersistentFlags().BoolVarP(&options.Debug, "verbose", "v", false, "Verbose logging")
	rootCmd.PersistentFlags().StringVar(&options.DebugFormat, "log-format", "text", "Logger output format (text|json)")

	rootCmd.PersistentFlags().StringVar(&dB.Path, "db-path", "ctftool.sqlite", "Path to the database file")

	rootCmd.PersistentFlags().BoolP("version", "V", false, "Print version information")

	rootCmd.PersistentFlags().MarkHidden("debug")
	rootCmd.PersistentFlags().MarkHidden("log-format")
	rootCmd.PersistentFlags().MarkHidden("version")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if options.ConfigFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(options.ConfigFile)
	} else {
		cwd, err := os.Getwd()
		cobra.CheckErr(err)

		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search .config in home directory with name "ctftool"
		homeConfig := path.Join(home, ".config", "ctftool")

		// Search config in home/cwd directory with name ".ctftool" (without extension).
		viper.AddConfigPath(cwd)
		viper.AddConfigPath(homeConfig)
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
	if options.Debug {
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

func GetRateLimit() ratelimit.Limiter {
	var rl ratelimit.Limiter

	if options.RateLimit > 0 && options.RateLimit < 100 {
		rl = ratelimit.New(options.RateLimit)
	} else {
		rl = ratelimit.New(100)
	}

	return rl
}
