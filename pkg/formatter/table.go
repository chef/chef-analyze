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

const MinTermWidth = 120

func CookbooksReportSummary(state *reporting.CookbooksStatus) FormattedResult {
	if state == nil || len(state.Records) == 0 {
		return FormattedResult{"No available cookbooks to generate a report", ""}
	}

	var (
		buffer                = bytes.NewBufferString("\n-- REPORT SUMMARY --\n\n")
		table                 = tablewriter.NewWriter(buffer)
		CookbooksReportHeader = []string{"Cookbook", "Version"}
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

	for _, record := range state.Records {
		row := []string{record.Name, record.Version}

		// only include violations if we ran cookstyle
		if state.RunCookstyle {
			row = append(row,
				strconv.Itoa(record.NumOffenses()),
				strconv.Itoa(record.NumCorrectable()),
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

	return FormattedResult{buffer.String(), errMsg.String()}
}

func NodesReportSummary(records []*reporting.NodeReportItem) FormattedResult {
	if len(records) == 0 {
		return FormattedResult{"No nodes found to analyze.", ""}
	}

	var (
		buffer           = bytes.NewBufferString("\n-- REPORT SUMMARY --\n\n")
		table            = tablewriter.NewWriter(buffer)
		NodeReportHeader = []string{"Node Name", "Chef Version", "Operating System", "Cookbooks"}
	)

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

	return FormattedResult{buffer.String(), errMsg.String()}
}
