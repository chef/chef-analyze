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

package formatter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	subject "github.com/chef/chef-analyze/pkg/formatter"
	"github.com/chef/chef-analyze/pkg/reporting"
)

func TestMakeCookbooksReportTXT_Nil(t *testing.T) {
	assert.Equal(t,
		&subject.FormattedResult{Report: "", Errors: ""},
		subject.MakeCookbooksReportTXT(nil))
}

func TestMakeCookbooksReportTXT_Empty(t *testing.T) {
	assert.Equal(t,
		&subject.FormattedResult{Report: "", Errors: ""},
		subject.MakeCookbooksReportTXT(&reporting.CookbooksReport{}))
}

func TestMakeCookbooksReportTXT_EmptyFromFilter(t *testing.T) {
	assert.Equal(t,
		&subject.FormattedResult{Report: "Node filter applied: blah\n", Errors: ""},
		subject.MakeCookbooksReportTXT(&reporting.CookbooksReport{NodeFilter: "blah"}))
}
func TestMakeCookbooksReportTXT_WithUnverifiedRecords(t *testing.T) {
	cbStatus := reporting.CookbooksReport{
		RunCookstyle: false,
		Records: []*reporting.CookbookRecord{
			&reporting.CookbookRecord{Name: "my-cookbook", Version: "1.0", Nodes: []string{"node-1", "node-2"}},
		},
	}

	actual := subject.MakeCookbooksReportTXT(&cbStatus)
	lines := strings.Split(actual.Report, "\n")
	assert.Equal(t, 3, len(lines))
	assert.Contains(t, actual.Report, "Nodes affected: node-1, node-2")
	assert.Contains(t, actual.Report, "> Cookbook: my-cookbook (1.0)")
	assert.Equal(t, "", lines[2])
}

func TestMakeCookbooksReportTXT_WithVerifiedRecords(t *testing.T) {
	cbStatus := reporting.CookbooksReport{
		RunCookstyle: true,
		Records: []*reporting.CookbookRecord{
			&reporting.CookbookRecord{Name: "my-cookbook", Version: "1.0", Nodes: []string{"node-1", "node-2"},
				Files: []reporting.CookbookFile{
					reporting.CookbookFile{Path: "/path/to/file.rb",
						Offenses: []reporting.CookstyleOffense{
							reporting.CookstyleOffense{CopName: "ChefDeprecations/Blah", Message: "some description", Correctable: true},
						}}}}}}

	actual := subject.MakeCookbooksReportTXT(&cbStatus)
	assert.Contains(t, actual.Report, "Nodes affected: node-1, node-2")
	assert.Contains(t, actual.Report, "Violations: 1")
	assert.Contains(t, actual.Report, "Auto correctable: 1")
	assert.Contains(t, actual.Report, "Files and offenses:")
	assert.Contains(t, actual.Report, "path/to/file.rb:")
	assert.Contains(t, actual.Report, "\tChefDeprecations/Blah (true) some description")

	lines := strings.Split(actual.Report, "\n")
	assert.Equal(t, 8, len(lines))
	assert.Equal(t, "", lines[7])
}

func TestMakeCookbooksReportTXT_ErrorReport(t *testing.T) {
	cbStatus := reporting.CookbooksReport{
		Records: []*reporting.CookbookRecord{
			&reporting.CookbookRecord{Name: "my-cookbook", Version: "1.0", DownloadError: errors.New("could not download")},
			&reporting.CookbookRecord{Name: "their-cookbook", Version: "1.1", UsageLookupError: errors.New("could not look up usage")},
			&reporting.CookbookRecord{Name: "our-cookbook", Version: "1.2", CookstyleError: errors.New("cookstyle error")},
		},
	}

	actual := subject.MakeCookbooksReportTXT(&cbStatus)
	lines := strings.Split(actual.Errors, "\n")
	assert.Equal(t, 4, len(lines))
	assert.Equal(t, lines[0], " - my-cookbook (1.0): could not download")
	assert.Equal(t, lines[1], " - our-cookbook (1.2): cookstyle error")
	assert.Equal(t, lines[2], " - their-cookbook (1.1): could not look up usage")
	assert.Equal(t, lines[3], "")
}

