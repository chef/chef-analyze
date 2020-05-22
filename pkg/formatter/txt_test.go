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
	assert.Equal(t, 4, len(lines))
	assert.Contains(t, actual.Report, "Nodes affected: node-1, node-2")
	assert.Contains(t, actual.Report, "> Cookbook: my-cookbook (1.0)")
	assert.Contains(t, actual.Report, "Policy Group: none")
	assert.Equal(t, "", lines[3])
}

func TestMakeCookbooksReportTXT_WithUnverifiedPolicyRecords(t *testing.T) {
	cbStatus := reporting.CookbooksReport{
		RunCookstyle: false,
		Records: []*reporting.CookbookRecord{
			&reporting.CookbookRecord{Name: "my-cookbook", PolicyGroup: "my-policygroup", Policy: "my-policy", PolicyVer: "123xyzx", Nodes: []string{"node-1", "node-2"}},
		},
	}

	actual := subject.MakeCookbooksReportTXT(&cbStatus)
	lines := strings.Split(actual.Report, "\n")
	assert.Equal(t, 4, len(lines))
	assert.Contains(t, actual.Report, "Nodes affected: node-1, node-2")
	assert.Contains(t, actual.Report, "> Cookbook: my-cookbook (policy my-policy, revision 123xyzx)")
	assert.Contains(t, actual.Report, "Policy Group: my-policygroup")
	assert.Equal(t, "", lines[3])
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
	assert.Contains(t, actual.Report, "Policy Group: none")
	assert.Contains(t, actual.Report, "\tChefDeprecations/Blah (true) some description")

	lines := strings.Split(actual.Report, "\n")
	assert.Equal(t, 9, len(lines))
	assert.Equal(t, "", lines[8])
}

