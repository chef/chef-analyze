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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/chef/chef-analyze/pkg/reporting"
)

func TestStringOrEmptyPlaceholder(t *testing.T) {
	assert.Equal(t, "-", stringOrEmptyPlaceholder(""))
	assert.Equal(t, "foo", stringOrEmptyPlaceholder("foo"))
}

func TestStringOrUnknownPlaceholder(t *testing.T) {
	assert.Equal(t, "unknown", stringOrUnknownPlaceholder(""))
	assert.Equal(t, "foo", stringOrUnknownPlaceholder("foo"))
}

func TestStringOrPlaceholder(t *testing.T) {
	assert.Equal(t, "placeholder", stringOrPlaceholder("", "placeholder"))
	assert.Equal(t, "foo", stringOrPlaceholder("foo", "placeholder"))
}

func TestSortCookbookRecords(t *testing.T) {
	records := []*reporting.CookbookRecord{
		&reporting.CookbookRecord{
			Name:    "foo",
			Version: "1.0.0",
		},
		&reporting.CookbookRecord{
			Name:    "baz",
			Version: "1.0.0",
		},
		&reporting.CookbookRecord{
			Name:    "zaz",
			Version: "1.0.0",
		},
		&reporting.CookbookRecord{
			Name:    "foo",
			Version: "2.0.0",
		},
		&reporting.CookbookRecord{
			Name:    "bar",
			Version: "1.0.0",
		},
		&reporting.CookbookRecord{
			Name:    "foo",
			Version: "1.1.0",
		},
		&reporting.CookbookRecord{
			Name:    "raz",
			Version: "1.0.0",
		},
		&reporting.CookbookRecord{
			Name:    "foo",
			Version: "1.0.1",
		},
	}
	expected := []*reporting.CookbookRecord{
		&reporting.CookbookRecord{
			Name:    "bar",
			Version: "1.0.0",
		},
		&reporting.CookbookRecord{
			Name:    "baz",
			Version: "1.0.0",
		},
		&reporting.CookbookRecord{
			Name:    "foo",
			Version: "1.0.0",
		},
		&reporting.CookbookRecord{
			Name:    "foo",
			Version: "1.0.1",
		},
		&reporting.CookbookRecord{
			Name:    "foo",
			Version: "1.1.0",
		},
		&reporting.CookbookRecord{
			Name:    "foo",
			Version: "2.0.0",
		},
		&reporting.CookbookRecord{
			Name:    "raz",
			Version: "1.0.0",
		},
		&reporting.CookbookRecord{
			Name:    "zaz",
			Version: "1.0.0",
		},
	}
	sortCookbookRecords(records)
	assert.Equal(t, expected, records)
}

func TestSortCookbookRecordsSortsNodes(t *testing.T) {
	records := []*reporting.CookbookRecord{
		&reporting.CookbookRecord{
			Name: "foo",
			Nodes: []string{
				"node2",
				"node1",
			},
		},
	}
	expected := []string{
		"node1",
		"node2",
	}
	sortCookbookRecords(records)
	assert.Equal(t, expected, records[0].Nodes)
}

func TestSortNodeRecords(t *testing.T) {
	records := []*reporting.NodeReportItem{
		&reporting.NodeReportItem{
			Name: "foo",
		},
		&reporting.NodeReportItem{
			Name: "baz",
		},
		&reporting.NodeReportItem{
			Name: "123",
		},
		&reporting.NodeReportItem{
			Name: "zzz",
		},
		&reporting.NodeReportItem{
			Name: "ZZZ",
		},
		&reporting.NodeReportItem{
			Name: "BAZ",
		},
		&reporting.NodeReportItem{
			Name: "FOO",
		},
	}
	expected := []*reporting.NodeReportItem{
		&reporting.NodeReportItem{
			Name: "123",
		},
		&reporting.NodeReportItem{
			Name: "baz",
		},
		&reporting.NodeReportItem{
			Name: "BAZ",
		},
		&reporting.NodeReportItem{
			Name: "foo",
		},
		&reporting.NodeReportItem{
			Name: "FOO",
		},
		&reporting.NodeReportItem{
			Name: "zzz",
		},
		&reporting.NodeReportItem{
			Name: "ZZZ",
		},
	}
	sortNodeRecords(records)
	assert.Equal(t, expected, records)
}

func TestSortNodeRecordsSortCookbooks(t *testing.T) {
	records := []*reporting.NodeReportItem{
		&reporting.NodeReportItem{
			Name: "foo",
			CookbookVersions: []reporting.CookbookVersion{
				reporting.CookbookVersion{Name: "zzz", Version: "1.0"},
				reporting.CookbookVersion{Name: "zzz", Version: "10.0"},
				reporting.CookbookVersion{Name: "zzz", Version: "0.0"},
				reporting.CookbookVersion{Name: "xxx", Version: "9.9"},
				reporting.CookbookVersion{Name: "xxx", Version: "9.0"},
			},
		},
		&reporting.NodeReportItem{
			Name: "baz",
			CookbookVersions: []reporting.CookbookVersion{
				reporting.CookbookVersion{Name: "zzz", Version: "1.0"},
				reporting.CookbookVersion{Name: "zzz", Version: "10.0"},
				reporting.CookbookVersion{Name: "zzz", Version: "0.0"},
				reporting.CookbookVersion{Name: "xxx", Version: "9.9"},
				reporting.CookbookVersion{Name: "xxx", Version: "9.0"},
			},
		},
	}
	expected := []*reporting.NodeReportItem{
		&reporting.NodeReportItem{
			Name: "baz",
			CookbookVersions: []reporting.CookbookVersion{
				reporting.CookbookVersion{Name: "xxx", Version: "9.0"},
				reporting.CookbookVersion{Name: "xxx", Version: "9.9"},
				reporting.CookbookVersion{Name: "zzz", Version: "0.0"},
				reporting.CookbookVersion{Name: "zzz", Version: "1.0"},
				reporting.CookbookVersion{Name: "zzz", Version: "10.0"},
			},
		},
		&reporting.NodeReportItem{
			Name: "foo",
			CookbookVersions: []reporting.CookbookVersion{
				reporting.CookbookVersion{Name: "xxx", Version: "9.0"},
				reporting.CookbookVersion{Name: "xxx", Version: "9.9"},
				reporting.CookbookVersion{Name: "zzz", Version: "0.0"},
				reporting.CookbookVersion{Name: "zzz", Version: "1.0"},
				reporting.CookbookVersion{Name: "zzz", Version: "10.0"},
			},
		},
	}
	sortNodeRecords(records)
	assert.Equal(t, expected, records)
}
