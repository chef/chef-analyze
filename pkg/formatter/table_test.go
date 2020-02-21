//
// Copyright 2019 Chef Software, Inc.
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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	subject "github.com/chef/chef-analyze/pkg/formatter"
	"github.com/chef/chef-analyze/pkg/reporting"
)

func TestNodesReportSummary_Nil(t *testing.T) {
	expected := subject.FormattedResult{"No nodes found to analyze.", ""}
	assert.Equal(t, expected, subject.NodesReportSummary(nil, ""))
}

func TestNodesReportSummary_noRecords(t *testing.T) {
	nri := []*reporting.NodeReportItem{}
	expected := subject.FormattedResult{"No nodes found to analyze.", ""}
	assert.Equal(t, expected, subject.NodesReportSummary(nri, ""))
}
func TestNodesReportSummary_noRecordsAndFilter(t *testing.T) {
	nri := []*reporting.NodeReportItem{}
	expected := subject.FormattedResult{"No nodes found with filter applied: name:blah", ""}
	assert.Equal(t, expected, subject.NodesReportSummary(nri, "name:blah"))
}

func TestNodesReportSummary_withRecords(t *testing.T) {
	nri := []*reporting.NodeReportItem{
		&reporting.NodeReportItem{
			Name:        "abc-1",
			ChefVersion: "15.4",
			OS:          "os",
			OSVersion:   "1.0",
			CookbookVersions: []reporting.CookbookVersion{
				reporting.CookbookVersion{Name: "cookbook1", Version: "0.3.0"},
			},
		},
		&reporting.NodeReportItem{
			Name:        "abc-2",
			ChefVersion: "13.1.20",
			OS:          "os",
			OSVersion:   "1.0",
			CookbookVersions: []reporting.CookbookVersion{
				reporting.CookbookVersion{Name: "cool", Version: "0.1.0"},
				reporting.CookbookVersion{Name: "awesome", Version: "1.2.3"},
				reporting.CookbookVersion{Name: "cookbook1", Version: "0.3.0"},
			},
		},
	}
	report := subject.NodesReportSummary(nri, "")

	assert.Contains(t,
		report.Report,
		"REPORT SUMMARY",
		"stdout missing report summary header")
	assert.Contains(t,
		report.Report,
		"Node Name",
		"stdout missing Node Name header")
	assert.Contains(t,
		report.Report,
		"Chef Version", "stdout missing Chef Version header")
	assert.Contains(t,
		report.Report,
		"Operating System", "stdout missing Operating System header")
	assert.Contains(t,
		report.Report,
		"Cookbooks", "stdout missing Operating System header")

	listOfStringTheReportMustHave := []string{
		"abc-1",
		"abc-2",
		"15.4",
		"os v1.0",
		"13.1.20",
	}
	for _, s := range listOfStringTheReportMustHave {
		assert.Containsf(t, report.Report, s,
			"there is something missing in the stdout: '%s' is missing", s,
		)
	}
}

func TestNodesReportSummary_withRecordsSorted(t *testing.T) {
	nri := []*reporting.NodeReportItem{
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
		&reporting.NodeReportItem{Name: "aaa", ChefVersion: "13.11", OS: "", OSVersion: "",
			CookbookVersions: []reporting.CookbookVersion{
				reporting.CookbookVersion{Name: "ddd", Version: "1.0"},
				reporting.CookbookVersion{Name: "ccc", Version: "9.9"},
			},
		},
	}
	report := subject.NodesReportSummary(nri, "")

	lines := strings.Split(report.Report, "\n")
	if assert.Equal(t, 10, len(lines)) {
		assert.Equal(t, "-- REPORT SUMMARY --", lines[1])
		assert.Equal(t, "  Node Name   Chef Version   Operating System   Number Cookbooks  ", lines[3])
		assert.Equal(t, "        123          13.11   -                                 0  ", lines[5])
		assert.Equal(t, "  Aaa                13.11   -                                 4  ", lines[6])
		assert.Equal(t, "  aaa                13.11   -                                 2  ", lines[7])
		assert.Equal(t, "  zzz                13.11   -                                 2  ", lines[8])
		assert.Equal(t, "", lines[9])
	}
}

