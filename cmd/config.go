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

	"github.com/spf13/cobra"

	"github.com/chef/chef-analyze/pkg/dist"
)

var (
	configCmd = &cobra.Command{
		Use:    "config",
		Hidden: true, // this will avoid the command to be displayed in the help/usage message
		Short: fmt.Sprintf(
			"Manage your local %s configuration (default: $HOME/.chef/credentials)",
			dist.ServerProduct,
		),
	}
	configVerifyCmd = &cobra.Command{
		Use:   "verify",
		Short: fmt.Sprintf("Verify your %s configuration", dist.ServerProduct),
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, args []string) error {
			// TODO @afiune verify the config
			return nil
		},
	}
	configInitCmd = &cobra.Command{
		Use:   "init",
		Short: fmt.Sprintf("Initialize a local %s configuration", dist.ServerProduct),
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, args []string) error {
			// TODO @afiune help user initialize a local config file
			return nil
		},
	}
)

func init() {
	// adds the verify command as a sub-command of the config command
	// => chef-analyze config verify
	configCmd.AddCommand(configVerifyCmd)
	// adds the verify command as a sub-command of the config command
	// => chef-analyze config init
	configCmd.AddCommand(configInitCmd)
}
