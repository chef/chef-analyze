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
	"reflect"
	"testing"

	subject "github.com/chef/chef-analyze/pkg/reporting"
	chef "github.com/chef/go-chef"
)

// This module contains mocks implementations for the client interfaces
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
	createDirOnDownload      bool
}

func (cm CookbookMock) ListAvailableVersions(limit string) (chef.CookbookListResult, error) {
	return cm.desiredCookbookList, cm.desiredCookbookListError
}

func (cm CookbookMock) DownloadTo(name, version, localDir string) error {
	if cm.createDirOnDownload {
		dirToMock := filepath.Join(localDir, fmt.Sprintf("%s-%s", name, version))
		err := os.MkdirAll(dirToMock, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
	return cm.desiredDownloadError
}

type CBAMock struct {
	desiredCBAList       chef.CBAGetResponse
	desiredCBAListError  error
	desiredDownloadError error
	createDirOnDownload  bool
}

func (cm CBAMock) List() (chef.CBAGetResponse, error) {
	return cm.desiredCBAList, cm.desiredCBAListError
}

func (cm CBAMock) DownloadTo(name, id, localDir string) error {
	if cm.createDirOnDownload {
		dirToMock := filepath.Join(localDir, fmt.Sprintf("%v-%v", name, id[0:20]))
		err := os.MkdirAll(dirToMock, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
	return cm.desiredDownloadError
}

type PolicyGroupMock struct {
	desiredPolicyGroupList      chef.PolicyGroupGetResponse
	desiredPolicyGroupListError error
}

func (pgm PolicyGroupMock) List() (chef.PolicyGroupGetResponse, error) {
	return pgm.desiredPolicyGroupList, pgm.desiredPolicyGroupListError
}

type PolicyMock struct {
	desiredPolicyDetail      chef.RevisionDetailsResponse
	desiredPolicyDetailError error
}

func (pm PolicyMock) GetRevisionDetails(policyName string, revisionID string) (chef.RevisionDetailsResponse, error) {
	return pm.desiredPolicyDetail, pm.desiredPolicyDetailError
}

type DataBagMock struct {
	desiredDataBagList      chef.DataBagListResult
	desiredDataBagListError error

	desiredDataBagItem      chef.DataBagItem
	desiredDataBagItemError error

	desiredDataBagItemList      chef.DataBagListResult
	desiredDataBagItemListError error
}

func (dm DataBagMock) GetItem(databagName string, databagItemName string) (chef.DataBagItem, error) {
	return dm.desiredDataBagItem, dm.desiredDataBagItemError
}

func (dm DataBagMock) ListItems(name string) (*chef.DataBagListResult, error) {
	return &dm.desiredDataBagItemList, dm.desiredDataBagItemListError
}

func (dm DataBagMock) List() (*chef.DataBagListResult, error) {
	return &dm.desiredDataBagList, dm.desiredDataBagListError
}

func newMockDataBagLister(res chef.DataBagListResult, err error) *DataBagMock {
	return &DataBagMock{desiredDataBagList: res, desiredDataBagListError: err}
}

func newMockDataBagItemGetter(res chef.DataBagItem, err error) *DataBagMock {
	return &DataBagMock{desiredDataBagItem: res, desiredDataBagItemError: err}
}

func newMockDataBagItemLister(res chef.DataBagListResult, err error) *DataBagMock {
	return &DataBagMock{desiredDataBagItemList: res, desiredDataBagItemListError: err}
}

func newMockCookbook(cookbookList chef.CookbookListResult, desiredCbListErr, desiredDownloadErr error) *CookbookMock {
	return &CookbookMock{
		desiredCookbookList:      cookbookList,
		desiredCookbookListError: desiredCbListErr,
		desiredDownloadError:     desiredDownloadErr,
		createDirOnDownload:      true,
	}
}

func newMockCookbookArtifact(CBAList chef.CBAGetResponse, desiredCbListErr, desiredDownloadErr error) *CBAMock {
	return &CBAMock{
		desiredCBAList:       CBAList,
		desiredCBAListError:  desiredCbListErr,
		desiredDownloadError: desiredDownloadErr,
		createDirOnDownload:  true,
	}
}

func newMockPolicyGroup(policyGroupList chef.PolicyGroupGetResponse, desiredPolicyGroupListErr error) *PolicyGroupMock {
	return &PolicyGroupMock{
		desiredPolicyGroupList:      policyGroupList,
		desiredPolicyGroupListError: desiredPolicyGroupListErr,
	}
}

func newMockPolicy(policyDetail chef.RevisionDetailsResponse, desiredPolicyDetailErr error) *PolicyMock {
	return &PolicyMock{
		desiredPolicyDetail:      policyDetail,
		desiredPolicyDetailError: desiredPolicyDetailErr,
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

type NodeMock struct {
	Error error
	Node  chef.Node
}

func (nm NodeMock) Get(name string) (chef.Node, error) {
	if nm.Error != nil {
		return chef.Node{}, nm.Error
	}
	return nm.Node, nil
}

type RoleMock struct {
	Error error
	Role  *chef.Role
}

func (rm RoleMock) Get(name string) (*chef.Role, error) {
	if rm.Error != nil {
		return nil, rm.Error
	}
	return rm.Role, nil
}

type EnvMock struct {
	Error error
	Env   *chef.Environment
}

func (em EnvMock) Get(name string) (*chef.Environment, error) {
	if em.Error != nil {
		return nil, em.Error
	}
	return em.Env, nil
}

func TestNewChefAnalyzeClient(t *testing.T) {
	type args struct {
		chefClient *chef.Client
	}
	chefCookbook := chef.CookbookService{}

	tests := []struct {
		name string
		args args
		want *subject.ChefAnalyzeClient
	}{
		{"TestWithCookbookService", args{chefClient: &chef.Client{Cookbooks: &chefCookbook}}, &subject.ChefAnalyzeClient{Cookbooks: &chefCookbook}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := subject.NewChefAnalyzeClient(tt.args.chefClient); !reflect.DeepEqual(got.Cookbooks, tt.want.Cookbooks) {
				t.Errorf("NewChefAnalyzeClient() = %v, want %v", got, tt.want)
			}
		})
	}
}
