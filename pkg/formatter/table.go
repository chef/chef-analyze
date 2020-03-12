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
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/chef/chef-analyze/pkg/reporting"
)

// MinTermWidth is required terminal width
const (
	MinTermWidth                   = 120
	filteredEmptyCookbookResultFmt = "\nNo cookbooks were found in the run lists of any filtered nodes.\n" +
		"Please verify that your node filter is correct, or use one that is less restrictive.\n" +
		"Node Filter: %s\n"
	filteredEmptyNodeResultFmt = "No nodes found with filter applied: %s"
	emptyCookbookResultMsg     = "No available cookbooks to generate a report"
	emptyNodeResultMsg         = "No nodes found to analyze."
	appliedNodesFilterFmt      = "\n\nNode Filter applied: %s\n"
)

// CookbooksReportSummary prints smaller, summarized report
func CookbooksReportSummary(state *reporting.CookbooksReport) FormattedResult {
	if state == nil {
		return FormattedResult{emptyCookbookResultMsg, ""}
	}

	if len(state.Records) == 0 {
		if len(state.NodeFilter) > 0 && state.TotalCookbooks > 0 {
			// cookbooks were found, but they were filtered out because no nodes used them.
			return FormattedResult{fmt.Sprintf(filteredEmptyCookbookResultFmt, state.NodeFilter), ""}
		} else {
			return FormattedResult{emptyCookbookResultMsg, ""}
		}
	}

	var (
		buffer                = bytes.NewBufferString("\n-- REPORT SUMMARY --\n\n")
		table                 = tablewriter.NewWriter(buffer)
		CookbooksReportHeader = []string{"Cookbook", "Version"}
		redColor              = "\033[1;31m%s\033[0m"
		yellowColor           = "\033[1;33m%s\033[0m"
	)

	if state.RunCookstyle {
		CookbooksReportHeader = append(CookbooksReportHeader, "Violations", "Auto-correctable")
	}

	CookbooksReportHeader = append(CookbooksReportHeader, "Nodes Affected")

	table.SetAutoWrapText(true)
	table.SetReflowDuringAutoWrap(true)
	table.SetHeader(CookbooksReportHeader)
	table.SetAutoFormatHeaders(false) // don't make our headers capitalized
	table.SetRowLine(false)           // don't show row seps
	table.SetColumnSeparator(" ")
	table.SetBorder(false)

	// sets max for each col to 30 chars, this is not strictly enforced.
	// unwrappable content will expand beyond this limit.
	table.SetColWidth(MinTermWidth / len(CookbooksReportHeader))

	sortCookbookRecords(state.Records)

	for _, record := range state.Records {
		row := []string{record.Name, record.Version}

		// only include violations if we ran cookstyle
		if state.RunCookstyle {
			numOffenses := record.NumOffenses()
			numCorrectable := record.NumCorrectable()
			numOffensesStr := strconv.Itoa(numOffenses)
			numCorrectableStr := strconv.Itoa(numCorrectable)
			if numOffenses != 0 {
				// Enable per-cell colorization
				if numOffenses == numCorrectable {
					numOffensesStr = fmt.Sprintf(yellowColor,numOffensesStr)
					numCorrectableStr = fmt.Sprintf(yellowColor,numCorrectableStr)
				} else {
					numOffensesStr = fmt.Sprintf(redColor,numOffensesStr)
					numCorrectableStr = fmt.Sprintf(redColor,numCorrectableStr)
				}
			}
			row = append(row,
				numOffensesStr,
				numCorrectableStr,
			)
		}

		row = append(row, strconv.Itoa(record.NumNodesAffected()))

		table.Append(row)
	}

	table.Render()

	var (
		errMsg            strings.Builder
		bufStr            = buffer.String()
		lines             = strings.SplitN(bufStr, "\n", 2)
		width             = tablewriter.DisplayWidth(lines[0])
		termWidth, _, err = terminal.GetSize(int(os.Stdout.Fd()))
	)
	if err != nil {
		termWidth = MinTermWidth
	}

	if termWidth < width {
		errMsg.WriteString("\nNote:  To view the report with correct formatting, please expand")
		errMsg.WriteString(fmt.Sprintf("\n       your terminal window to be at least %v characters wide\n", width))
	}
	if len(state.NodeFilter) > 0 {
		fmt.Fprintf(buffer, appliedNodesFilterFmt, state.NodeFilter)
	}
	return FormattedResult{buffer.String(), errMsg.String()}
}

// NodesReportSummary prints smaller, summarized report
func NodesReportSummary(records []*reporting.NodeReportItem, appliedNodesFilter string) FormattedResult {
	if len(records) == 0 {
		if len(appliedNodesFilter) > 0 {
			return FormattedResult{fmt.Sprintf(filteredEmptyNodeResultFmt, appliedNodesFilter), ""}
		} else {
			return FormattedResult{emptyNodeResultMsg, ""}
		}
	}

	var (
		buffer           = bytes.NewBufferString("\n-- REPORT SUMMARY --\n\n")
		table            = tablewriter.NewWriter(buffer)
		NodeReportHeader = []string{"Node Name", "Chef Version", "Operating System", "Number Cookbooks"}
	)
	if len(appliedNodesFilter) > 0 {
		NodeReportHeader[0] = fmt.Sprintf("Node Name (filter applied: %s)", appliedNodesFilter)
	}

	// Let's look at content to pre-determine the best column widths
	table.SetAutoWrapText(true)
	table.SetReflowDuringAutoWrap(true)
	table.SetHeader(NodeReportHeader)
	table.SetAutoFormatHeaders(false) // don't make our headers capitalized
	table.SetRowLine(false)           // don't show row seps
	table.SetColumnSeparator(" ")
	table.SetBorder(false)

	// sets max for each col to 30 chars, this is not strictly enforced
	// unwrappable content will expand beyond this limit
	table.SetColWidth(MinTermWidth / len(NodeReportHeader))

	sortNodeRecords(records)

	for _, record := range records {
		table.Append(
			[]string{
				record.Name,
				stringOrEmptyPlaceholder(record.ChefVersion),
				stringOrEmptyPlaceholder(record.OSVersionPretty()),
				strconv.Itoa(len(record.CookbooksList())),
			},
		)
	}

	table.Render()

	// A bit of a hack to find the actual width of the string used to render a line
	// of the table. We get the first line only  - this is the minimum width needed
	// to avoid wrapping  the teriminal line and making the table look bad.
	// multibyte characters accounted for by using DisplayWidth.
	var (
		errMsg            strings.Builder
		bufStr            = buffer.String()
		lines             = strings.SplitN(bufStr, "\n", 2)
		width             = tablewriter.DisplayWidth(lines[0])
		termWidth, _, err = terminal.GetSize(int(os.Stdout.Fd()))
	)
	if err != nil {
		termWidth = MinTermWidth
	}

	if termWidth < width {
		errMsg.WriteString("\nNote:  To view the report with correct formatting, please expand")
		errMsg.WriteString(fmt.Sprintf("\n       your terminal window to be at least %v characters wide\n", width))
	}
	if len(appliedNodesFilter) > 0 {
		fmt.Fprintf(buffer, appliedNodesFilterFmt, appliedNodesFilter)
	}
	return FormattedResult{buffer.String(), errMsg.String()}
}
