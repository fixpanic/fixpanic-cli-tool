package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	version string
	commit  string
	date    string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "fixpanic",
	Short: "Fixpanic CLI - Deploy and manage Fixpanic agents",
	Long: `Fixpanic CLI is a command-line tool for deploying and managing Fixpanic agents
on customer servers. It handles agent installation, configuration, and lifecycle management.

The CLI downloads and manages the connectivity layer binary, sets up systemd services,
and provides commands for testing and validation.`,
	Version: "dev",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

// SetVersionInfo sets the version information for the CLI
func SetVersionInfo(v, c, d string) {
	version = v
	commit = c
	date = d

	// Update the root command version
	if v != "" && v != "dev" {
		rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", v, c, d)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.fixpanic.yaml)")
	rootCmd.PersistentFlags().String("socket-server", "socket.fixpanic.com:8080", "Socket server address")
	viper.BindPFlag("socket_server", rootCmd.PersistentFlags().Lookup("socket-server"))
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

		// Search config in home directory with name ".fixpanic" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".fixpanic")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
