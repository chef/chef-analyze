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
	"strings"

	"github.com/olekukonko/tablewriter"

	"github.com/chef/chef-analyze/pkg/reporting"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	MinTermWidth          = 120
	EmptyValuePlaceholder = "-"
)

func WriteNodeReport(records []*reporting.NodeReportItem) {
	termWidth, _, err := terminal.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		termWidth = MinTermWidth
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoWrapText(true)
	table.SetReflowDuringAutoWrap(true)
	table.SetColMinWidth(0, int(float64(termWidth)*0.30)) // fqdns can get pretty long
	table.SetColMinWidth(1, int(float64(termWidth)*0.10)) // chef version is tiny
	table.SetColMinWidth(2, int(float64(termWidth)*0.15)) // OS+version string
	// Help prevent any one column from expanding to fill all available space:
	table.SetColWidth(int(float64(termWidth) * 0.35))
	table.SetHeader([]string{"Node Name", "Chef Version", "Operating System", "Cookbooks"})
	table.SetAutoFormatHeaders(false) // Don't make our headers capitalize
	table.SetRowLine(false)           // don't show row seps
	table.SetColumnSeparator(" ")
	table.SetBorder(false)
	for _, record := range records {
		table.Append(nodeReportItemToArray(record))
	}

	fmt.Print("\n")
	table.Render()
	if termWidth < MinTermWidth {
		fmt.Print("\nNote:  If the report is not formatted correctly, please")
		fmt.Print("\n       please expand your terminal window to be at least")
		fmt.Printf("\n       %v characters wide.\n", MinTermWidth)
	}
}

func nodeReportItemToArray(nri *reporting.NodeReportItem) []string {
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
