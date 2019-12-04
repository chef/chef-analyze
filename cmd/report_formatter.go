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

// @afiune move this code to a new pkg/formatter
package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"

	"github.com/chef/chef-analyze/pkg/reporting"
)

// TODO @afiune make this configurable
const AnalyzeCacheDir = ".analyze-cache"

// TODO different output depending on flags or TTY?
func PrintCookbooksReportSummary(records []*reporting.CookbookRecord) {
	if len(records) == 0 {
		// TODO @afiune what is the right text here?
		fmt.Println("No cookbooks were found to analyze")
		return
	}

	fmt.Println("\n-- REPORT SUMMARY --\n")
	for _, record := range records {
		var strBuilder strings.Builder

		strBuilder.WriteString(fmt.Sprintf("%v (%v) ", record.Name, record.Version))
		strBuilder.WriteString(fmt.Sprintf("%v violations, %v auto-correctable, %v nodes affected",
			record.NumOffenses(), record.NumCorrectable(), len(record.Nodes)),
		)

		if record.DownloadError != nil {
			strBuilder.WriteString("\nERROR: could not download cookbook (see end of report)")
		} else if record.CookstyleError != nil {
			strBuilder.WriteString("\nERROR: could not run cookstyle (see end of report)")
		} else if record.UsageLookupError != nil {
			strBuilder.WriteString("\nERROR: unknown violations (see end of report)")
		}

		fmt.Println(strBuilder.String())
	}
}

func StoreCookbooksReportTXT(records []*reporting.CookbookRecord) error {
	var (
		downloadErrors   strings.Builder
		usageFetchErrors strings.Builder
		cookstyleErrors  strings.Builder
		reportsDir       = filepath.Join(AnalyzeCacheDir, "reports")
		timestamp        = time.Now().Format("20060102150405")
		reportName       = fmt.Sprintf("cookbooks-%s.txt", timestamp)
		reportPath       = filepath.Join(reportsDir, reportName)
	)

	if len(records) == 0 {
		// nothing to do
		return nil
	}

	// create reports directory
	err := os.MkdirAll(reportsDir, os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "unable to create reports/ directory")
	}

	// create a new report file
	reportFile, err := os.Create(reportPath)
	if err != nil {
		return errors.Wrap(err, "unable to create report")
	}

	anyError := false
	for _, record := range records {
		var strBuilder strings.Builder

		strBuilder.WriteString(fmt.Sprintf("> Cookbook: %v (%v)\n", record.Name, record.Version))
		strBuilder.WriteString(fmt.Sprintf("  Violations: %v\n", record.NumOffenses()))
		strBuilder.WriteString(fmt.Sprintf("  Auto correctable: %v\n", record.NumCorrectable()))

		strBuilder.WriteString("  Nodes affected: ")
		if len(record.Nodes) == 0 {
			strBuilder.WriteString("none")
		} else {
			strBuilder.WriteString(strings.Join(record.Nodes, ", "))
		}

		anyOffense := false
		strBuilder.WriteString("\n  Files and offenses:")
		for _, f := range record.Files {
			if len(f.Offenses) == 0 {
				continue
			}

			// at least one offense exist
			anyOffense = true

			strBuilder.WriteString(fmt.Sprintf("\n   - %s:", f.Path))
			for _, o := range f.Offenses {
				strBuilder.WriteString(fmt.Sprintf("\n\t%s (%t) %s", o.CopName, o.Correctable, o.Message))
			}
		}

		if anyOffense {
			strBuilder.WriteString("\n")
		} else {
			strBuilder.WriteString(" none\n")
		}

		// verify errors
		// TODO @afiune we could refactor this to not be duplicated on every report
		if record.DownloadError != nil {
			anyError = true
			downloadErrors.WriteString(
				fmt.Sprintf(" - %s (%s): %v\n", record.Name, record.Version, record.DownloadError))
		}
		if record.CookstyleError != nil {
			anyError = true
			cookstyleErrors.WriteString(
				fmt.Sprintf(" - %s (%s): %v\n", record.Name, record.Version, record.CookstyleError))
		}
		if record.UsageLookupError != nil {
			anyError = true
			usageFetchErrors.WriteString(
				fmt.Sprintf(" - %s (%s): %v\n", record.Name, record.Version, record.UsageLookupError))
		}

		reportFile.WriteString(strBuilder.String())
	}

	fmt.Printf("\nCookbooks report generated at: \n  => %s\n", reportPath)
	if anyError {
		return StoreErrorsFromBuilders(timestamp, downloadErrors, cookstyleErrors, usageFetchErrors)
	}

	return nil
}

