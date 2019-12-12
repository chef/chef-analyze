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

func MakeCookbooksReportTXT(records []*reporting.CookbookRecord) *FormattedResult {
	var (
		errorBuilder strings.Builder
		strBuilder   strings.Builder
	)

	if len(records) == 0 {
		// nothing to do
		return &FormattedResult{"", ""}
	}

	for _, record := range records {

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

		for _, e := range record.Errors() {
			errorBuilder.WriteString(fmt.Sprintf(" - %s (%s): %v\n", record.Name, record.Version, e))
		}

	}
	return &FormattedResult{strBuilder.String(), errorBuilder.String()}
}
