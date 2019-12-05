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

package formatter

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/chef/chef-analyze/pkg/reporting"
)

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