func StoreCookbooksReportCSV(records []*reporting.CookbookRecord) error {
	var (
		downloadErrors   strings.Builder
		usageFetchErrors strings.Builder
		cookstyleErrors  strings.Builder
		strBuilder       strings.Builder
		csvWriter        = csv.NewWriter(&strBuilder)
		reportsDir       = filepath.Join(AnalyzeCacheDir, "reports")
		timestamp        = time.Now().Format("20060102150405")
		reportName       = fmt.Sprintf("cookbooks-%s.csv", timestamp)
		reportPath       = filepath.Join(reportsDir, reportName)
	)

	if len(records) == 0 {
		// nothing to do
		return nil
	}

	// create reports directory
	err := os.MkdirAll(reportsDir, os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "unable to create reports/ directory")
	}

	// create a new report file
	reportFile, err := os.Create(reportPath)
	if err != nil {
		return errors.Wrap(err, "unable to create report")
	}

	// table headers
	csvWriter.Write([]string{
		"Cookbook Name",
		"Version",
		"File",
		"Offense",
		"Automatically Correctable",
		"Message",
		"Nodes"})

	anyError := false
	for _, record := range records {
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

		// verify errors
		// TODO @afiune we could refactor this to not be duplicated on every report
		if record.DownloadError != nil {
			anyError = true
			downloadErrors.WriteString(
				fmt.Sprintf(" - %s (%s): %v\n", record.Name, record.Version, record.DownloadError))
		}
		if record.CookstyleError != nil {
			anyError = true
			cookstyleErrors.WriteString(
				fmt.Sprintf(" - %s (%s): %v\n", record.Name, record.Version, record.CookstyleError))
		}
		if record.UsageLookupError != nil {
			anyError = true
			usageFetchErrors.WriteString(
				fmt.Sprintf(" - %s (%s): %v\n", record.Name, record.Version, record.UsageLookupError))
		}

	}

	csvWriter.Flush()
	reportFile.WriteString(strBuilder.String())
	fmt.Printf("\nCookbooks report generated at: \n  => %s\n", reportPath)

	if anyError {
		return StoreErrorsFromBuilders(timestamp, downloadErrors, cookstyleErrors, usageFetchErrors)
	}

	return nil
}

func StoreErrorsFromBuilders(timestamp string, errBuilders ...strings.Builder) error {
	var (
		errorsDir  = filepath.Join(AnalyzeCacheDir, "errors")
		errorsName = fmt.Sprintf("cookbooks-%s.err", timestamp)
		errorsPath = filepath.Join(errorsDir, errorsName)
	)

	// create reports directory
	err := os.MkdirAll(errorsDir, os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "unable to create reports/ directory")
	}

	// create a new errors file
	errorsFile, err := os.Create(errorsPath)
	if err != nil {
		return errors.Wrap(err, "unable to create report")
	}

	for _, errBldr := range errBuilders {
		if errBldr.Len() > 0 {
			errorsFile.WriteString(errBldr.String())
		}
	}

	fmt.Println("Error(s) found during the report generation, find more details at:")
	fmt.Printf("  => %s\n", errorsPath)
	return nil
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
