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
	"strings"

	"github.com/olekukonko/tablewriter"

	"github.com/chef/chef-analyze/pkg/reporting"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	MinTermWidth          = 120
	EmptyValuePlaceholder = "-"
)

var (
	NodeReportHeader = []string{"Node Name", "Chef Version", "Operating System", "Cookbooks"}
)

func WriteNodeReport(records []*reporting.NodeReportItem) {
	termWidth, _, err := terminal.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		termWidth = MinTermWidth
	}
	buffer := bytes.NewBufferString("")

	// Let's look at content to pre-determine the best column widths
	table := tablewriter.NewWriter(buffer)
	table.SetAutoWrapText(true)
	table.SetReflowDuringAutoWrap(true)
	table.SetHeader(NodeReportHeader)
	table.SetAutoFormatHeaders(false) // Don't make our headers capitalized
	table.SetRowLine(false)           // don't show row seps
	table.SetColumnSeparator(" ")
	table.SetBorder(false)

	// sets max for each col to 30 chars. This is not strictly enforced - unwrappable content will
	// expand beyond this limit.
	table.SetColWidth(MinTermWidth / len(NodeReportHeader))

	for _, record := range records {
		table.Append(NodeReportItemToArray(record))
	}

	fmt.Print("\n")
	table.Render()

	// A bit of a hack to find the actual width of the string used to render a line
	// of the table. We get the first line only  - this is the minimum width needed
	// to avoid wrapping  the teriminal line and making the table look bad.
	// multibyte characters accounted for by using DisplayWidth.
	bufStr := buffer.String()
	lines := strings.SplitN(bufStr, "\n", 2)
	width := tablewriter.DisplayWidth(lines[0])

	fmt.Printf(bufStr)
	if termWidth < width {
		fmt.Print("\nNote:  To view the report with correct formatting, please expand")
		fmt.Printf("\n       your terminal window to be at least %v characters wide\n", width)
	}
}

func NodeReportItemToArray(nri *reporting.NodeReportItem) []string {
	var cookbooks []string
	for _, v := range nri.CookbookVersions {
		cookbooks = append(cookbooks, v.String())
	}
	var chefVersion string
	if nri.ChefVersion == "" {
		chefVersion = EmptyValuePlaceholder
	} else {
		chefVersion = nri.ChefVersion
	}
	// This data seems to be all or none - you'll have both OS/Version fields,
	// or neither.
	var osInfo string
	if nri.OS == "" {
		osInfo = EmptyValuePlaceholder
	} else {
		osInfo = fmt.Sprintf("%s v%s", nri.OS, nri.OSVersion)
	}

	var cbInfo string
	if len(cookbooks) == 0 {
		cbInfo = EmptyValuePlaceholder
	} else {
		cbInfo = strings.Join(cookbooks, " ")
	}

	return []string{nri.Name, chefVersion, osInfo, cbInfo}
}
