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
	"strings"

	"github.com/chef/chef-analyze/pkg/reporting"
)

// TODO different output depending on flags or TTY?
func MakeCookbooksReportSummary(records []*reporting.CookbookRecord) string {
	if len(records) == 0 {
		// TODO @afiune what is the right text here?
		fmt.Println("There are no cookbook records to generate a report")
		return ""
	}

	var strBuilder strings.Builder
	fmt.Printf("\n-- REPORT SUMMARY --\n\n")
	for _, record := range records {

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
		strBuilder.WriteString("\n")

	}
	return strBuilder.String()
}
