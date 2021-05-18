package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var cfgFile string
var circleAPIToken string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "circleci-helper",
	Short: "CircleCI helper binary",
	Long: `CircleCI helper binary that allows performing higher level logic
`,
}

// Execute provides a method to execute the root command
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.circleci-helper.yaml)")
	rootCmd.PersistentFlags().StringVar(&circleAPIToken, "token", "", "CircleCI API token")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".circleci-helper" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".circleci-helper")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
