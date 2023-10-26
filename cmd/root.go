package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/ritchies/ctftool/internal/lib"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/ratelimit"
)

var (
	options = lib.NewOptions() // global options
	log     = logrus.New()     // global logger
)

func contains(slice []string, str string) bool {
	for _, a := range slice {
		if a == str {
			return true
		}
	}
	return false
}

func filterOutCategorizedFlags(cmd *cobra.Command, categorizedFlags []string) *pflag.FlagSet {
	filteredFlags := pflag.NewFlagSet("filtered", pflag.ContinueOnError)
	addFlags := func(flagSet *pflag.FlagSet) {
		flagSet.VisitAll(func(flag *pflag.Flag) {
			if !contains(categorizedFlags, flag.Name) && flag.Name != "help" {
				filteredFlags.AddFlag(flag)
			}
		})
	}
	// Here, use cmd.Flags() to include both local and inherited flags.
	addFlags(cmd.Flags())
	return filteredFlags
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ctftool",
	Short: "ctftool is a tool for interacting with CTF instances",
	Long: `ctftool is a tool for interacting with CTF instances.

It can interact with the CTFTime.org API to retrieve the latest upcoming CTFs,
and can interact with CTFd API to retrieve the challenges and files.`,
	Run: func(cmd *cobra.Command, args []string) {
		if viper.ConfigFileUsed() != "" {
			ctfdCmd.Run(cmd, args)
		} else {
			err := cmd.Help()
			CheckErr(err)
		}
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		// Add the current command name to the Aliases slice for usability
		if !contains(cmd.Aliases, cmd.Name()) {
			cmd.Aliases = append(cmd.Aliases, cmd.Name())
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	CheckErr(err)
}

func filterExistingFlags(cmd *cobra.Command, flagNames []string) []string {
	var existingFlags []string
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if contains(flagNames, flag.Name) {
			existingFlags = append(existingFlags, flag.Name)
		}
	})
	return existingFlags
}

type FlagCategory struct {
	Name  string
	Flags []string
}

