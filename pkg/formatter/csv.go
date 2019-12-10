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
	"strings"

	"github.com/chef/chef-analyze/pkg/reporting"
)

func MakeCookbooksReportCSV(records []*reporting.CookbookRecord) *FormattedResult {
	var (
		strBuilder strings.Builder
		errBuilder strings.Builder
		csvWriter  = csv.NewWriter(&strBuilder)
	)

	if len(records) == 0 {
		return &FormattedResult{"", ""}
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

		if record.DownloadError != nil {
			errBuilder.WriteString(
				fmt.Sprintf(" - %s (%s): %v\n", record.Name, record.Version, record.DownloadError))
		}
		if record.CookstyleError != nil {
			errBuilder.WriteString(
				fmt.Sprintf(" - %s (%s): %v\n", record.Name, record.Version, record.CookstyleError))
		}
		if record.UsageLookupError != nil {
			errBuilder.WriteString(
				fmt.Sprintf(" - %s (%s): %v\n", record.Name, record.Version, record.UsageLookupError))
		}
	}

	csvWriter.Flush()
	return &FormattedResult{strBuilder.String(), errBuilder.String()}
}