func TestMakeNodesReportTXT_Nil(t *testing.T) {
	assert.Equal(t,
		&subject.FormattedResult{Report: "", Errors: ""},
		subject.MakeNodesReportTXT(nil, ""))
}

func TestMakeNodesReportTXT_Empty(t *testing.T) {
	assert.Equal(t,
		&subject.FormattedResult{Report: "", Errors: ""},
		subject.MakeNodesReportTXT([]*reporting.NodeReportItem{}, ""))
}
func TestMakeNodesReportTXT_EmptyFromFilter(t *testing.T) {
	assert.Equal(t,
		&subject.FormattedResult{Report: "Node Filter Applied: name:blah\n", Errors: ""},
		subject.MakeNodesReportTXT([]*reporting.NodeReportItem{}, "name:blah"))
}

func TestMakeNodesReportTXT_WithRecords(t *testing.T) {
	nodesReport := []*reporting.NodeReportItem{
		&reporting.NodeReportItem{Name: "node1", ChefVersion: "12.22", OS: "windows", OSVersion: "10.1",
			CookbookVersions: []reporting.CookbookVersion{
				reporting.CookbookVersion{Name: "mycookbook", Version: "1.0"}},
		},
		&reporting.NodeReportItem{Name: "node2", ChefVersion: "13.11", OS: "", OSVersion: "",
			CookbookVersions: []reporting.CookbookVersion{
				reporting.CookbookVersion{Name: "mycookbook", Version: "1.0"},
				reporting.CookbookVersion{Name: "test", Version: "9.9"},
			},
		},
		&reporting.NodeReportItem{Name: "node3", ChefVersion: "15.00", OS: "ubuntu", OSVersion: "16.04",
			CookbookVersions: nil},
	}

	var (
		actual         = subject.MakeNodesReportTXT(nodesReport, "")
		lines          = strings.Split(actual.Report, "\n")
		expectedReport = `> Node: node1
  Chef Version: 12.22
  Operating System: windows v10.1
  Cookbooks Applied: mycookbook(1.0)
> Node: node2
  Chef Version: 13.11
  Operating System: unknown
  Cookbooks Applied: mycookbook(1.0), test(9.9)
> Node: node3
  Chef Version: 15.00
  Operating System: ubuntu v16.04
  Cookbooks Applied: none
`
	)
	if assert.Equal(t, 13, len(lines)) {
		assert.Equal(t, expectedReport, actual.Report)
	}
}

func TestMakeNodesReportTXT_WithRecordsAndFilter(t *testing.T) {
	nodesReport := []*reporting.NodeReportItem{
		&reporting.NodeReportItem{Name: "node1", ChefVersion: "12.22", OS: "windows", OSVersion: "10.1",
			CookbookVersions: []reporting.CookbookVersion{
				reporting.CookbookVersion{Name: "mycookbook", Version: "1.0"}},
		},
		&reporting.NodeReportItem{Name: "node2", ChefVersion: "13.11", OS: "", OSVersion: "",
			CookbookVersions: []reporting.CookbookVersion{
				reporting.CookbookVersion{Name: "mycookbook", Version: "1.0"},
				reporting.CookbookVersion{Name: "test", Version: "9.9"},
			},
		},
		&reporting.NodeReportItem{Name: "node3", ChefVersion: "15.00", OS: "ubuntu", OSVersion: "16.04",
			CookbookVersions: nil},
	}

	var (
		actual         = subject.MakeNodesReportTXT(nodesReport, "name:blah")
		lines          = strings.Split(actual.Report, "\n")
		expectedReport = `Node Filter Applied: name:blah
> Node: node1
  Chef Version: 12.22
  Operating System: windows v10.1
  Cookbooks Applied: mycookbook(1.0)
> Node: node2
  Chef Version: 13.11
  Operating System: unknown
  Cookbooks Applied: mycookbook(1.0), test(9.9)
> Node: node3
  Chef Version: 15.00
  Operating System: ubuntu v16.04
  Cookbooks Applied: none
`
	)
	if assert.Equal(t, 14, len(lines)) {
		assert.Equal(t, expectedReport, actual.Report)
	}
}
