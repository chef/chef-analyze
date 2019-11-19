package reporting_test

import (
	"encoding/json"

	chef "github.com/chef/go-chef"
)

// This contains mocks implementations for the client interfaces
// defined in client.go

type SearchMock struct {
	desiredError   error
	desiredResults chef.SearchResult
}

func (sm SearchMock) PartialExec(idx, statement string,
	params map[string]interface{}) (res chef.SearchResult, err error) {
	return sm.desiredResults, sm.desiredError
}

type CookbookMock struct {
	desiredCookbookListError error
	desiredDownloadError     error
	desiredCookbookList      chef.CookbookListResult
}

func (cm CookbookMock) ListAvailableVersions(limit string) (chef.CookbookListResult, error) {
	return cm.desiredCookbookList, cm.desiredCookbookListError
}

func (cm CookbookMock) DownloadTo(name, version, localDir string) error {
	return cm.desiredDownloadError
}

func makeMockSearch(searchResultJSON string, desiredError error) SearchMock {
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
	return SearchMock{desiredError: desiredError, desiredResults: result}
}