func TestMakeCookbooksReportTXT_WithVerifiedPolicyRecords(t *testing.T) {
	cbStatus := reporting.CookbooksReport{
		RunCookstyle: true,
		Records: []*reporting.CookbookRecord{
			&reporting.CookbookRecord{Name: "my-cookbook", PolicyGroup: "my-policygroup", Policy: "my-policy", PolicyVer: "123xyzx", Nodes: []string{"node-1", "node-2"},
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
	assert.Contains(t, actual.Report, "Policy Group: my-policygroup")
	assert.Contains(t, actual.Report, "my-cookbook (policy my-policy, revision 123xyzx)")
	assert.Contains(t, actual.Report, "\tChefDeprecations/Blah (true) some description")

	lines := strings.Split(actual.Report, "\n")
	assert.Equal(t, 9, len(lines))
	assert.Equal(t, "", lines[8])
}

func TestMakeCookbooksReportTXT_ErrorReport(t *testing.T) {
	cbStatus := reporting.CookbooksReport{
		Records: []*reporting.CookbookRecord{
			&reporting.CookbookRecord{Name: "my-cookbook", Version: "1.0", DownloadError: errors.New("could not download")},
			&reporting.CookbookRecord{Name: "their-cookbook", Version: "1.1", UsageLookupError: errors.New("could not look up usage")},
			&reporting.CookbookRecord{Name: "our-cookbook", Version: "1.2", CookstyleError: errors.New("cookstyle error")},
			&reporting.CookbookRecord{Name: "her-cookbook", PolicyGroup: "my-policygroup", Policy: "my-policy", PolicyVer: "123xyzx", CookstyleError: errors.New("not found 404")},
		},
	}

	actual := subject.MakeCookbooksReportTXT(&cbStatus)
	lines := strings.Split(actual.Errors, "\n")
	assert.Equal(t, 5, len(lines))
	assert.Equal(t, lines[0], " - my-cookbook (1.0): could not download")
	assert.Equal(t, lines[1], " - our-cookbook (1.2): cookstyle error")
	assert.Equal(t, lines[2], " - their-cookbook (1.1): could not look up usage")
	assert.Equal(t, lines[3], " - her-cookbook (PolicyGroup my-policygroup, Policy my-policy, PolicyRevision 123xyzx): not found 404")
	assert.Equal(t, lines[4], "")
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
		&reporting.NodeReportItem{Name: "node2", ChefVersion: "13.11", OS: "", OSVersion: "", PolicyGroup: "prod", Policy: "grafana", PolicyRev: "xyz1234567890",
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
  Policy Group: no group
  Policy: no policy
  Cookbooks Applied (alphanumeric order): mycookbook(1.0)
> Node: node2
  Chef Version: 13.11
  Operating System: unknown
  Policy Group: prod
  Policy: grafana(rev xyz1234567890)
  Cookbooks Applied (alphanumeric order): mycookbook, test
> Node: node3
  Chef Version: 15.00
  Operating System: ubuntu v16.04
  Policy Group: no group
  Policy: no policy
  Cookbooks Applied: none
`
	)
	if assert.Equal(t, 19, len(lines)) {
		assert.Equal(t, expectedReport, actual.Report)
	}
}

func TestMakeNodesReportTXT_WithRecordsSorted(t *testing.T) {
	nodesReport := []*reporting.NodeReportItem{
		&reporting.NodeReportItem{Name: "zzz", ChefVersion: "13.11", OS: "", OSVersion: "",
			CookbookVersions: []reporting.CookbookVersion{
				reporting.CookbookVersion{Name: "yyy", Version: "1.0"},
				reporting.CookbookVersion{Name: "ccc", Version: "9.9"},
			},
		},
		&reporting.NodeReportItem{Name: "Aaa", ChefVersion: "13.11", OS: "", OSVersion: "",
			CookbookVersions: []reporting.CookbookVersion{
				reporting.CookbookVersion{Name: "yyy", Version: "1.0"},
				reporting.CookbookVersion{Name: "YYY", Version: "9.9"},
				reporting.CookbookVersion{Name: "yyy", Version: "10.0"},
				reporting.CookbookVersion{Name: "YYY", Version: "99.9"},
			},
		},
		&reporting.NodeReportItem{Name: "123", ChefVersion: "13.11", OS: "", OSVersion: "",
			CookbookVersions: []reporting.CookbookVersion{},
		},
		&reporting.NodeReportItem{Name: "aaa", ChefVersion: "13.11", OS: "", OSVersion: "", PolicyGroup: "prod", Policy: "grafana", PolicyRev: "xyz1234567890",
			CookbookVersions: []reporting.CookbookVersion{
				reporting.CookbookVersion{Name: "ddd", Version: "1.0"},
				reporting.CookbookVersion{Name: "ccc", Version: "9.9"},
			},
		},
	}

	var (
		actual         = subject.MakeNodesReportTXT(nodesReport, "")
		lines          = strings.Split(actual.Report, "\n")
		expectedReport = `> Node: 123
  Chef Version: 13.11
  Operating System: unknown
  Policy Group: no group
  Policy: no policy
  Cookbooks Applied: none
> Node: Aaa
  Chef Version: 13.11
  Operating System: unknown
  Policy Group: no group
  Policy: no policy
  Cookbooks Applied (alphanumeric order): yyy(1.0), yyy(10.0), YYY(9.9), YYY(99.9)
> Node: aaa
  Chef Version: 13.11
  Operating System: unknown
  Policy Group: prod
  Policy: grafana(rev xyz1234567890)
  Cookbooks Applied (alphanumeric order): ccc, ddd
> Node: zzz
  Chef Version: 13.11
  Operating System: unknown
  Policy Group: no group
  Policy: no policy
  Cookbooks Applied (alphanumeric order): ccc(9.9), yyy(1.0)
`
	)
	if assert.Equal(t, 25, len(lines)) {
		assert.Equal(t, expectedReport, actual.Report)
	}
}

func TestMakeNodesReportTXT_WithRecordsAndFilter(t *testing.T) {
	nodesReport := []*reporting.NodeReportItem{
		&reporting.NodeReportItem{Name: "node1", ChefVersion: "12.22", OS: "windows", OSVersion: "10.1",
			CookbookVersions: []reporting.CookbookVersion{
				reporting.CookbookVersion{Name: "mycookbook", Version: "1.0"}},
		},
		&reporting.NodeReportItem{Name: "node2", ChefVersion: "13.11", OS: "", OSVersion: "", PolicyGroup: "staging", Policy: "seven-zip", PolicyRev: "99999xxxx99999",
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
  Policy Group: no group
  Policy: no policy
  Cookbooks Applied (alphanumeric order): mycookbook(1.0)
> Node: node2
  Chef Version: 13.11
  Operating System: unknown
  Policy Group: staging
  Policy: seven-zip(rev 99999xxxx99999)
  Cookbooks Applied (alphanumeric order): mycookbook, test
> Node: node3
  Chef Version: 15.00
  Operating System: ubuntu v16.04
  Policy Group: no group
  Policy: no policy
  Cookbooks Applied: none
`
	)
	if assert.Equal(t, 20, len(lines)) {
		assert.Equal(t, expectedReport, actual.Report)
	}
}
