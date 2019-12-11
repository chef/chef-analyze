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
	"testing"

	"github.com/stretchr/testify/assert"

	subject "github.com/chef/chef-analyze/pkg/formatter"
	"github.com/chef/chef-analyze/pkg/reporting"
)

func TestNodeReportItemToArray_Nil(t *testing.T) {
	expected := []string{"-", "-", "-", "-"}
	assert.Equal(t, expected, subject.NodeReportItemToArray(nil))
}

func TestNodeReportItemToArray_valid(t *testing.T) {
	cbv := reporting.CookbookVersion{Name: "name", Version: "version"}
	nri := reporting.NodeReportItem{
		Name:             "name",
		ChefVersion:      "16.01",
		OS:               "os",
		OSVersion:        "1.0",
		CookbookVersions: []reporting.CookbookVersion{cbv},
	}
	expected := []string{"name", "16.01", "os v1.0", "name(version)"}
	assert.Equal(t, expected, subject.NodeReportItemToArray(&nri))
}

func TestNodeReportItemToArray_noOS(t *testing.T) {
	cbv := reporting.CookbookVersion{Name: "name", Version: "version"}
	nri := reporting.NodeReportItem{
		Name:             "name",
		ChefVersion:      "16.01",
		OS:               "",
		OSVersion:        "",
		CookbookVersions: []reporting.CookbookVersion{cbv},
	}
	expected := []string{"name", "16.01", "-", "name(version)"}
	assert.Equal(t, expected, subject.NodeReportItemToArray(&nri))
}

func TestNodeReportItemToArray_noCookbooks(t *testing.T) {
	nri := reporting.NodeReportItem{
		Name:             "name",
		ChefVersion:      "16.01",
		OS:               "os",
		OSVersion:        "1.0",
		CookbookVersions: []reporting.CookbookVersion{},
	}
	expected := []string{"name", "16.01", "os v1.0", "-"}
	assert.Equal(t, expected, subject.NodeReportItemToArray(&nri))
}

func TestNodeReportItemToArray_noChefVersion(t *testing.T) {
	cbv := reporting.CookbookVersion{Name: "name", Version: "version"}
	nri := reporting.NodeReportItem{
		Name:             "name",
		ChefVersion:      "",
		OS:               "os",
		OSVersion:        "1.0",
		CookbookVersions: []reporting.CookbookVersion{cbv},
	}
	expected := []string{"name", "-", "os v1.0", "name(version)"}
	assert.Equal(t, expected, subject.NodeReportItemToArray(&nri))
}

func TestFormatNodeReport_noRecords(t *testing.T) {
	nri := []*reporting.NodeReportItem{}
	expected := subject.FormattedResult{"No nodes found to analyze.", ""}
	assert.Equal(t, expected, subject.FormatNodeReport(nri))
}

func TestFormatNodeReport_withRecords(t *testing.T) {
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
	report := subject.FormatNodeReport(nri)

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
		"cookbook1(0.3.0)",
		"cool(0.1.0)",
		"awesome(1.2.3)",
		"13.1.20",
	}
	for _, s := range listOfStringTheReportMustHave {
		assert.Containsf(t, report.Report, s,
			"there is something missing in the stdout: '%s' is missing", s,
		)
	}
}

func TestFormatNodeReport_withRecords_NoCookbooks(t *testing.T) {
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
	report := subject.FormatNodeReport(nri)

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
