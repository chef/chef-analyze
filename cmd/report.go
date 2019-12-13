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
	"path/filepath"
	"strings"
	"time"

	"github.com/chef/go-libs/credentials"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/chef/chef-analyze/pkg/formatter"
	"github.com/chef/chef-analyze/pkg/reporting"
)

// TODO @afiune make this configurable
const (
	AnalyzeCacheDir  = ".analyze-cache"
	repNameCookbooks = "cookbooks"
	repNameNodes     = "nodes"
	ErrExt           = "err"
	TxtExt           = "txt"
	CsvExt           = "csv"
)

var (
	timestamp  = time.Now().Format("20060102150405")
	reportsDir = filepath.Join(AnalyzeCacheDir, "reports")
	errorsDir  = filepath.Join(AnalyzeCacheDir, "errors")
	reportCmd  = &cobra.Command{
		Use:   "report",
		Short: "Generate reports from a Chef Infra Server",
	}
	reportCookbooksCmd = &cobra.Command{
		Use:   "cookbooks",
		Short: "Generates a cookbook oriented report",
		Args:  cobra.NoArgs,
		Long: `Generates cookbook oriented reports containing details about the number of
violations each cookbook has, which violations can be are auto-corrected and
the number of nodes using each cookbook.

These reports could take a long time to run depending on the number of cookbooks
to analyze and therefore reports will be written to disk. The location will be
provided when the report is generated.
`,
		RunE: func(_ *cobra.Command, _ []string) error {
			creds, err := credentials.FromViper(
				globalFlags.profile,
				overrideCredentials(),
			)

			if err != nil {
				return err
			}

			err = createOutputDirectories()
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
				cookbooksFlags.runCookstyle,
				cookbooksFlags.onlyUnused,
				cookbooksFlags.workers,
			)
			if err != nil {
				return err
			}

			var (
				formattedSummary = formatter.CookbooksReportSummary(cookbooksState)
				results          *formatter.FormattedResult
				ext              string
			)

			fmt.Println(formattedSummary.Report)

			switch reportsFlags.format {
			case "csv":
				ext = CsvExt
				results = formatter.MakeCookbooksReportCSV(cookbooksState)
			default:
				ext = TxtExt
				results = formatter.MakeCookbooksReportTXT(cookbooksState)
			}

			err = saveReport(repNameCookbooks, ext, results.Report)
			if err != nil {
				return err
			}
			err = saveErrorReport(repNameCookbooks, results.Errors)
			if err != nil {
				return err
			}

			return nil
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

			err = createOutputDirectories()
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

			fmt.Println("Analyzing nodes...")
			report, err := reporting.Nodes(chefClient.Search)
			if err != nil {
				return err
			}

			var (
				formattedSummary = formatter.NodesReportSummary(report)
				results          *formatter.FormattedResult
				ext              string
			)

			fmt.Println(formattedSummary.Report)

			switch reportsFlags.format {
			case "csv":
				ext = CsvExt
				results = formatter.MakeNodesReportCSV(report)
			default:
				ext = TxtExt
				results = formatter.MakeNodesReportTXT(report)
			}

			err = saveReport(repNameNodes, ext, results.Report)
			if err != nil {
				return err
			}
			err = saveErrorReport(repNameNodes, results.Errors)
			if err != nil {
				return err
			}

			return nil
		},
	}
	cookbooksFlags struct {
		onlyUnused   bool
		runCookstyle bool
		workers      int
	}
	reportsFlags struct {
		format string
	}
)

func init() {
	// global report commands flags
	reportCmd.PersistentFlags().StringVarP(
		&reportsFlags.format,
		"format", "f", "txt",
		"output format: txt is human readable, csv is machine readable",
	)

	// cookbooks cmd flags
	reportCookbooksCmd.PersistentFlags().IntVarP(
		&cookbooksFlags.workers,
		"workers", "w", 50,
		"maximum number of parallel workers at once",
	)
	reportCookbooksCmd.PersistentFlags().BoolVarP(
		&cookbooksFlags.onlyUnused,
		"only-unused", "u", false,
		"generate a report with only cookbooks that are not applied to any node",
	)
	reportCookbooksCmd.PersistentFlags().BoolVarP(
		&cookbooksFlags.runCookstyle,
		"verify-upgrade", "v", false,
		"verify the upgrade compatibility of every cookbook",
	)
	// adds the cookbooks command as a sub-command of the report command
	// => chef-analyze report cookbooks
	reportCmd.AddCommand(reportCookbooksCmd)

	// adds the nodes command as a sub-command of the report command
	// => chef-analyze report nodes
	reportCmd.AddCommand(reportNodesCmd)

}

func createOutputDirectories() error {
	err := os.MkdirAll(errorsDir, os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "unable to create reports/ directory")
	}
	err = os.MkdirAll(reportsDir, os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "unable to create errors/ directory")
	}
	return nil
}

func saveErrorReport(baseName string, content string) error {
	if content == "" {
		return nil
	}

	var (
		reportName      = fmt.Sprintf("%s-%s.%s", baseName, timestamp, "err")
		reportPath      = filepath.Join(errorsDir, reportName)
		reportFile, err = os.Create(reportPath)
	)
	if err != nil {
		return errors.Wrap(err, "unable to save errors report")
	}

	reportFile.WriteString(content)
	reportFile.Close()

	fmt.Printf("Error report saved to %s\n", reportPath)
	return nil
}

func saveReport(baseName string, ext string, content string) error {
	if len(content) == 0 {
		return nil
	}

	var (
		reportName      = fmt.Sprintf("%s-%s.%s", baseName, timestamp, ext)
		reportPath      = filepath.Join(reportsDir, reportName)
		reportFile, err = os.Create(reportPath) // create a new report file
	)
	if err != nil {
		return errors.Wrapf(err, "unable to save %s report", baseName)
	}

	reportFile.WriteString(content)
	reportFile.Close()

	fmt.Printf("%s report saved to %s\n", strings.Title(baseName), reportPath)
	return nil
}
