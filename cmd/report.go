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
	"github.com/chef/go-libs/credentials"
	"github.com/spf13/cobra"

	"github.com/chef/chef-analyze/pkg/formatter"
	"github.com/chef/chef-analyze/pkg/reporting"
)

var (
	reportCmd = &cobra.Command{
		Use:   "report",
		Short: "Generate reports from a Chef Infra Server",
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

			chefClient, err := reporting.NewChefClient(cfg)
			if err != nil {
				return err
			}

			cookbooksState, err := reporting.NewCookbooks(
				chefClient.Cookbooks,
				chefClient.Search,
				cookbooksFlags.skipUnused,
			)
			if err != nil {
				return err
			}

			formatter.PrintCookbooksReportSummary(cookbooksState.Records)

			switch cookbooksFlags.format {
			case "csv":
				return formatter.StoreCookbooksReportCSV(cookbooksState.Records)
			default:
				return formatter.StoreCookbooksReportTXT(cookbooksState.Records)
			}
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

			chefClient, err := reporting.NewChefClient(cfg)
			if err != nil {
				return err
			}

			results, err := reporting.Nodes(cfg, chefClient.Search)
			if err != nil {
				return err
			}

			formatter.WriteNodeReport(results)
			return nil
		},
	}
	cookbooksFlags struct {
		skipUnused bool
		format     string
	}
)

func init() {
	reportCookbooksCmd.PersistentFlags().BoolVarP(
		&cookbooksFlags.skipUnused,
		"skip-unused", "u", false,
		"do not include unused cookbooks and versions that are not applied to any nodes",
	)
	reportCookbooksCmd.PersistentFlags().StringVarP(
		&cookbooksFlags.format,
		"format", "f", "txt",
		"output format: txt is human readable, csv is machine readable",
	)
	// adds the cookbooks command as a sub-command of the report command
	// => chef-analyze report cookbooks
	reportCmd.AddCommand(reportCookbooksCmd)
	// adds the nodes command as a sub-command of the report command
	// => chef-analyze report nodes
	reportCmd.AddCommand(reportNodesCmd)
}
