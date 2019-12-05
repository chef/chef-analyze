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
	"os"

	"github.com/olekukonko/tablewriter"

	"github.com/chef/chef-analyze/pkg/reporting"
)

func WriteNodeReport(records []reporting.NodeReportItem) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Node Name", "Chef Version", "OS", "OS Version", "Cookbooks"})
	table.SetReflowDuringAutoWrap(true)
	table.SetRowLine(true)
	table.SetAutoWrapText(true)
	table.SetReflowDuringAutoWrap(true)
	table.SetBorder(true)
	for _, record := range records {
		table.Append(record.Array())
	}
	table.Render()
}
