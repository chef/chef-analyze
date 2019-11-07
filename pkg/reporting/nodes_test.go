package reporting_test

import (
	"encoding/json"
	"testing"

	"fmt"

	"github.com/chef/chef-analyze/pkg/credentials"
	"github.com/chef/chef-analyze/pkg/reporting"
	chef "github.com/chef/go-chef"
	"github.com/stretchr/testify/assert"
)

type PartialSearchMock struct {
	desiredError   error
	desiredResults chef.SearchResult
}

func (psm PartialSearchMock) PartialExec(idx, statement string,
	params map[string]interface{}) (res chef.SearchResult, err error) {

	if psm.desiredError == nil {
		return psm.desiredResults, nil
	}
	return psm.desiredResults, psm.desiredError
}

func TestNodes(t *testing.T) {
	creds := credentials.Credentials{}

	cfg := &reporting.Reporting{Credentials: creds}
	testValidResultsWithNulls(t, cfg)
	testErrorResult(t, cfg)
}

func testErrorResult(t *testing.T, cfg *reporting.Reporting) {
	expectedError := fmt.Errorf("error here")
	mocksearch := makeMockSearch("", expectedError)
	_, err := reporting.Nodes(cfg, mocksearch)
	assert.NotNil(t, err)
}

func testValidResultsWithNulls(t *testing.T, cfg *reporting.Reporting) {
	// It's a little less verbose and a little more readable to format
	// this as JSON then convert it where we need it than to create it as a golang map.
	rows := `[
    { "data" : {"name" : "node1", "chef_version": "12.22", "os" : "windows", "os_version": "10.1",   "cookbooks" : { "mycookbook" : { "version" : "1.0" } } } },
    { "data" : {"name" : "node2", "chef_version": "13.11", "os" : null,       "os_version":    null, "cookbooks" : { "mycookbook" : { "version" : "1.0" } , "test" : { "version" : "9.9" } } } },
		{ "data" : {"name" : "node3", "chef_version": "15.00", "os" : "ubuntu", "os_version": "16.04", "cookbooks" : null }}
	]`
	mocksearch := makeMockSearch(rows, nil)
	results, err := reporting.Nodes(cfg, mocksearch)
	if err != nil {
		panic(err)
	} // valid results don't mock an error.

	if len(results) != 3 {
		t.Errorf("3 input records should give 3 output records, got %d", len(results))
	}

	// TODO - should we test one field or record at a time, or is it safe practice to compare the full results?
	//        there are a lot of things packed into this test if we compare full results.
	expected := []reporting.NodeReportItem{
		reporting.NodeReportItem{Name: "node1", ChefVersion: "12.22", OS: "windows", OSVersion: "10.1",
			CookbookVersions: []reporting.CookbookVersion{
				reporting.CookbookVersion{Name: "mycookbook", Version: "1.0"}},
		},
		reporting.NodeReportItem{Name: "node2", ChefVersion: "13.11", OS: "", OSVersion: "",
			CookbookVersions: []reporting.CookbookVersion{
				reporting.CookbookVersion{Name: "mycookbook", Version: "1.0"},
				reporting.CookbookVersion{Name: "test", Version: "9.9"},
			},
		},
		reporting.NodeReportItem{Name: "node3", ChefVersion: "15.00", OS: "ubuntu", OSVersion: "16.04",
			CookbookVersions: nil},
	}

	assert.Equal(t, results, expected)
}

func makeMockSearch(searchResultJSON string, desiredError error) PartialSearchMock {
	var convertedSearchResult []interface{}
	if desiredError == nil {
		rawRows := json.RawMessage(searchResultJSON)
		bytes, err := rawRows.MarshalJSON()
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(bytes, &convertedSearchResult)
		if err != nil {
			panic(err)
		}
	}
	result := chef.SearchResult{Rows: convertedSearchResult}
	return PartialSearchMock{desiredError: desiredError, desiredResults: result}
}
