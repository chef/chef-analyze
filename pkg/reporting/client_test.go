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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	chef "github.com/chef/go-chef"
)

// This contains mocks implementations for the client interfaces
// defined in client.go

type SearchMock struct {
	desiredResults chef.SearchResult
	desiredError   error
}

func (sm SearchMock) PartialExec(idx, statement string, params map[string]interface{}) (res chef.SearchResult, err error) {
	return sm.desiredResults, sm.desiredError
}

type CookbookMock struct {
	desiredCookbookList      chef.CookbookListResult
	desiredCookbookListError error
	desiredDownloadError     error
}

func (cm CookbookMock) ListAvailableVersions(limit string) (chef.CookbookListResult, error) {
	return cm.desiredCookbookList, cm.desiredCookbookListError
}

func (cm CookbookMock) DownloadTo(name, version, localDir string) error {
	var (
		dirToMock = filepath.Join(localDir, fmt.Sprintf("%s-%s", name, version))
		err       = os.MkdirAll(dirToMock, os.ModePerm)
	)
	if err != nil {
		panic(err)
	}
	return cm.desiredDownloadError
}

func newMockCookbook(cookbookList chef.CookbookListResult, desiredCbListErr, desiredDownloadErr error) *CookbookMock {
	return &CookbookMock{
		desiredCookbookList:      cookbookList,
		desiredCookbookListError: desiredCbListErr,
		desiredDownloadError:     desiredDownloadErr,
	}
}

func makeMockSearch(searchResultJSON string, desiredError error) *SearchMock {
	var convertedSearchResult []interface{}

	if searchResultJSON != "" {
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

	return &SearchMock{
		desiredResults: chef.SearchResult{Rows: convertedSearchResult},
		desiredError:   desiredError,
	}
}
