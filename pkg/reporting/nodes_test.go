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

	"github.com/stretchr/testify/assert"

	subject "github.com/chef/chef-analyze/pkg/reporting"
)

func TestNodes(t *testing.T) {
	// It's a little less verbose and a little more readable to format
	// this as JSON then convert it where we need it than to create it as a golang map.
	mocksearch := makeMockSearch(mockedNodesSearchRows(), nil)
	results, err := subject.Nodes(mocksearch)
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
	report, err := subject.Nodes(mocksearch)
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
