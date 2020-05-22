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

package reporting_test

import (
	"fmt"
	"testing"

	"github.com/chef/chef-analyze/pkg/reporting"
	subject "github.com/chef/chef-analyze/pkg/reporting"
	"github.com/stretchr/testify/assert"
)

func TestNodes(t *testing.T) {
	// It's a little less verbose and a little more readable to format
	// this as JSON then convert it where we need it than to create it as a golang map.
	mocksearch := makeMockSearch(mockedNodesSearchRows(), nil)
	results, err := subject.GenerateNodesReport(mocksearch, "")
	// valid results don't mock an error.
	assert.Nil(t, err)

	assert.Equalf(t, 3, len(results),
		"3 input records should give 3 output records, got %d", len(results))

	// TODO - should we test one field or record at a time, or is it safe practice to compare the full results?
	//        there are a lot of things packed into this test if we compare full results.
	expected := []*subject.NodeReportItem{
		&subject.NodeReportItem{Name: "node1", ChefVersion: "12.22", OS: "windows", OSVersion: "10.1",
			CookbookVersions: []subject.CookbookVersion{
				subject.CookbookVersion{Name: "mycookbook", Version: "1.0"}},
		},
		&subject.NodeReportItem{Name: "node2", ChefVersion: "13.11", OS: "", OSVersion: "",
			CookbookVersions: []subject.CookbookVersion{
				subject.CookbookVersion{Name: "mycookbook", Version: "1.0"},
				subject.CookbookVersion{Name: "test", Version: "9.9"},
			},
		},
		&subject.NodeReportItem{Name: "node3", ChefVersion: "15.00", OS: "ubuntu", OSVersion: "16.04",
			CookbookVersions: nil},
	}

	if assert.Equal(t, len(expected), len(results)) {
		for _, e := range expected {
			found := false
			for _, r := range results {
				if e.Name == r.Name {
					found = true
					equalsNodeReportItem(t, e, r)
					break
				}
			}
			assert.Truef(t, found, "Did not find expected NodeReportItem %s", e.Name)
		}
	}
}

func TestErrorResult(t *testing.T) {
	expectedError := fmt.Errorf("error here")
	mocksearch := makeMockSearch("", expectedError)
	report, err := subject.GenerateNodesReport(mocksearch, "")
	if assert.NotNil(t, err) {
		assert.Equal(t, "unable to get node(s) information: error here", err.Error())
		assert.Nil(t, report)
	}
}

func TestCookbookVersionString(t *testing.T) {
	cbv := subject.CookbookVersion{Name: "name", Version: "version"}
	assert.Equal(t, "name(version)", cbv.String())

	cbv = subject.CookbookVersion{Name: "name", Version: ""}
	assert.Equal(t, "name()", cbv.String())

	cbv = subject.CookbookVersion{Name: "", Version: "version"}
	assert.Equal(t, "(version)", cbv.String())
}

func TestNodeReportItemOSVersionPretty(t *testing.T) {
	cases := []struct {
		expected string
		report   subject.NodeReportItem
	}{
		{expected: "os v1.0",
			report: subject.NodeReportItem{OS: "os", OSVersion: "1.0"}},
		{expected: "darwin v10.14.1",
			report: subject.NodeReportItem{OS: "darwin", OSVersion: "10.14.1"}},
		{expected: "windows v200.33",
			report: subject.NodeReportItem{OS: "windows", OSVersion: "200.33"}},
		{expected: "",
			report: subject.NodeReportItem{}},
	}
	for _, kase := range cases {
		assert.Equal(t, kase.expected, kase.report.OSVersionPretty())
	}
}

func TestNodeReportItemCookbooksList(t *testing.T) {
	cases := []struct {
		expected []string
		report   subject.NodeReportItem
	}{
		{expected: []string{"foo(1.0)"},
			report: subject.NodeReportItem{
				CookbookVersions: []subject.CookbookVersion{
					subject.CookbookVersion{Name: "foo", Version: "1.0"}},
			}},
		{expected: []string{"a(1.0)", "b(9.9)", "c(123.123.123)"},
			report: subject.NodeReportItem{
				CookbookVersions: []subject.CookbookVersion{
					subject.CookbookVersion{Name: "a", Version: "1.0"},
					subject.CookbookVersion{Name: "b", Version: "9.9"},
					subject.CookbookVersion{Name: "c", Version: "123.123.123"},
				},
			},
		},
		{expected: []string{"mycookbook(1.0)", "test(9.9)"},
			report: subject.NodeReportItem{
				CookbookVersions: []subject.CookbookVersion{
					subject.CookbookVersion{Name: "mycookbook", Version: "1.0"},
					subject.CookbookVersion{Name: "test", Version: "9.9"},
				},
			},
		},
		{expected: []string{},
			report: subject.NodeReportItem{CookbookVersions: nil}},
	}
	for _, kase := range cases {
		assert.Equal(t, kase.expected, kase.report.CookbooksList())
	}
}

func TestCookbookByNameVersion_Len(t *testing.T) {
	crs := []subject.CookbookVersion{
		subject.CookbookVersion{Name: "cookbook_name"},
	}
	assert.Equal(t, 1, subject.CookbookByNameVersion.Len(crs))
}

