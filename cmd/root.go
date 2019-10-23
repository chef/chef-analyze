package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/chef/chef-analyze/pkg/config"
)

var (
	credsFile string
	rootCmd   = &cobra.Command{
		Use:   "chef-analyze",
		Short: "Analyze your Chef inventory",
		Long: `Analyze your Chef inventory by generating reports to understand the effort
to upgrade to the latest version of the Chef tools, automatically fixing
violations and/or deprecations, and generating Effortless packages.
`,
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&credsFile, "credentials", "c", "", "Chef credentials file (default $HOME/.chef/credentials)")
	rootCmd.PersistentFlags().StringP("client_name", "n", "", "Chef Infra Server API client username")
	rootCmd.PersistentFlags().StringP("client_key", "k", "", "Chef Infra Server API client key")
	rootCmd.PersistentFlags().StringP("chef_server_url", "s", "", "Chef Infra Server URL")
	rootCmd.PersistentFlags().StringP("profile", "p", "default", "Chef Infra Server URL")
	// @afiune we can't use viper to bind the flags since our config doesn't really match
	// any valid toml structure. (that is, the .chef/credentials toml file)
	//
	// TODO: revisit
	//viper.BindPFlag("client_name", rootCmd.PersistentFlags().Lookup("client_name"))
	//viper.BindPFlag("client_key", rootCmd.PersistentFlags().Lookup("client_key"))
	//viper.BindPFlag("chef_server_url", rootCmd.PersistentFlags().Lookup("chef_server_url"))
	//viper.BindPFlag("profile", rootCmd.PersistentFlags().Lookup("profile"))

	// adds the report command from 'cmd/report.go'
	rootCmd.AddCommand(reportCmd)
	// adds the config command from 'cmd/config.go'
	rootCmd.AddCommand(configCmd)
}

func initConfig() {
	if credsFile != "" {
		// Use credentials file from the flag
		viper.SetConfigFile(credsFile)
	} else {
		// Find the credentials and pass it to viper
		credsFile, err := config.FindCredentialsFile()
		// @afiune we don't exit with and error code here because if we do
		// the user will never be able to fix the config with the commands:
		// $ chef-analyze config init
		//
		// This verification has been moved to config.FromViper()
		if err == nil {
			viper.SetConfigFile(credsFile)
		} else {
			if !hasMinimumParams() {
				// throw error
				fmt.Println("Error: ---")
				os.Exit(-1)
			}
			//debug("Unable to file credentials:  %s", err.Error())
		}
	}

	viper.SetConfigType("toml")
	viper.AutomaticEnv()
}

// tells you if the command was run with the minimum parameters for
// this tool to work, with or without credentials config
// TODO @afiune revisit
func hasMinimumParams() bool {
	if flagString("chef_server_url") != "" &&
		flagString("client_name") != "" &&
		flagString("client_key") != "" {
		return true
	}

	return false
}
func flagString(name string) string {
	f := rootCmd.PersistentFlags().Lookup(name)
	if f == nil {
		return ""
	}
	return f.Value.String()
}

// overrides the credentials from the viper bound flags
func overrideCredentials() config.OverrideFunc {
	return func(c *config.Config) {
		if flagString("client_name") != "" {
			c.ClientName = flagString("client_name")
		}
		if flagString("client_key") != "" {
			c.ClientKey = flagString("client_key")
		}
		if flagString("chef_server_url") != "" {
			c.ChefServerUrl = flagString("chef_server_url")
		}
	}
}