func TestNodesReportSummary_withRecordsAndFilter(t *testing.T) {
	nri := []*reporting.NodeReportItem{
		&reporting.NodeReportItem{
			Name:        "abc-1",
			ChefVersion: "15.4",
			OS:          "os",
			OSVersion:   "1.0",
			CookbookVersions: []reporting.CookbookVersion{
				reporting.CookbookVersion{Name: "cookbook1", Version: "0.3.0"},
			},
		},
		&reporting.NodeReportItem{
			Name:        "abc-2",
			ChefVersion: "13.1.20",
			OS:          "os",
			OSVersion:   "1.0",
			CookbookVersions: []reporting.CookbookVersion{
				reporting.CookbookVersion{Name: "cool", Version: "0.1.0"},
				reporting.CookbookVersion{Name: "awesome", Version: "1.2.3"},
				reporting.CookbookVersion{Name: "cookbook1", Version: "0.3.0"},
			},
		},
	}
	report := subject.NodesReportSummary(nri, "name:blah")

	assert.Contains(t,
		report.Report,
		"REPORT SUMMARY",
		"stdout missing report summary header")
	assert.Contains(t,
		report.Report,
		"Node Name",
		"stdout missing Node Name header")
	assert.Contains(t,
		report.Report,
		"Chef Version", "stdout missing Chef Version header")
	assert.Contains(t,
		report.Report,
		"Operating System", "stdout missing Operating System header")
	assert.Contains(t,
		report.Report,
		"Node Filter applied: name:blah", "stdout missing filter description",
		report.Report)

	listOfStringTheReportMustHave := []string{
		"abc-1",
		"abc-2",
		"15.4",
		"os v1.0",
		"13.1.20",
	}
	for _, s := range listOfStringTheReportMustHave {
		assert.Containsf(t, report.Report, s,
			"there is something missing in the stdout: '%s' is missing", s,
		)
	}
}
func TestNodesReportSummary_withRecords_NoCookbooks(t *testing.T) {
	nri := []*reporting.NodeReportItem{
		&reporting.NodeReportItem{
			Name:        "node-1",
			ChefVersion: "15.4",
			OS:          "darwin",
			OSVersion:   "1.0",
		},
		&reporting.NodeReportItem{
			Name:        "node-2",
			ChefVersion: "13.1.20",
			OS:          "ubuntu",
			OSVersion:   "1.0",
		},
	}
	report := subject.NodesReportSummary(nri, "")

	assert.Contains(t,
		report.Report,
		"REPORT SUMMARY",
		"stdout missing report summary header")
	assert.Contains(t,
		report.Report,
		"Node Name",
		"stdout missing Node Name header")
	assert.Contains(t,
		report.Report,
		"Chef Version", "stdout missing Chef Version header")
	assert.Contains(t,
		report.Report,
		"Operating System", "stdout missing Operating System header")
	assert.Contains(t,
		report.Report,
		"Cookbooks", "stdout missing Operating System header")

	listOfStringTheReportMustHave := []string{
		"node-1",
		"node-2",
		"15.4",
		"darwin v1.0",
		"ubuntu v1.0",
		"13.1.20",
	}
	for _, s := range listOfStringTheReportMustHave {
		assert.Containsf(t, report.Report, s,
			"there is something missing in the stdout: '%s' is missing", s,
		)
	}
}

func TestCookbooksReportSummary_Nil(t *testing.T) {
	assert.Equal(t,
		subject.FormattedResult{
			"No available cookbooks to generate a report",
			"",
		},
		subject.CookbooksReportSummary(nil),
	)
}

func TestCookbooksReportSummary_noRecords(t *testing.T) {
	state := &reporting.CookbooksReport{}
	assert.Equal(t,
		subject.FormattedResult{
			"No available cookbooks to generate a report",
			"",
		},
		subject.CookbooksReportSummary(state),
	)
}

func TestCookbooksReportSummary_noRecordsWithFilter(t *testing.T) {
	state := &reporting.CookbooksReport{NodeFilter: "blah"}
	assert.Equal(t,
		subject.FormattedResult{
			"No available cookbooks to generate a report",
			"",
		},
		subject.CookbooksReportSummary(state),
	)
}
func TestCookbooksReportSummary_noRecordsBecauseFiltered(t *testing.T) {
	state := &reporting.CookbooksReport{NodeFilter: "blah", TotalCookbooks: 1}
	summary := subject.CookbooksReportSummary(state)
	assert.Contains(t, summary.Report, "run lists of any filtered nodes")
	assert.Equal(t, summary.Errors, "")
}

