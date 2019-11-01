//
// Copyright 2019 Chef Software, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/chef/chef-analyze/pkg/config"
)

var (
	globalFlags struct {
		credsFile     string
		clientName    string
		clientKey     string
		chefServerUrl string
		profile       string
		noSSLCheck    bool
	}
	rootCmd = &cobra.Command{
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
	rootCmd.PersistentFlags().StringVarP(
		&globalFlags.credsFile,
		"credentials", "c", "",
		"Chef credentials file (default $HOME/.chef/credentials)",
	)
	rootCmd.PersistentFlags().StringVarP(
		&globalFlags.clientName,
		"client_name", "n", "",
		"Chef Infra Server API client username",
	)
	rootCmd.PersistentFlags().StringVarP(
		&globalFlags.clientKey,
		"client_key", "k", "",
		"Chef Infra Server API client key",
	)
	rootCmd.PersistentFlags().StringVarP(
		&globalFlags.chefServerUrl,
		"chef_server_url", "s", "",
		"Chef Infra Server URL",
	)
	rootCmd.PersistentFlags().StringVarP(
		&globalFlags.profile,
		"profile", "p", "default",
		"Chef Infra Server URL",
	)
	rootCmd.PersistentFlags().BoolVarP(
		&globalFlags.noSSLCheck,
		"no-ssl-check", "k", false,
		"Disable SSL certificate verification",
	)
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
	if globalFlags.credsFile != "" {
		// Use credentials file from the flag
		viper.SetConfigFile(globalFlags.credsFile)
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

			if !hasMinimumParams() && !isHelpCommand() {
				fmt.Printf("Error: %s\n", MissingMinimumParametersErr)
				rootCmd.Usage()
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
	if globalFlags.chefServerUrl != "" &&
		globalFlags.clientName != "" &&
		globalFlags.clientKey != "" {
		return true
	}

	return false
}
func isHelpCommand() bool {
	if len(os.Args) <= 1 {
		return false
	}
	if os.Args[1] == "help" {
		return true
	}
	return false
}

// overrides the credentials from the viper bound flags
func overrideCredentials() config.OverrideFunc {
	return func(c *config.Config) {
		if globalFlags.clientName != "" {
			c.ClientName = globalFlags.clientName
		}
		if globalFlags.clientKey != "" {
			c.ClientKey = globalFlags.clientKey
		}
		if globalFlags.chefServerUrl != "" {
			c.ChefServerUrl = globalFlags.chefServerUrl
		}
		if globalFlags.noSSLCheck {
			c.SkipSSL = true
		}
	}
}