func getFlag(cmd *cobra.Command, flagName string) *pflag.Flag {
	flag := cmd.Flags().Lookup(flagName)
	if flag == nil {
		flag = cmd.PersistentFlags().Lookup(flagName)
		if flag == nil && cmd.Parent() != nil {
			flag = cmd.Parent().PersistentFlags().Lookup(flagName)
		}
	}
	return flag
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().IntVarP(&options.RateLimit, "rate-limit", "", 10, "Limit the number of API requests per second")
	rootCmd.PersistentFlags().StringVar(&options.ConfigFile, "config", "", "Config file (default is .ctftool.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&options.Debug, "verbose", "v", false, "Verbose logging")
	rootCmd.PersistentFlags().StringVar(&options.DebugFormat, "log-format", "text", "Format for logging output (text or json)")

	var ctftimeFlags = FlagCategory{
		Name:  "CTFTime",
		Flags: []string{"team-id", "event-id", "limit"},
	}

	var ctfdFlags = FlagCategory{
		Name:  "CTFd",
		Flags: []string{"url", "submission-id", "submission", "unsolved", "skip-check", "output", "overwrite", "max-file-size"},
	}

	var authFlags = FlagCategory{
		Name:  "Authentication",
		Flags: []string{"username", "password", "token"},
	}

	var notificationFlags = FlagCategory{
		Name:  "Notifications",
		Flags: []string{"notify", "watch", "watch-interval"},
	}

	var allFlagCategories = []FlagCategory{ctftimeFlags, ctfdFlags, authFlags, notificationFlags}

	usageTemplate := `Usage:
  {{.CommandPath}} [flags]
  {{- if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]
  {{- end}}

{{- if .HasAvailableSubCommands}}

Available Commands:
{{- range .Commands}}
  {{rpad .Name .NamePadding }} {{.Short}}
{{- end}}
{{- end}}

{{- if and (gt (len .Aliases) 0) .HasParent}}

Aliases:
  {{join .Aliases ", "}}
{{- end}}

{{- if .HasExample}}

Examples:
{{.Example}}
{{- end}}

{{- range $category := .FlagCategories}}
{{- if and $category.Flags (gt (len $category.Flags) 0)}}
  
{{$category.Name}}:
{{- range $flag := $category.Flags}}
  {{- with (flag $flag)}}
  {{- if .Shorthand}}
  {{printf "-%-1s, --%-19s %s" .Shorthand .Name .Usage}}
  {{- else}}
  {{printf "    --%-19s %s" .Name .Usage}}
  {{- end}}
  {{- end}}
{{- end}}
{{- end}}
{{- end}}

Flags:
{{.CombinedFlagUsages}}
`

	rootCmd.SetUsageTemplate(usageTemplate)
	rootCmd.SetUsageFunc(func(cmd *cobra.Command) error {
		// Initialize the template
		tmpl, err := template.New("usage").Funcs(template.FuncMap{
			"flag": func(flagName string) *pflag.Flag {
				return getFlag(cmd, flagName)
			},
			"join": strings.Join,
			"rpad": func(s string, padding int) string {
				return fmt.Sprintf("%-*s", padding, s)
			},
		}).Parse(usageTemplate)
		if err != nil {
			return err
		}

		// Filter out non-existing flags for each category
		for i := range allFlagCategories {
			allFlagCategories[i].Flags = filterExistingFlags(cmd, allFlagCategories[i].Flags)
		}

		// Collect all categorized flag names into a single slice
		var allCategorizedFlags []string
		for _, category := range allFlagCategories {
			allCategorizedFlags = append(allCategorizedFlags, category.Flags...)
		}

		// Generate flag usages for categorized flags
		var categorizedFlagUsages string
		for _, flagName := range allCategorizedFlags {
			flag := getFlag(cmd, flagName)
			if flag != nil {
				if flag.Shorthand != "" {
					categorizedFlagUsages += fmt.Sprintf("  -%s, --%-19s %s\n", flag.Shorthand, flag.Name, flag.Usage)
				} else {
					categorizedFlagUsages += fmt.Sprintf("      --%-19s %s\n", flag.Name, flag.Usage)
				}
			}
		}

		// Filter out the categorized flags and 'help' flag from all flags
		filteredFlags := filterOutCategorizedFlags(cmd, append(allCategorizedFlags, "help"))

		// Generate flag usages for the filtered flags
		var filteredFlagUsages string
		filteredFlags.VisitAll(func(flag *pflag.Flag) {
			if flag.Shorthand != "" {
				filteredFlagUsages += fmt.Sprintf("  -%s, --%-19s %s\n", flag.Shorthand, flag.Name, flag.Usage)
			} else {
				filteredFlagUsages += fmt.Sprintf("      --%-19s %s\n", flag.Name, flag.Usage)
			}
		})

		// Only use filteredFlagUsages to avoid duplicates
		var combinedFlagUsages string = filteredFlagUsages

		// Execute the template
		return tmpl.Execute(os.Stdout, struct {
			*cobra.Command
			FlagCategories     []FlagCategory
			CombinedFlagUsages string
		}{
			Command:            cmd,
			FlagCategories:     allFlagCategories,
			CombinedFlagUsages: combinedFlagUsages,
		})
	})

	err := viper.BindPFlags(rootCmd.PersistentFlags())
	CheckErr(err)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if options.ConfigFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(options.ConfigFile)
	} else {
		cwd, err := os.Getwd()
		CheckErr(err)

		// Find home directory.
		home, err := os.UserHomeDir()
		CheckErr(err)

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

// CheckErr prints the msg and exits.
// If the msg is nil, it does nothing.
func CheckErr(msg interface{}) {
	if msg != nil {
		log.Fatal(msg)
	}
}

// CheckWarn prints the msg.
// If the msg is nil, it does nothing.
func CheckWarn(msg interface{}) {
	if msg != nil {
		log.Warn(msg)
	}
}

// ShowHelp prints the help for the command.
// If the msg is not nil, it prints the msg after the help.
func ShowHelp(cmd *cobra.Command, msg interface{}) {
	err := cmd.Help()
	CheckErr(err)

	if msg != nil {
		fmt.Println()
		log.Error(msg)
	}
	os.Exit(1)
}