func TestCookbooksReportSummary_withRecords(t *testing.T) {
	state := &reporting.CookbooksReport{
		Records: []*reporting.CookbookRecord{
			&reporting.CookbookRecord{
				Name:    "foo",
				Version: "0.1.0",
				Nodes:   []string{"node1", "node2"},
				Files: []reporting.CookbookFile{
					reporting.CookbookFile{
						Path: "metadata.rb",
						Offenses: []reporting.CookstyleOffense{
							reporting.CookstyleOffense{Correctable: false},
						},
					},
					reporting.CookbookFile{
						Path: "recipes/default.rb",
						Offenses: []reporting.CookstyleOffense{
							reporting.CookstyleOffense{Correctable: true},
							reporting.CookstyleOffense{Correctable: false},
							reporting.CookstyleOffense{Correctable: true},
							reporting.CookstyleOffense{Correctable: true},
						},
					},
				},
			},
		},
		RunCookstyle: true,
	}
	report := subject.CookbooksReportSummary(state)

	assert.Contains(t,
		report.Report,
		"REPORT SUMMARY",
		"stdout missing report summary header")
	assert.Contains(t,
		report.Report,
		"Cookbook", "stdout missing header")
	assert.Contains(t,
		report.Report,
		"Version", "stdout missing header")
	assert.Contains(t,
		report.Report,
		"Violations", "stdout missing header")
	assert.Contains(t,
		report.Report,
		"Auto-correctable", "stdout missing header")
	assert.Contains(t,
		report.Report,
		"Nodes Affected", "stdout missing header")

	listOfStringTheReportMustHave := []string{
		"foo",
		"0.1.0",
		"5", // Violations
		"3", // auto-correctable
		"2", // nodes affected
	}
	for _, s := range listOfStringTheReportMustHave {
		assert.Containsf(t, report.Report, s,
			"there is something missing in the stdout: '%s' is missing", s,
		)
	}
}

func TestCookbooksReportSummary_withFilterFooter(t *testing.T) {
	state := &reporting.CookbooksReport{
		Records: []*reporting.CookbookRecord{
			&reporting.CookbookRecord{
				Name:    "foo",
				Version: "0.1.0",
				Nodes:   []string{"node1", "node2"},
				Files: []reporting.CookbookFile{
					reporting.CookbookFile{
						Path:     "metadata.rb",
						Offenses: []reporting.CookstyleOffense{},
					},
				},
			},
		},
		RunCookstyle: true,
		NodeFilter:   "blah",
	}
	report := subject.CookbooksReportSummary(state)

	assert.Contains(t,
		report.Report,
		"REPORT SUMMARY",
		"stdout missing report summary header")
	assert.Contains(t,
		report.Report,
		"Node Filter applied: blah",
		"stdout missing nodes filter footer")
}

func TestCookbooksReportSummary_withRecords_NoNodes(t *testing.T) {
	state := &reporting.CookbooksReport{
		Records: []*reporting.CookbookRecord{
			&reporting.CookbookRecord{
				Name: "foo", Version: "0.1.0", Nodes: nil,
			},
			&reporting.CookbookRecord{
				Name: "bar", Version: "0.2.0", Nodes: nil,
			},
			&reporting.CookbookRecord{
				Name: "xyz", Version: "0.3.0", Nodes: nil,
			},
		},
		RunCookstyle: false,
	}
	report := subject.CookbooksReportSummary(state)

	assert.Contains(t,
		report.Report,
		"REPORT SUMMARY",
		"stdout missing report summary header")
	assert.Contains(t,
		report.Report,
		"Cookbook", "stdout missing header")
	assert.Contains(t,
		report.Report,
		"Version", "stdout missing header")
	assert.Contains(t,
		report.Report,
		"Nodes Affected", "stdout missing header")

	// Should NOT Contains the following
	assert.NotContains(t,
		report.Report,
		"Violations", "stdout shoud not display header")
	assert.NotContains(t,
		report.Report,
		"Auto-correctable", "stdout shoud not display header")

	listOfStringTheReportMustHave := []string{
		"foo",
		"bar",
		"xyz",
		"0.1.0",
		"0.2.0",
		"0.3.0",
	}
	for _, s := range listOfStringTheReportMustHave {
		assert.Containsf(t, report.Report, s,
			"there is something missing in the stdout: '%s' is missing", s,
		)
	}
}
