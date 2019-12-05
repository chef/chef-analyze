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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/chef/chef-analyze/pkg/reporting"
)

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