func TestCookbookByNameVersion_Swap(t *testing.T) {
	crs := []subject.CookbookVersion{
		subject.CookbookVersion{Name: "cb1"},
		subject.CookbookVersion{Name: "cb2"},
	}
	expected := []subject.CookbookVersion{
		subject.CookbookVersion{Name: "cb2"},
		subject.CookbookVersion{Name: "cb1"},
	}
	subject.CookbookByNameVersion.Swap(crs, 0, 1)
	assert.Equal(t, expected, crs)
}

func TestCookbookByNameVersion_Less(t *testing.T) {
	crs := []subject.CookbookVersion{
		subject.CookbookVersion{Name: "cb1", Version: "1.0.1"},
		subject.CookbookVersion{Name: "cb2", Version: "1.0.1"},
		subject.CookbookVersion{Name: "cb2", Version: "1.0.0"},
		subject.CookbookVersion{Name: "cb1", Version: "1.0.0"},
		subject.CookbookVersion{Name: "CB1", Version: "1.0.0"},
		subject.CookbookVersion{Name: "CB2", Version: "1.0.0"},
	}
	assert.True(t, subject.CookbookByNameVersion.Less(crs, 0, 1))
	assert.True(t, subject.CookbookByNameVersion.Less(crs, 0, 2))
	assert.False(t, subject.CookbookByNameVersion.Less(crs, 0, 3))
	assert.False(t, subject.CookbookByNameVersion.Less(crs, 1, 0))
	assert.False(t, subject.CookbookByNameVersion.Less(crs, 1, 2))
	assert.False(t, subject.CookbookByNameVersion.Less(crs, 1, 3))
	assert.False(t, subject.CookbookByNameVersion.Less(crs, 2, 0))
	assert.True(t, subject.CookbookByNameVersion.Less(crs, 2, 1))
	assert.False(t, subject.CookbookByNameVersion.Less(crs, 2, 3))
	assert.True(t, subject.CookbookByNameVersion.Less(crs, 3, 0))
	assert.True(t, subject.CookbookByNameVersion.Less(crs, 3, 1))
	assert.True(t, subject.CookbookByNameVersion.Less(crs, 3, 2))
	assert.False(t, subject.CookbookByNameVersion.Less(crs, 4, 3))
	assert.False(t, subject.CookbookByNameVersion.Less(crs, 5, 2))
}

func TestNodeRecordsByName_Len(t *testing.T) {
	crs := []*subject.NodeReportItem{
		&subject.NodeReportItem{Name: "node_name"},
	}
	assert.Equal(t, 1, subject.NodeRecordsByName.Len(crs))
}

func TestNodeRecordsByName_Swap(t *testing.T) {
	crs := []*subject.NodeReportItem{
		&subject.NodeReportItem{Name: "n1"},
		&subject.NodeReportItem{Name: "n2"},
	}
	expected := []*subject.NodeReportItem{
		&subject.NodeReportItem{Name: "n2"},
		&subject.NodeReportItem{Name: "n1"},
	}
	subject.NodeRecordsByName.Swap(crs, 0, 1)
	assert.Equal(t, expected, crs)
}

func TestNodeRecordsByName_Less(t *testing.T) {
	nodes := []*subject.NodeReportItem{
		&subject.NodeReportItem{Name: "n1"},
		&subject.NodeReportItem{Name: "n2"},
		&subject.NodeReportItem{Name: "N2"},
		&subject.NodeReportItem{Name: "N1"},
	}
	assert.True(t, subject.NodeRecordsByName.Less(nodes, 0, 1))
	assert.True(t, subject.NodeRecordsByName.Less(nodes, 0, 2))
	assert.False(t, subject.NodeRecordsByName.Less(nodes, 0, 3))
	assert.False(t, subject.NodeRecordsByName.Less(nodes, 1, 2))
	assert.False(t, subject.NodeRecordsByName.Less(nodes, 1, 3))
	assert.False(t, subject.NodeRecordsByName.Less(nodes, 0, 3))
	assert.False(t, subject.NodeRecordsByName.Less(nodes, 2, 3))
}

func equalsCookbookVersionsArray(t *testing.T, expected, actual []subject.CookbookVersion) {
	if assert.Equal(t, len(expected), len(actual)) {
		for _, a := range actual {
			assert.Containsf(t, expected, a, "Missing CookbookVersion item")
		}
	}
}

func equalsNodeReportItem(t *testing.T, expected *subject.NodeReportItem, actual *subject.NodeReportItem) {
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.ChefVersion, actual.ChefVersion)
	assert.Equal(t, expected.OS, actual.OS)
	assert.Equal(t, expected.OSVersion, actual.OSVersion)
	equalsCookbookVersionsArray(t, expected.CookbookVersions, actual.CookbookVersions)
}

func mockedEmptyNodesSearchRows() string {
	return `[]`
}

