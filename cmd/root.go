//
// Copyright 2019 Chef Software, Inc.
// Author: Salim Afiune <afiune@chef.io>
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

	"github.com/chef/go-libs/credentials"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/chef/chef-analyze/pkg/dist"
)

var (
	rootCmd = &cobra.Command{
		Use:   dist.Exec,
		Short: "A CLI to analyze artifacts from a Chef Infra Server",
		Long: `Analyze your Chef Infra Server artifacts to understand the effort to upgrade
your infrastructure by generating reports, automatically fixing violations
and/or deprecations, and generating Effortless packages.
`,
	}
)

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

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
	// adds the upload command from 'cmd/upload.go'
	rootCmd.AddCommand(uploadCmd)
	// adds the session command from 'cmd/session.go'
	rootCmd.AddCommand(sessionCmd)
}

func initConfig() {
	if reportsFlags.credsFile != "" {
		// Use credentials file from the flag
		viper.SetConfigFile(reportsFlags.credsFile)
	} else {
		// Find the credentials and pass it to viper
		credsFile, err := credentials.FindCredentialsFile()
		// @afiune we don't exit with and error code here because if we do
		// the user will never be able to fix the config with the commands:
		// $ chef-analyze config init
		//
		// This verification has been moved to config.FromViper()
		if err == nil {
			viper.SetConfigFile(credsFile)
		} else {

			if !hasMinimumParams() && isReportCommand() {
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
	if reportsFlags.chefServerURL != "" &&
		reportsFlags.clientName != "" &&
		reportsFlags.clientKey != "" {
		return true
	}

	return false
}
func isReportCommand() bool {
	if len(os.Args) <= 1 {
		return false
	}
	if os.Args[1] == "report" {
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
func overrideCredentials() credentials.OverrideFunc {
	return func(c *credentials.Credentials) {
		if reportsFlags.clientName != "" {
			c.ClientName = reportsFlags.clientName
		}
		if reportsFlags.clientKey != "" {
			c.ClientKey = reportsFlags.clientKey
		}
		if reportsFlags.chefServerURL != "" {
			c.ChefServerUrl = reportsFlags.chefServerURL
		}
	}
}
