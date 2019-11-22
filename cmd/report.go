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
	"strings"

	"github.com/chef/go-libs/credentials"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

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

			chefClient, err := reporting.NewChefClient(cfg)
			if err != nil {
				return err
			}

			results, err := reporting.Nodes(cfg, chefClient.Search)
			if err != nil {
				return err
			}

			writeNodeReport(results)
			return nil
		},
	}
	cookbookStateCmd = &cobra.Command{
		Use:   "cookbook-state",
		Short: "Generates cookbook report that shows current remediation state and usage",
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

			results, err := reporting.CookbookState(cfg, chefClient.Cookbooks, chefClient.Search, reporting.ExecCommandRunner{})
			if err != nil {
				return err
			} else {
				writeCookbookStateReport(results)
			}

			return nil

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
	// adds the cookbook-state command as a sub-command of the report command
	// => chef-analyze report nodes
	reportCmd.AddCommand(cookbookStateCmd)
}

// TODO different output depending on flags or TTY?

func writeCookbookStateReport(records []*reporting.CookbookStateRecord) {
	var downloadErrors strings.Builder
	var usageFetchErrors strings.Builder
	for _, record := range records {
		var str strings.Builder

		str.WriteString(fmt.Sprintf("%v (%v) ", record.Name, record.Version))
		if record.CBDownloadError == nil {
			str.WriteString(fmt.Sprintf("%v violations, %v auto-correctable, ", record.Violations, record.Autocorrectable))
		} else {
			str.WriteString("violations unknown - see end of report, ")
			downloadErrors.WriteString(fmt.Sprintf(" - %s (%s): %v\n", record.Name, record.Version, record.CBDownloadError))
		}

		if record.UsageLookupError == nil {
			str.WriteString(fmt.Sprintf("%v nodes affected", len(record.Nodes)))
		} else {
			str.WriteString("node count unknown - see end of report ")
			downloadErrors.WriteString(fmt.Sprintf(" - %s (%s): %v\n", record.Name, record.Version, record.UsageLookupError))
		}

		fmt.Println(str.String())
	}

	fmt.Println()
	if downloadErrors.Len() > 0 {
		fmt.Println("Cookbook download errors prevented me from scanning some cookbooks:")
		fmt.Print(downloadErrors.String())
	}
	// TODO - Same here.  This woul donly happy because of an API failure.  They can't fix the API failure,
	//        and these errors won't really explain that very well in any case.
	if usageFetchErrors.Len() > 0 {
		fmt.Println("Node usage check errors prevented me from getting the number of nodes using some cookbooks:")
		fmt.Print(usageFetchErrors.String())
	}
}

func writeNodeReport(records []reporting.NodeReportItem) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Node Name", "Chef Version", "OS", "OS Version", "Cookbooks"})
	table.SetReflowDuringAutoWrap(true)
	table.SetRowLine(true)
	table.SetAutoWrapText(true)
	table.SetReflowDuringAutoWrap(true)
	table.SetBorder(true)
	for _, record := range records {
		table.Append(record.Array())
	}
	table.Render()
}