func mockedNodesSearchRows() string {
	return `[
  {
    "data" : {
      "name" : "node1",
      "chef_version": "12.22",
      "os" : "windows",
      "os_version": "10.1",
      "cookbooks" : {
        "mycookbook" : {
          "version" : "1.0"
        }
      }
    }
  },
  {
    "data" : {
      "name" : "node2",
      "chef_version": "13.11",
      "os" : null,
      "os_version": null,
      "cookbooks" : {
        "mycookbook" : { "version" : "1.0" },
        "test" : { "version" : "9.9" }
      }
    }
  },
  {
    "data" : {
      "name" : "node3",
      "chef_version": "15.00",
      "os" : "ubuntu",
      "os_version": "16.04",
      "cookbooks" : null
    }
  }
]`
}

func TestNodeReportItem_GetPolicyGroup(t *testing.T) {
	type fields struct {
		Name             string
		ChefVersion      string
		OS               string
		OSVersion        string
		CookbookVersions []reporting.CookbookVersion
		PolicyGroup      string
		Policy           string
		PolicyRev        string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// Test cases
		{"TestGetPolicyGroupAbsent", fields{Name: "node1", ChefVersion: "16.0", OS: "Windows"}, "no group"},
		{"TestGetPolicyGroup", fields{Name: "node2", ChefVersion: "16.0", OS: "Windows", OSVersion: "10.001", PolicyGroup: "preprod"}, "preprod"},
		{"TestGetPolicyGroupEmpty", fields{Name: "node3", ChefVersion: "16.0", OS: "Windows", OSVersion: "10.001", PolicyGroup: ""}, "no group"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nri := &reporting.NodeReportItem{
				Name:             tt.fields.Name,
				ChefVersion:      tt.fields.ChefVersion,
				OS:               tt.fields.OS,
				OSVersion:        tt.fields.OSVersion,
				CookbookVersions: tt.fields.CookbookVersions,
				PolicyGroup:      tt.fields.PolicyGroup,
				Policy:           tt.fields.Policy,
				PolicyRev:        tt.fields.PolicyRev,
			}
			if got := nri.GetPolicyGroup(); got != tt.want {
				assert.Equal(t, got, tt.want)
			}
		})
	}
}

func TestNodeReportItem_GetPolicy(t *testing.T) {
	type fields struct {
		Name             string
		ChefVersion      string
		OS               string
		OSVersion        string
		CookbookVersions []reporting.CookbookVersion
		PolicyGroup      string
		Policy           string
		PolicyRev        string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// Test cases
		{"TestGetPolicyAbsent", fields{Name: "node1", ChefVersion: "16.0", OS: "Windows"}, "no policy"},
		{"TestGetPolicy", fields{Name: "node1", ChefVersion: "16.0", OS: "Windows", OSVersion: "10.001", PolicyGroup: "preprod",
			Policy: "seven-zip", PolicyRev: "12x23343434sx9000ttt00000"}, "seven-zip"},
		{"TestGetPolicyEmpty", fields{Name: "node1", ChefVersion: "16.0", OS: "Windows", OSVersion: "10.001", PolicyGroup: "", Policy: ""}, "no policy"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nri := &reporting.NodeReportItem{
				Name:             tt.fields.Name,
				ChefVersion:      tt.fields.ChefVersion,
				OS:               tt.fields.OS,
				OSVersion:        tt.fields.OSVersion,
				CookbookVersions: tt.fields.CookbookVersions,
				PolicyGroup:      tt.fields.PolicyGroup,
				Policy:           tt.fields.Policy,
				PolicyRev:        tt.fields.PolicyRev,
			}
			if got := nri.GetPolicy(); got != tt.want {
				assert.Equal(t, got, tt.want)
			}
		})
	}
}

func TestNodeReportItem_GetPolicyWithRev(t *testing.T) {
	type fields struct {
		Name             string
		ChefVersion      string
		OS               string
		OSVersion        string
		CookbookVersions []reporting.CookbookVersion
		PolicyGroup      string
		Policy           string
		PolicyRev        string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// Test cases
		{"TestGetPolicyWithRevAbsent", fields{Name: "node1", ChefVersion: "16.0", OS: "Windows"}, "no policy"},
		{"TestGetPolicyWithRev", fields{Name: "node1", ChefVersion: "16.0", OS: "Windows", OSVersion: "10.001", PolicyGroup: "preprod",
			Policy: "seven-zip", PolicyRev: "12x23343434sx9000ttt00000"}, "seven-zip(rev 12x23343434sx9000ttt00000)"},
		{"TestGetPolicyWithRevEmpty", fields{Name: "node1", ChefVersion: "16.0", OS: "Windows", OSVersion: "10.001", PolicyGroup: "", Policy: ""}, "no policy"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nri := &reporting.NodeReportItem{
				Name:             tt.fields.Name,
				ChefVersion:      tt.fields.ChefVersion,
				OS:               tt.fields.OS,
				OSVersion:        tt.fields.OSVersion,
				CookbookVersions: tt.fields.CookbookVersions,
				PolicyGroup:      tt.fields.PolicyGroup,
				Policy:           tt.fields.Policy,
				PolicyRev:        tt.fields.PolicyRev,
			}
			if got := nri.GetPolicyWithRev(); got != tt.want {
				assert.Equal(t, got, tt.want)
			}
		})
	}
}
