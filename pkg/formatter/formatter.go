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
	"sort"

	"github.com/chef/chef-analyze/pkg/reporting"
)

// FormattedResult contains the report and its errors
type FormattedResult struct {
	Report string
	Errors string
}

const (
	emptyValuePlaceholder   = "-"
	unknownValuePlaceholder = "unknown"
)

func stringOrEmptyPlaceholder(s string) string {
	return stringOrPlaceholder(s, emptyValuePlaceholder)
}

func stringOrUnknownPlaceholder(s string) string {
	return stringOrPlaceholder(s, unknownValuePlaceholder)
}

func stringOrPlaceholder(s, placeholder string) string {
	if len(s) == 0 {
		return placeholder
	}
	return s
}

func sortCookbookRecords(records []*reporting.CookbookRecord) {
	sort.Sort(reporting.CookbookRecordsByNameVersion(records))

	// Also sort the node names.
	for _, record := range records {
		sort.Strings(record.Nodes)
	}
}
