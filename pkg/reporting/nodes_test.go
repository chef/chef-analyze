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

	"github.com/chef/go-libs/credentials"
	"github.com/stretchr/testify/assert"

	subject "github.com/chef/chef-analyze/pkg/reporting"
)

func TestNodes(t *testing.T) {
	creds := credentials.Credentials{}

	cfg := &subject.Reporting{Credentials: creds}
	testValidResultsWithNulls(t, cfg)
	testErrorResult(t, cfg)
}

func TestCookbookVersionString(t *testing.T) {
	cbv := subject.CookbookVersion{Name: "name", Version: "version"}
	assert.Equal(t, cbv.String(), "name(version)")
}

func TestNodeReportItemArray(t *testing.T) {
	cbv := subject.CookbookVersion{Name: "name", Version: "version"}
	nri := subject.NodeReportItem{
		Name:             "name",
		ChefVersion:      "chefversion",
		OS:               "os",
		OSVersion:        "osversion",
		CookbookVersions: []subject.CookbookVersion{cbv},
	}
	expected := []string{"name", "chefversion", "os", "osversion", "name(version)"}
	assert.Equal(t, expected, nri.Array())
}

func testErrorResult(t *testing.T, cfg *subject.Reporting) {
	expectedError := fmt.Errorf("error here")
	mocksearch := makeMockSearch("", expectedError)
	_, err := subject.Nodes(cfg, mocksearch)
	assert.NotNil(t, err)
}

func equalsCookbookVersionsArray(t *testing.T, expected, actual []subject.CookbookVersion) {
	if assert.Equal(t, len(expected), len(actual)) {
		for _, a := range actual {
			assert.Containsf(t, expected, a, "Missing CookbookVersion item")
		}
	}
}

func equalsNodeReportItem(t *testing.T, expected, actual subject.NodeReportItem) {
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.ChefVersion, actual.ChefVersion)
	assert.Equal(t, expected.OS, actual.OS)
	assert.Equal(t, expected.OSVersion, actual.OSVersion)
	equalsCookbookVersionsArray(t, expected.CookbookVersions, actual.CookbookVersions)
}

func testValidResultsWithNulls(t *testing.T, cfg *subject.Reporting) {
	// It's a little less verbose and a little more readable to format
	// this as JSON then convert it where we need it than to create it as a golang map.
	mocksearch := makeMockSearch(mockedNodesSearchRows(), nil)
	results, err := subject.Nodes(cfg, mocksearch)
	if err != nil {
		panic(err)
	} // valid results don't mock an error.

	if len(results) != 3 {
		t.Errorf("3 input records should give 3 output records, got %d", len(results))
	}

	// TODO - should we test one field or record at a time, or is it safe practice to compare the full results?
	//        there are a lot of things packed into this test if we compare full results.
	expected := []subject.NodeReportItem{
		subject.NodeReportItem{Name: "node1", ChefVersion: "12.22", OS: "windows", OSVersion: "10.1",
			CookbookVersions: []subject.CookbookVersion{
				subject.CookbookVersion{Name: "mycookbook", Version: "1.0"}},
		},
		subject.NodeReportItem{Name: "node2", ChefVersion: "13.11", OS: "", OSVersion: "",
			CookbookVersions: []subject.CookbookVersion{
				subject.CookbookVersion{Name: "mycookbook", Version: "1.0"},
				subject.CookbookVersion{Name: "test", Version: "9.9"},
			},
		},
		subject.NodeReportItem{Name: "node3", ChefVersion: "15.00", OS: "ubuntu", OSVersion: "16.04",
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
