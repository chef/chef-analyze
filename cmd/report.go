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
	"encoding/csv"
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
	cookbookStateFlags struct {
		detailed   bool
		skipUnused bool
		format     string
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

			cookbookState, err := reporting.NewCookbookState(chefClient.Cookbooks, chefClient.Search, reporting.ExecCookstyleRunner{}, cookbookStateFlags.skipUnused)
			if err != nil {
				return err
			}

			if cookbookStateFlags.detailed {
				switch cookbookStateFlags.format {
				case "csv":
					writeDetailedCSV(cookbookState.Records)
				default:
					writeDetailedCookbookStateReport(cookbookState.Records)
				}
				return nil
			}

			writeCookbookStateReport(cookbookState.Records)
			return nil
		},
	}
)

func init() {
	cookbookStateCmd.PersistentFlags().BoolVarP(
		&cookbookStateFlags.detailed,
		"detailed", "d", false,
		"include detailed information about cookbook violations",
	)
	cookbookStateCmd.PersistentFlags().BoolVarP(
		&cookbookStateFlags.skipUnused,
		"skip-unused", "u", false,
		"do not include unused cookbooks and versions that are not applied to any nodes",
	)
	cookbookStateCmd.PersistentFlags().StringVarP(
		&cookbookStateFlags.format,
		"format", "f", "txt",
		"output format: txt is human readable, csv is machine readable",
	)
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
	var cookstyleErrors strings.Builder
	for _, record := range records {
		var str strings.Builder

		// skip unused cookbooks
		if len(record.Nodes) == 0 && cookbookStateFlags.skipUnused {
			continue
		}

		str.WriteString(fmt.Sprintf("%v (%v) ", record.Name, record.Version))
		if record.DownloadError != nil {
			str.WriteString("could not download cookbook - see end of report, ")
			downloadErrors.WriteString(fmt.Sprintf(" - %s (%s): %v\n", record.Name, record.Version, record.DownloadError))
		} else if record.CookstyleError != nil {
			str.WriteString("could not run Cookstyle - see end of report ")
			cookstyleErrors.WriteString(fmt.Sprintf(" - %s (%s): %v\n", record.Name, record.Version, record.CookstyleError))
		} else if record.UsageLookupError != nil {
			str.WriteString("violations unknown - see end of report, ")
			usageFetchErrors.WriteString(fmt.Sprintf(" - %s (%s): %v\n", record.Name, record.Version, record.UsageLookupError))
		}

		str.WriteString(fmt.Sprintf("%v violations, %v auto-correctable, ", record.NumOffenses(), record.NumCorrectable()))
		str.WriteString(fmt.Sprintf("%v nodes affected", len(record.Nodes)))

		fmt.Println(str.String())
	}

	if downloadErrors.Len() > 0 {
		fmt.Println("Cookbook download errors prevented me from scanning some cookbooks:")
		fmt.Print(downloadErrors.String())
	}
	if cookstyleErrors.Len() > 0 {
		fmt.Println("Cookstyle execution errors prevented me from scanning some cookbooks:")
		fmt.Print(downloadErrors.String())
	}
	if usageFetchErrors.Len() > 0 {
		fmt.Println("Node usage check errors prevented me from getting the number of nodes using some cookbooks:")
		fmt.Print(usageFetchErrors.String())
	}
}

func writeDetailedCookbookStateReport(records []*reporting.CookbookStateRecord) {
	var downloadErrors strings.Builder
	var usageFetchErrors strings.Builder
	var cookstyleErrors strings.Builder
	for _, record := range records {
		var str strings.Builder

		// skip unused cookbooks
		if len(record.Nodes) == 0 && cookbookStateFlags.skipUnused {
			continue
		}

		str.WriteString(fmt.Sprintf("%v (%v)\nNodesAffected: ", record.Name, record.Version))
		if len(record.Nodes) == 0 {
			str.WriteString("None\n")
		} else {
			str.WriteString(strings.Join(record.Nodes, ", ") + "\n")
		}
		str.WriteString("Files and offenses:\n")
		for _, f := range record.Files {
			if len(f.Offenses) == 0 {
				continue
			}
			str.WriteString(fmt.Sprintf(" - %s:\n", f.Path))
			for _, o := range f.Offenses {
				str.WriteString(fmt.Sprintf("\t%s (%t) %s\n", o.CopName, o.Correctable, o.Message))
			}
		}
		if record.DownloadError != nil {
			str.WriteString("could not download cookbook - see end of report, ")
			downloadErrors.WriteString(fmt.Sprintf(" - %s (%s): %v\n", record.Name, record.Version, record.DownloadError))
		} else if record.CookstyleError != nil {
			str.WriteString("could not run Cookstyle - see end of report, ")
			cookstyleErrors.WriteString(fmt.Sprintf(" - %s (%s): %v\n", record.Name, record.Version, record.CookstyleError))
		} else if record.UsageLookupError != nil {
			str.WriteString("violations unknown - see end of report, ")
			usageFetchErrors.WriteString(fmt.Sprintf(" - %s (%s): %v\n", record.Name, record.Version, record.UsageLookupError))
		}

		str.WriteString(fmt.Sprintf("%v violations, %v auto-correctable, ", record.NumOffenses(), record.NumCorrectable()))
		str.WriteString(fmt.Sprintf("%v nodes affected", len(record.Nodes)))

		fmt.Println(str.String())
	}

	fmt.Println()
	if downloadErrors.Len() > 0 {
		fmt.Println("Cookbook download errors prevented me from scanning some cookbooks:")
		fmt.Print(downloadErrors.String())
	}
	if cookstyleErrors.Len() > 0 {
		fmt.Println("Cookstyle execution errors prevented me from scanning some cookbooks:")
		fmt.Print(downloadErrors.String())
	}
	// TODO - Same here.  This woul donly happy because of an API failure.  They can't fix the API failure,
	//        and these errors won't really explain that very well in any case.
	if usageFetchErrors.Len() > 0 {
		fmt.Println("Node usage check errors prevented me from getting the number of nodes using some cookbooks:")
		fmt.Print(usageFetchErrors.String())
	}
}

func writeDetailedCSV(records []*reporting.CookbookStateRecord) {
	var str strings.Builder
	csvWriter := csv.NewWriter(&str)
	csvWriter.Write([]string{"Cookbook Name", "Version", "File", "Offense", "Automatically Correctable", "Message", "Nodes"})
	for _, record := range records {
		// skip unused cookbooks
		if len(record.Nodes) == 0 && cookbookStateFlags.skipUnused {
			continue
		}

		firstRow := []string{record.Name, record.Version, "", "", "", "", strings.Join(record.Nodes, " ")}
		firstRowPopulated := false
		for _, file := range record.Files {
			if len(file.Offenses) == 0 {
				continue
			}
			if firstRowPopulated == false {
				firstRow[2] = file.Path
				firstOffense := file.Offenses[0]
				file.Offenses = file.Offenses[1:]
				firstRow[3] = firstOffense.CopName
				if firstOffense.Correctable {
					firstRow[4] = "Y"
				} else {
					firstRow[4] = "N"
				}
				firstRow[5] = firstOffense.Message
				csvWriter.Write(firstRow)
				firstRowPopulated = true
			} else {
				for _, offense := range file.Offenses {
					row := []string{"", "", "", offense.CopName, "", offense.Message, ""}
					if offense.Correctable {
						row[4] = "Y"
					} else {
						row[4] = "N"
					}
					csvWriter.Write(row)
				}
			}
		}
	}
	csvWriter.Flush()
	fmt.Println(str.String())
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
