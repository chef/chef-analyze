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
	"github.com/spf13/cobra"

	"github.com/chef/chef-analyze/pkg/credentials"
	"github.com/chef/chef-analyze/pkg/reporting"
)

var (
	reportCmd = &cobra.Command{
		Use:   "report",
		Short: "Generate reports about your Chef inventory",
	}
	reportCookbooksCmd = &cobra.Command{
		Use:   "cookbooks",
		Short: "Generates a cookbook oriented report",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			creds, err := credentials.FromViper(
				globalFlags.profile,
				overrideCredentials(),
			)
			if err != nil {
				return err
			}

			cfg := &reporting.Reporting{Credentials: creds}
			if globalFlags.noSSLverify {
				cfg.NoSSLVerify = true
			}

			return reporting.Cookbooks(cfg)
		},
	}
	reportNodesCmd = &cobra.Command{
		Use:   "nodes",
		Short: "Generates a nodes oriented report",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			creds, err := credentials.FromViper(
				globalFlags.profile,
				overrideCredentials(),
			)
			if err != nil {
				return err
			}

			cfg := &reporting.Reporting{Credentials: creds}
			if globalFlags.noSSLverify {
				cfg.NoSSLVerify = true
			}

			return reporting.Nodes(cfg)
		},
	}
)

func init() {
	// adds the cookbooks command as a sub-command of the report command
	// => chef-analyze report cookbooks
	reportCmd.AddCommand(reportCookbooksCmd)
	// adds the nodes command as a sub-command of the report command
	// => chef-analyze report nodes
	reportCmd.AddCommand(reportNodesCmd)
}
