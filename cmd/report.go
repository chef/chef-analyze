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
	var (
		downloadErrors   strings.Builder
		usageFetchErrors strings.Builder
		cookstyleErrors  strings.Builder
	)
	for _, record := range records {
		var strBuilder strings.Builder

		// skip unused cookbooks
		if len(record.Nodes) == 0 && cookbookStateFlags.skipUnused {
			continue
		}

		strBuilder.WriteString(fmt.Sprintf("%v (%v) ", record.Name, record.Version))
		strBuilder.WriteString(fmt.Sprintf("%v violations, %v auto-correctable, %v nodes affected",
			record.NumOffenses(), record.NumCorrectable(), len(record.Nodes)),
		)

		if record.DownloadError != nil {
			strBuilder.WriteString("\nERROR: could not download cookbook (see end of report)")
			downloadErrors.WriteString(fmt.Sprintf(" - %s (%s): %v\n", record.Name, record.Version, record.DownloadError))
		} else if record.CookstyleError != nil {
			strBuilder.WriteString("\nERROR: could not run cookstyle (see end of report)")
			cookstyleErrors.WriteString(fmt.Sprintf(" - %s (%s): %v\n", record.Name, record.Version, record.CookstyleError))
		} else if record.UsageLookupError != nil {
			strBuilder.WriteString("\nERROR: unknown violations (see end of report)")
			usageFetchErrors.WriteString(fmt.Sprintf(" - %s (%s): %v\n", record.Name, record.Version, record.UsageLookupError))
		}

		// TODO @afiune write report to disk
		fmt.Println(strBuilder.String())
	}

	writeErrorBuilders(downloadErrors, cookstyleErrors, usageFetchErrors)
}

func writeDetailedCookbookStateReport(records []*reporting.CookbookStateRecord) {
	var (
		downloadErrors   strings.Builder
		usageFetchErrors strings.Builder
		cookstyleErrors  strings.Builder
	)
	for _, record := range records {
		var strBuilder strings.Builder

		// skip unused cookbooks
		if len(record.Nodes) == 0 && cookbookStateFlags.skipUnused {
			continue
		}

		strBuilder.WriteString(fmt.Sprintf("%v (%v) ", record.Name, record.Version))
		strBuilder.WriteString(fmt.Sprintf("%v violations, %v auto-correctable, %v nodes affected",
			record.NumOffenses(), record.NumCorrectable(), len(record.Nodes)),
		)

		strBuilder.WriteString("\nNodes affected: ")
		if len(record.Nodes) == 0 {
			strBuilder.WriteString("none")
		} else {
			strBuilder.WriteString(strings.Join(record.Nodes, ", "))
		}
		strBuilder.WriteString("\nFiles and offenses:")
		for _, f := range record.Files {
			if len(f.Offenses) == 0 {
				continue
			}
			strBuilder.WriteString(fmt.Sprintf("\n - %s:", f.Path))
			for _, o := range f.Offenses {
				strBuilder.WriteString(fmt.Sprintf("\n\t%s (%t) %s", o.CopName, o.Correctable, o.Message))
			}
		}

		if record.DownloadError != nil {
			strBuilder.WriteString("\nERROR: could not download cookbook (see end of report)")
			downloadErrors.WriteString(fmt.Sprintf(" - %s (%s): %v\n", record.Name, record.Version, record.DownloadError))
		} else if record.CookstyleError != nil {
			strBuilder.WriteString("\nERROR: could not run cookstyle (see end of report)")
			cookstyleErrors.WriteString(fmt.Sprintf(" - %s (%s): %v\n", record.Name, record.Version, record.CookstyleError))
		} else if record.UsageLookupError != nil {
			strBuilder.WriteString("\nERROR: unknown violations (see end of report)")
			usageFetchErrors.WriteString(fmt.Sprintf(" - %s (%s): %v\n", record.Name, record.Version, record.UsageLookupError))
		}

		// TODO @afiune write report to disk
		fmt.Println(strBuilder.String())
	}

	writeErrorBuilders(downloadErrors, cookstyleErrors, usageFetchErrors)
}

func writeDetailedCSV(records []*reporting.CookbookStateRecord) {
	var (
		strBuilder strings.Builder
		csvWriter  = csv.NewWriter(&strBuilder)
	)
	// table headers
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

	// TODO @afiune write report to disk
	fmt.Println(strBuilder.String())
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

func writeErrorBuilders(errBuilders ...strings.Builder) {
	firstMsg := true
	for _, errBldr := range errBuilders {
		if errBldr.Len() > 0 {
			if firstMsg {
				fmt.Fprintln(os.Stderr, "* ERROR(S) DETAILS:")
				firstMsg = false
			}
			fmt.Fprintln(os.Stderr, errBldr.String())
		}
	}
}
