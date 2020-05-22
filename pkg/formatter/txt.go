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

// MakeCookbooksReportTXT text output of long, non-summarize report
func MakeCookbooksReportTXT(state *reporting.CookbooksReport) *FormattedResult {
	var (
		errorBuilder strings.Builder
		strBuilder   strings.Builder
	)
	if state == nil {
		return &FormattedResult{strBuilder.String(), ""}
	}

	if len(state.NodeFilter) != 0 {
		strBuilder.WriteString(fmt.Sprintf("Node filter applied: %s\n", state.NodeFilter))
	}
	if len(state.Records) == 0 {
		// nothing to do
		return &FormattedResult{strBuilder.String(), ""}
	}

	sortCookbookRecords(state.Records)

	for _, record := range state.Records {

		if record.PolicyGroup != "" {
			strBuilder.WriteString(fmt.Sprintf("> Cookbook: %v (policy %s, revision %s)\n", record.Name, record.Policy, record.PolicyVer))
			strBuilder.WriteString(fmt.Sprintf("  Policy Group: %s\n", record.PolicyGroup))
		} else {
			strBuilder.WriteString(fmt.Sprintf("> Cookbook: %v (%v)\n", record.Name, record.Version))
			strBuilder.WriteString(fmt.Sprintf("  Policy Group: none\n"))
		}

		if record.NumNodesAffected() == 0 {
			strBuilder.WriteString("  Nodes affected: none\n")
		} else {
			strBuilder.WriteString("  Nodes affected: ")
			strBuilder.WriteString(strings.Join(record.Nodes, ", "))
			strBuilder.WriteString("\n")
		}

		if state.RunCookstyle {
			strBuilder.WriteString(fmt.Sprintf("  Violations: %v\n", record.NumOffenses()))
			strBuilder.WriteString(fmt.Sprintf("  Auto correctable: %v\n", record.NumCorrectable()))
			strBuilder.WriteString("  Files and offenses:")
			for _, f := range record.Files {
				if len(f.Offenses) == 0 {
					continue
				}

				strBuilder.WriteString(fmt.Sprintf("\n   - %s:", f.Path))
				for _, o := range f.Offenses {
					strBuilder.WriteString(fmt.Sprintf("\n\t%s (%t) %s", o.CopName, o.Correctable, o.Message))
				}
			}

			if record.NumOffenses() == 0 {
				strBuilder.WriteString(" none\n")
			} else {
				strBuilder.WriteString("\n")
			}
		}

		for _, e := range record.Errors() {
			if record.PolicyGroup != "" {
				errorBuilder.WriteString(fmt.Sprintf(" - %s (PolicyGroup %s, Policy %s, PolicyRevision %s): %v\n", record.Name, record.PolicyGroup, record.Policy, record.PolicyVer, e))
			} else {
				errorBuilder.WriteString(fmt.Sprintf(" - %s (%s): %v\n", record.Name, record.Version, e))
			}
		}

	}
	return &FormattedResult{strBuilder.String(), errorBuilder.String()}
}

// MakeNodesReportTXT text output of long, non-summarize report
func MakeNodesReportTXT(records []*reporting.NodeReportItem, nodeFilter string) *FormattedResult {
	var (
		errorBuilder strings.Builder
		strBuilder   strings.Builder
	)
	if len(nodeFilter) > 0 {
		strBuilder.WriteString(fmt.Sprintf("Node Filter Applied: %s\n", nodeFilter))
	}

	if len(records) == 0 {
		// nothing to do
		return &FormattedResult{strBuilder.String(), ""}
	}

	sortNodeRecords(records)

	for _, record := range records {
		strBuilder.WriteString(fmt.Sprintf("> Node: %s\n", record.Name))
		strBuilder.WriteString(
			fmt.Sprintf("  Chef Version: %s\n",
				stringOrUnknownPlaceholder(record.ChefVersion)),
		)
		strBuilder.WriteString(
			fmt.Sprintf("  Operating System: %s\n",
				stringOrUnknownPlaceholder(record.OSVersionPretty())),
		)

		strBuilder.WriteString(
			fmt.Sprintf("  Policy Group: %s\n",
				stringOrUnknownPlaceholder(record.GetPolicyGroup())),
		)

		strBuilder.WriteString(
			fmt.Sprintf("  Policy: %s\n",
				stringOrUnknownPlaceholder(record.GetPolicyWithRev())),
		)

		if len(record.CookbooksList()) == 0 {
			strBuilder.WriteString("  Cookbooks Applied: none\n")
		} else {
			strBuilder.WriteString("  Cookbooks Applied (alphanumeric order): ")
			cookbooksString := strings.Join(record.CookbooksList(), ", ")
			if record.GetPolicyGroup() != "no group" {
				cookbooksString = stringReplace(`\([\d,.]+\)`, cookbooksString, "")
			}
			strBuilder.WriteString(cookbooksString)
			strBuilder.WriteString("\n")
		}
	}

	return &FormattedResult{strBuilder.String(), errorBuilder.String()}
}
