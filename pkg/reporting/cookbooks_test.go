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

package reporting_test

import (
	"fmt"
	"os"
	"testing"

	chef "github.com/chef/go-chef"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	subject "github.com/chef/chef-analyze/pkg/reporting"
)

const Workers = 50

func TestCookbooksRecordCorrectableAndOffenses(t *testing.T) {
	csr := subject.CookbookRecord{
		Files: []subject.CookbookFile{
			subject.CookbookFile{
				Offenses: []subject.CookstyleOffense{
					subject.CookstyleOffense{Correctable: false},
				},
			},
			subject.CookbookFile{
				Offenses: []subject.CookstyleOffense{
					subject.CookstyleOffense{Correctable: true},
					subject.CookstyleOffense{Correctable: false},
					subject.CookstyleOffense{Correctable: true},
					subject.CookstyleOffense{Correctable: true},
				},
			},
		},
	}
	assert.Equal(t, 3, csr.NumCorrectable())
	assert.Equal(t, 5, csr.NumOffenses())
}

func TestCookbooksRecordNumNodesAffected(t *testing.T) {
	rec := subject.CookbookRecord{}
	assert.Equal(t, 0, rec.NumNodesAffected())
	rec = subject.CookbookRecord{Nodes: []string{}}
	assert.Equal(t, 0, rec.NumNodesAffected())
	rec = subject.CookbookRecord{Nodes: []string{"a", "b", "c"}}
	assert.Equal(t, 3, rec.NumNodesAffected())
}

func TestCookbooksEmpty(t *testing.T) {

	chefAnalyzeClient := subject.ChefAnalyzeClient{

		Cookbooks:         newMockCookbook(chef.CookbookListResult{}, nil, nil),
		CookbookArtifacts: newMockCookbookArtifact(chef.CBAGetResponse{}, nil, nil),
		PolicyGroups:      newMockPolicyGroup(chef.PolicyGroupGetResponse{}, nil),
		Policies:          newMockPolicy(chef.RevisionDetailsResponse{}, nil),
		Search:            makeMockSearch("[]", nil),
	}

	cbr, err := subject.NewCookbooksReport(
		&chefAnalyzeClient,
		false,
		false,
		Workers,
		"",    // no filter
		false, // anonymize
	)
	assert.Nil(t, err)
	if assert.NotNil(t, cbr) {
		assert.Empty(t, cbr.Records)
		assert.Equal(t, 0, cbr.TotalCookbooks)
	}
}

// Given a valid set of cookbooks in use by nodes,
// And the --only-unused flag is provided,
// Verify that the result set is as expected.
func TestCookbooksInUseDisplayOnlyUnused(t *testing.T) {
	cookbookList := chef.CookbookListResult{
		"foo": chef.CookbookVersions{
			Versions: []chef.CookbookVersion{
				chef.CookbookVersion{Version: "0.1.0"},
				chef.CookbookVersion{Version: "0.2.0"},
				chef.CookbookVersion{Version: "0.3.0"},
			},
		},
		"bar": chef.CookbookVersions{
			Versions: []chef.CookbookVersion{
				chef.CookbookVersion{Version: "0.1.0"},
			},
		},
	}

	CBAList := chef.CBAGetResponse{
		"alpha": chef.CBA{
			CBAVersions: []chef.CBAVersion{
				chef.CBAVersion{Identifier: "123456789012345678901234567890xyz"},
			},
		},
		"gama": chef.CBA{
			CBAVersions: []chef.CBAVersion{
				chef.CBAVersion{Identifier: "abc456789012345678901234567890xyz"},
			},
		},
	}

	policyGroupList := chef.PolicyGroupGetResponse{
		"my-policygroup": chef.PolicyGroup{
			Policies: map[string]chef.Revision{
				"my-policy": chef.Revision{
					"revision_id": "123xyz",
				},
			},
		},
	}

	policyDetail := chef.RevisionDetailsResponse{
		Name:       "my-policy",
		RevisionID: "123xyz",
		CookbookLocks: map[string]chef.CookbookLock{
			"alpha": chef.CookbookLock{
				Identifier: "123456789012345678901234567890xyz",
			},
			"gama": chef.CookbookLock{
				Identifier: "abc456789012345678901234567890xyz",
			},
		},
	}

	chefAnalyzeClient := subject.ChefAnalyzeClient{

		Cookbooks:         newMockCookbook(cookbookList, nil, nil),
		CookbookArtifacts: newMockCookbookArtifact(CBAList, nil, nil),
		PolicyGroups:      newMockPolicyGroup(policyGroupList, nil),
		Policies:          newMockPolicy(policyDetail, nil),
		Search:            makeMockSearch(mockedNodesSearchRows(), nil), // nodes are found
	}

	c, err := subject.NewCookbooksReport(
		&chefAnalyzeClient,
		false, // do not download or cookstyle
		true,  // display only unused cookbooks
		Workers,
		"",    // no filter
		false, // anonymize
	)
	assert.Nil(t, err)
	if assert.NotNil(t, c) {
		c.Generate()

		assert.Equal(t, 6, len(c.Progress))
		assert.Equal(t, 6, c.TotalCookbooks)
		assert.Empty(t, c.Records)
	}

	chefAnalyzeClient = subject.ChefAnalyzeClient{

		Cookbooks:         newMockCookbook(cookbookList, nil, nil),
		CookbookArtifacts: newMockCookbookArtifact(CBAList, nil, nil),
		PolicyGroups:      newMockPolicyGroup(policyGroupList, nil),
		Policies:          newMockPolicy(policyDetail, nil),
		Search:            makeMockSearch(mockedNodesSearchRows(), nil), // nodes are found
	}

	// Verify that used cookbooks list is complete and correct
	c, err = subject.NewCookbooksReport(
		&chefAnalyzeClient,
		false, // do not download or cookstyle
		false, // display only used cookbooks, as determined by search results
		Workers,
		"",    // no filter
		false, // anonymize
	)
	assert.Nil(t, err)
	if assert.NotNil(t, c) {
		assert.Equal(t, 6, c.TotalCookbooks)
		c.Generate()
		assert.Equal(t, 6, len(c.Progress))
		if assert.Equal(t, 6, len(c.Records)) {

			var (
				foo   = 0
				bar   = 0
				alpha = 0
				gama  = 0
			)

			for _, rec := range c.Records {
				switch rec.Name {
				case "foo":
					foo++
				case "bar":
					bar++
				case "alpha":
					alpha++
				case "gama":
					gama++
				default:
					t.Fatal("unexpected cookbook name")
				}

				assert.NotEmpty(t, rec.Nodes, "nodes should be found")
			}

			assert.Equal(t, 3, foo, "unexpected number of cookbooks foo")
			assert.Equal(t, 1, bar, "unexpected number of cookbooks bar")
			assert.Equal(t, 1, alpha, "unexpected number of cookbooks alpha")
			assert.Equal(t, 1, gama, "unexpected number of cookbooks gama")
		}
	}
}

// Given a valid set of cookbooks in use by nodes
// And anonymization is required
// Verify that the result set is anonymized as expected.
func TestCookbooksAnonymized(t *testing.T) {
	cookbookList := chef.CookbookListResult{
		"foo": chef.CookbookVersions{
			Versions: []chef.CookbookVersion{
				chef.CookbookVersion{Version: "0.1.0"},
			},
		},
	}

	chefAnalyzeClient := subject.ChefAnalyzeClient{

		Cookbooks:         newMockCookbook(cookbookList, nil, nil),
		CookbookArtifacts: newMockCookbookArtifact(nil, nil, nil),
		PolicyGroups:      newMockPolicyGroup(nil, nil),
		Search:            makeMockSearch(mockedNodesSearchRows(), nil),
	}

	c, err := subject.NewCookbooksReport(
		&chefAnalyzeClient,
		false,
		false, // display only used cookbooks
		Workers,
		"",   // no filter
		true, // anonymize
	)
	assert.Nil(t, err)
	if assert.NotNil(t, c) {
		c.Generate()
		assert.Equal(t, 1, c.TotalCookbooks)
		assert.Equal(t, "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae", c.Records[0].Name)
	}

}

// Given a valid set of cookbooks in use by nodes via policy files
// And anonymization is required
// Verify that the result set is anonymized as expected.
func TestPolicyAnonymized(t *testing.T) {
	CBAList := chef.CBAGetResponse{
		"alpha": chef.CBA{
			CBAVersions: []chef.CBAVersion{
				chef.CBAVersion{Identifier: "123456789012345678901234567890xyz"},
			},
		},
	}

	policyGroupList := chef.PolicyGroupGetResponse{
		"my-policygroup": chef.PolicyGroup{
			Policies: map[string]chef.Revision{
				"my-policy": chef.Revision{
					"revision_id": "123xyz",
				},
			},
		},
	}

	policyDetail := chef.RevisionDetailsResponse{
		Name:       "my-policy",
		RevisionID: "123xyz",
		CookbookLocks: map[string]chef.CookbookLock{
			"alpha": chef.CookbookLock{
				Identifier: "123456789012345678901234567890xyz",
			},
		},
	}

	chefAnalyzeClient := subject.ChefAnalyzeClient{
		Cookbooks:         newMockCookbook(nil, nil, nil),
		CookbookArtifacts: newMockCookbookArtifact(CBAList, nil, nil),
		PolicyGroups:      newMockPolicyGroup(policyGroupList, nil),
		Policies:          newMockPolicy(policyDetail, nil),
		Search:            makeMockSearch(mockedNodesSearchRows(), nil), // nodes are found
	}

	c, err := subject.NewCookbooksReport(
		&chefAnalyzeClient,
		false,
		false, // display only used cookbooks
		Workers,
		"",   // no filter
		true, // anonymize
	)
	assert.Nil(t, err)
	if assert.NotNil(t, c) {
		c.Generate()
		assert.Equal(t, 1, c.TotalCookbooks)
		assert.Equal(t, "8ed3f6ad685b959ead7022518e1af76cd816f8e8ec7ccdda1ed4018e8f2223f8", c.Records[0].Name)
		assert.Equal(t, "1f94add5d386d19dc3ef8a5c4b5b740c495937f2bbf60996f9082fbb100402eb", c.Records[0].PolicyGroup)
		assert.Equal(t, "bc575ae6b32fc77c9d06c4a71eb56880aa8f27e1b839e6c97f7d7645951c1a41", c.Records[0].Policy)
	}

}

// Given a valid set of cookbooks that are not being used by any node,
// And the --only-unused flag is provided,
// Verify that the result set is as expected.
func TestCookbooksNotUsedDisplayOnlyUnused(t *testing.T) {
	cookbookList := chef.CookbookListResult{
		"foo": chef.CookbookVersions{
			Versions: []chef.CookbookVersion{
				chef.CookbookVersion{Version: "0.1.0"},
				chef.CookbookVersion{Version: "0.2.0"},
				chef.CookbookVersion{Version: "0.3.0"},
			},
		},
		"bar": chef.CookbookVersions{
			Versions: []chef.CookbookVersion{
				chef.CookbookVersion{Version: "0.1.0"},
			},
		},
	}

	CBAList := chef.CBAGetResponse{
		"alpha": chef.CBA{
			CBAVersions: []chef.CBAVersion{
				chef.CBAVersion{Identifier: "123456789012345678901234567890xyz"},
			},
		},
		"gama": chef.CBA{
			CBAVersions: []chef.CBAVersion{
				chef.CBAVersion{Identifier: "abc456789012345678901234567890xyz"},
			},
		},
	}

	policyGroupList := chef.PolicyGroupGetResponse{
		"my-policygroup": chef.PolicyGroup{
			Policies: map[string]chef.Revision{
				"my-policy": chef.Revision{
					"revision_id": "123xyz",
				},
			},
		},
	}

	policyDetail := chef.RevisionDetailsResponse{
		Name:       "my-policy",
		RevisionID: "123xyz",
		CookbookLocks: map[string]chef.CookbookLock{
			"alpha": chef.CookbookLock{
				Identifier: "123456789012345678901234567890xyz",
			},
			"gama": chef.CookbookLock{
				Identifier: "abc456789012345678901234567890xyz",
			},
		},
	}

	chefAnalyzeClient := subject.ChefAnalyzeClient{
		Cookbooks:         newMockCookbook(cookbookList, nil, nil),
		CookbookArtifacts: newMockCookbookArtifact(CBAList, nil, nil),
		PolicyGroups:      newMockPolicyGroup(policyGroupList, nil),
		Policies:          newMockPolicy(policyDetail, nil),
		Search:            makeMockSearch("[]", nil), // no nodes are returned
	}

	c, err := subject.NewCookbooksReport(
		&chefAnalyzeClient,
		false,
		true, // display only unused cookbooks
		Workers,
		"",    // no filter
		false, // anonymize
	)
	assert.Nil(t, err)
	c.Generate()
	assert.Equal(t, 6, len(c.Progress)) // verify all progress events arrived

	if assert.NotNil(t, c) {
		assert.Equal(t, 6, c.TotalCookbooks)
		if assert.Equal(t, 6, len(c.Records)) {
			var (
				foo   = 0
				bar   = 0
				alpha = 0
				gama  = 0
			)

			for _, rec := range c.Records {
				switch rec.Name {
				case "foo":
					foo++
				case "bar":
					bar++
				case "alpha":
					alpha++
				case "gama":
					gama++
				default:
					t.Fatal("unexpected cookbook name")
				}

				assert.Empty(t, rec.Nodes, "no nodes should be found")
			}

			assert.Equal(t, 3, foo, "unexpected number of cookbooks foo")
			assert.Equal(t, 1, bar, "unexpected number of cookbooks bar")
			assert.Equal(t, 1, alpha, "unexpected number of cookbooks alpha")
			assert.Equal(t, 1, gama, "unexpected number of cookbooks gama")
		}
	}

	chefAnalyzeClient = subject.ChefAnalyzeClient{
		Cookbooks:         newMockCookbook(cookbookList, nil, nil),
		CookbookArtifacts: newMockCookbookArtifact(CBAList, nil, nil),
		PolicyGroups:      newMockPolicyGroup(policyGroupList, nil),
		Policies:          newMockPolicy(policyDetail, nil),
		Search:            makeMockSearch("[]", nil), // no nodes are returned
	}

	c, err = subject.NewCookbooksReport(
		&chefAnalyzeClient,
		false,
		false, // display only used cookbooks
		Workers,
		"",    // no filter
		false, // anonymize
	)
	assert.Nil(t, err)
	if assert.NotNil(t, c) {
		c.Generate()
		assert.Empty(t, c.Records)
		assert.Equal(t, 6, len(c.Progress)) // verify all progress events arrived
		assert.Equal(t, 6, c.TotalCookbooks)
	}

}

// Given a valid set of cookbooks in use by nodes,
// verify that the result set is as expected.
func TestCookbooks(t *testing.T) {
	savedPath := setupBinstubsDir()
	defer os.Setenv("PATH", savedPath)
	defer os.RemoveAll(createConfigToml(t))

	cookbookList := chef.CookbookListResult{
		"foo": chef.CookbookVersions{
			Versions: []chef.CookbookVersion{
				chef.CookbookVersion{Version: "0.1.0"},
				chef.CookbookVersion{Version: "0.2.0"},
				chef.CookbookVersion{Version: "0.3.0"},
			},
		},
		"bar": chef.CookbookVersions{
			Versions: []chef.CookbookVersion{
				chef.CookbookVersion{Version: "0.1.0"},
			},
		},
	}

	chefAnalyzeClient := subject.ChefAnalyzeClient{
		Cookbooks:         newMockCookbook(cookbookList, nil, nil),
		CookbookArtifacts: newMockCookbookArtifact(chef.CBAGetResponse{}, nil, nil),
		PolicyGroups:      newMockPolicyGroup(chef.PolicyGroupGetResponse{}, nil),
		Policies:          newMockPolicy(chef.RevisionDetailsResponse{}, nil),
		Search:            makeMockSearch(mockedNodesSearchRows(), nil),
	}

	c, err := subject.NewCookbooksReport(
		&chefAnalyzeClient,
		true,
		false,
		Workers,
		"",    // no filter
		false, //anonymize
	)
	assert.Nil(t, err)
	if assert.NotNil(t, c) {
		assert.Equal(t, 4, c.TotalCookbooks)
		c.Generate()
		assert.Equal(t, 4, len(c.Progress)) // verify all progress events arrived
		if assert.Equal(t, 4, len(c.Records)) {
			// we check for only one bar and 3 foo cookbooks
			var (
				foo = 0
				bar = 0
			)

			for _, rec := range c.Records {
				switch rec.Name {
				case "foo":
					foo++
				case "bar":
					bar++
				default:
					t.Fatal("unexpected cookbook name")
				}
				assert.Emptyf(t, rec.Errors(),
					"there should not be any errors for %s-%s", rec.Name, rec.Version)
			}

			assert.Equal(t, 3, foo, "unexpected number of cookbooks foo")
			assert.Equal(t, 1, bar, "unexpected number of cookbooks bar")
		}
	}
}

// Given a valid set of cookbooks in use by nodes,
// verify that the result set is as expected.
func TestCookbooksWithNodeFilter(t *testing.T) {
	savedPath := setupBinstubsDir()
	defer os.Setenv("PATH", savedPath)
	defer os.RemoveAll(createConfigToml(t))

	cookbookList := chef.CookbookListResult{
		"foo": chef.CookbookVersions{
			Versions: []chef.CookbookVersion{
				chef.CookbookVersion{Version: "0.1.0"},
				chef.CookbookVersion{Version: "0.2.0"},
				chef.CookbookVersion{Version: "0.3.0"},
			},
		},
		"bar": chef.CookbookVersions{
			Versions: []chef.CookbookVersion{
				chef.CookbookVersion{Version: "0.1.0"},
			},
		},
	}

	chefAnalyzeClient := subject.ChefAnalyzeClient{
		Cookbooks:         newMockCookbook(cookbookList, nil, nil),
		CookbookArtifacts: newMockCookbookArtifact(chef.CBAGetResponse{}, nil, nil),
		PolicyGroups:      newMockPolicyGroup(chef.PolicyGroupGetResponse{}, nil),
		Policies:          newMockPolicy(chef.RevisionDetailsResponse{}, nil),
		Search:            makeMockSearch(mockedNodesSearchRows(), nil),
	}

	c, err := subject.NewCookbooksReport(
		&chefAnalyzeClient,
		true,
		false,
		Workers,
		"name:blah", // node name filter
		false,       // anonymize
	)
	assert.Nil(t, err)
	if assert.NotNil(t, c) {
		assert.Equal(t, 4, c.TotalCookbooks)
		c.Generate()
		assert.Equal(t, 4, len(c.Progress)) // verify all progress events arrived
		if assert.Equal(t, 4, len(c.Records)) {
			// we check for only one bar and 3 foo cookbooks
			var (
				foo = 0
				bar = 0
			)

			for _, rec := range c.Records {
				switch rec.Name {
				case "foo":
					foo++
				case "bar":
					bar++
				default:
					t.Fatal("unexpected cookbook name")
				}
				assert.Emptyf(t, rec.Errors(),
					"there should not be any errors for %s-%s", rec.Name, rec.Version)
			}

			assert.Equal(t, 3, foo, "unexpected number of cookbooks foo")
			assert.Equal(t, 1, bar, "unexpected number of cookbooks bar")
		}
	}
}

// Given a failure trying to get the list of cookbooks,
// verify the result set is as expected
func TestCookbooks_ListCookbooksErrors(t *testing.T) {

	policyGroupList := chef.PolicyGroupGetResponse{}
	policyDetail := chef.RevisionDetailsResponse{}

	chefAnalyzeClient := subject.ChefAnalyzeClient{
		Cookbooks:         newMockCookbook(chef.CookbookListResult{}, errors.New("i/o timeout"), nil),
		CookbookArtifacts: newMockCookbookArtifact(chef.CBAGetResponse{}, nil, nil),
		PolicyGroups:      newMockPolicyGroup(policyGroupList, nil),
		Policies:          newMockPolicy(policyDetail, nil),
		Search:            makeMockSearch("[]", nil),
	}

	c, err := subject.NewCookbooksReport(
		&chefAnalyzeClient,
		false,
		false,
		Workers,
		"",    // no filter
		false, // anonymize
	)
	if assert.NotNil(t, err) {
		assert.Equal(t, "unable to retrieve cookbooks: i/o timeout", err.Error())
	}
	assert.Nil(t, c)
}

func TestCookbooks_ListAvailableVersionsError(t *testing.T) {
	cookbookList := chef.CookbookListResult{}

	policyGroupList := chef.PolicyGroupGetResponse{}
	policyDetail := chef.RevisionDetailsResponse{}

	chefAnalyzeClient := subject.ChefAnalyzeClient{
		Cookbooks:         newMockCookbook(cookbookList, errors.New("list error"), nil),
		CookbookArtifacts: newMockCookbookArtifact(chef.CBAGetResponse{}, nil, nil),
		PolicyGroups:      newMockPolicyGroup(policyGroupList, nil),
		Policies:          newMockPolicy(policyDetail, nil),
	}

	c, err := subject.NewCookbooksReport(
		&chefAnalyzeClient,
		false,
		false,
		Workers,
		"",    // no filter
		false, // anonymize
	)
	assert.Nil(t, c)
	assert.EqualError(t, err, "unable to retrieve cookbooks: list error")
}

func TestCookbooks_NoneAvailable(t *testing.T) {

	chefAnalyzeClient := subject.ChefAnalyzeClient{
		Cookbooks:         newMockCookbook(chef.CookbookListResult{}, nil, nil),
		CookbookArtifacts: newMockCookbookArtifact(chef.CBAGetResponse{}, nil, nil),
		PolicyGroups:      newMockPolicyGroup(chef.PolicyGroupGetResponse{}, nil),
		Policies:          newMockPolicy(chef.RevisionDetailsResponse{}, nil),
		Search:            makeMockSearch(mockedEmptyNodesSearchRows(), nil),
	}

	c, err := subject.NewCookbooksReport(
		&chefAnalyzeClient,
		false,
		false,
		Workers,
		"",    // no filter
		false, // anonymize
	)
	assert.Nil(t, err)
	if assert.NotNil(t, c) {
		assert.Equal(t, 0, c.TotalCookbooks)
	}
}

// Given a failure in downloading a cookbook, verify
// the result set is as expected
func TestCookbooks_DownloadErrors(t *testing.T) {
	savedPath := setupBinstubsDir()
	defer os.Setenv("PATH", savedPath)
	defer os.RemoveAll(createConfigToml(t))

	cookbookList := chef.CookbookListResult{
		"foo": chef.CookbookVersions{
			Versions: []chef.CookbookVersion{
				chef.CookbookVersion{Version: "0.1.0"},
			},
		},
	}

	chefAnalyzeClient := subject.ChefAnalyzeClient{
		Cookbooks:         newMockCookbook(cookbookList, nil, errors.New("download error")),
		CookbookArtifacts: newMockCookbookArtifact(chef.CBAGetResponse{}, nil, nil),
		PolicyGroups:      newMockPolicyGroup(chef.PolicyGroupGetResponse{}, nil),
		Policies:          newMockPolicy(chef.RevisionDetailsResponse{}, nil),
		Search:            makeMockSearch(mockedNodesSearchRows(), nil),
	}

	c, err := subject.NewCookbooksReport(
		&chefAnalyzeClient,
		true, // run cookstyle, which means that we will download the cookbooks
		false,
		Workers,
		"",    // no filter
		false, // anonymize
	)
	assert.Nil(t, err)
	if assert.NotNil(t, c) {
		assert.Equal(t, 1, c.TotalCookbooks)
		c.Generate()
		assert.Equal(t, 1, len(c.Progress)) // Verify that all progress events got queued
		if assert.Equal(t, 1, len(c.Records)) {
			for _, rec := range c.Records {
				assert.EqualError(t, rec.DownloadError, "unable to download cookbook foo: download error")
				assert.NoError(t, rec.UsageLookupError, "unexpected UsageLookupError")
				assert.NoError(t, rec.CookstyleError, "unexpected UsageLookupError")
			}
		}
	}
}

// Given a failure in fetching node usage for a cookbook,
// the result set is as expected
func TestCookbooks_UsageStatErrors(t *testing.T) {
	savedPath := setupBinstubsDir()
	defer os.Setenv("PATH", savedPath)

	cookbookList := chef.CookbookListResult{
		"foo": chef.CookbookVersions{
			Versions: []chef.CookbookVersion{
				chef.CookbookVersion{Version: "0.1.0"},
			},
		},
	}

	chefAnalyzeClient := subject.ChefAnalyzeClient{
		Cookbooks:         newMockCookbook(cookbookList, nil, nil),
		CookbookArtifacts: newMockCookbookArtifact(chef.CBAGetResponse{}, nil, nil),
		PolicyGroups:      newMockPolicyGroup(chef.PolicyGroupGetResponse{}, nil),
		Policies:          newMockPolicy(chef.RevisionDetailsResponse{}, nil),
		Search:            makeMockSearch(mockedNodesSearchRows(), errors.New("lookup error")),
	}

	c, err := subject.NewCookbooksReport(
		&chefAnalyzeClient,
		false, // do not download or cookstyle
		// Because the usage error makes it look like no nodes are attached to this cookbook
		// we need to return only unused cookbooks
		true,
		Workers,
		"",    // no filter
		false, // anonymize
	)
	assert.Nil(t, err)
	if assert.NotNil(t, c) {
		assert.Equal(t, 1, c.TotalCookbooks)
		c.Generate()
		assert.Equal(t, 1, len(c.Progress)) // Verify that all progress events got queued
		if assert.Equal(t, 1, len(c.Records)) {
			for _, rec := range c.Records {
				assert.NoError(t, rec.DownloadError, "unexpected DownloadError")
				assert.EqualError(t, rec.UsageLookupError, "unable to get cookbook usage information: lookup error")
				assert.NoError(t, rec.CookstyleError, "unexpected UsageLookupError")
			}
		}
	}
}

func TestCookbookRecord_Errors(t *testing.T) {
	var cr subject.CookbookRecord
	cr.DownloadError = errors.New("Download error")
	cr.UsageLookupError = errors.New("usage lookup error")
	cr.CookstyleError = errors.New("cookstyle error")
	errors := cr.Errors()
	assert.Equal(t, 3, len(errors))
	assert.Contains(t, errors, cr.DownloadError)
	assert.Contains(t, errors, cr.UsageLookupError)
	assert.Contains(t, errors, cr.CookstyleError)
}

func TestCookbookRecord_ErrorsWhenEmpty(t *testing.T) {
	var cr subject.CookbookRecord
	errors := cr.Errors()
	assert.Equal(t, len(errors), 0)
}

func TestCookbooks_withCookstyleError(t *testing.T) {
	defer os.RemoveAll(createConfigToml(t))
	cookbookList := chef.CookbookListResult{
		"foo": chef.CookbookVersions{
			Versions: []chef.CookbookVersion{
				chef.CookbookVersion{Version: "0.1.0"},
				chef.CookbookVersion{Version: "0.2.0"},
				chef.CookbookVersion{Version: "0.3.0"},
			},
		},
		"bar": chef.CookbookVersions{
			Versions: []chef.CookbookVersion{
				chef.CookbookVersion{Version: "0.1.0"},
			},
		},
	}

	chefAnalyzeClient := subject.ChefAnalyzeClient{
		Cookbooks:         newMockCookbook(cookbookList, nil, nil),
		CookbookArtifacts: newMockCookbookArtifact(chef.CBAGetResponse{}, nil, nil),
		PolicyGroups:      newMockPolicyGroup(chef.PolicyGroupGetResponse{}, nil),
		Policies:          newMockPolicy(chef.RevisionDetailsResponse{}, nil),
		Search:            makeMockSearch(mockedNodesSearchRows(), nil),
	}

	c, err := subject.NewCookbooksReport(
		&chefAnalyzeClient,
		true,
		false,
		Workers,
		"",    // no filter
		false, // anonymize
	)
	assert.Nil(t, err)
	if assert.NotNil(t, c) {
		assert.Equal(t, 4, c.TotalCookbooks)
		c.Generate()
		assert.Equal(t, 4, len(c.Progress)) // Verify that all progress events got queued
		if assert.Equal(t, 4, len(c.Records)) {
			for _, rec := range c.Records {
				allErrors := rec.Errors()
				if assert.Equal(t, 1, len(allErrors)) {
					assert.Contains(t, allErrors[0].Error(), "\"cookstyle\": executable file not found")
				}
			}
		}
	}
}

func TestCookbooks_withHeavyLoad(t *testing.T) {
	savedPath := setupBinstubsDir()
	defer os.Setenv("PATH", savedPath)
	defer os.RemoveAll(createConfigToml(t))

	TotalCookbooks := 500
	cookbookList := chef.CookbookListResult{}
	for i := 0; i < TotalCookbooks; i++ {
		cookbookList[fmt.Sprintf("foo%d", i)] = chef.CookbookVersions{
			Versions: []chef.CookbookVersion{
				chef.CookbookVersion{Version: "0.1.0"},
				chef.CookbookVersion{Version: "0.2.0"},
				chef.CookbookVersion{Version: "0.3.0"},
			},
		}
	}

	assert.Equal(t, TotalCookbooks, len(cookbookList))

	chefAnalyzeClient := subject.ChefAnalyzeClient{
		Cookbooks:         newMockCookbook(cookbookList, nil, nil),
		CookbookArtifacts: newMockCookbookArtifact(chef.CBAGetResponse{}, nil, nil),
		PolicyGroups:      newMockPolicyGroup(chef.PolicyGroupGetResponse{}, nil),
		Policies:          newMockPolicy(chef.RevisionDetailsResponse{}, nil),
		Search:            makeMockSearch(mockedNodesSearchRows(), nil),
	}

	c, err := subject.NewCookbooksReport(
		&chefAnalyzeClient,
		true,
		false,
		Workers,
		"",    // no filter
		false, // anonymize
	)
	assert.Nil(t, err)
	if assert.NotNil(t, c) {
		// we should have a total of three times the number of TotalCookbooks
		// since every cookbook contains three versions
		expectedCookbookCount := TotalCookbooks * 3
		assert.Equal(t, expectedCookbookCount, c.TotalCookbooks)
		c.Generate()
		assert.Equal(t, expectedCookbookCount, len(c.Progress)) // Verify that all progress events got queued
		if assert.Equal(t, expectedCookbookCount, len(c.Records)) {
			for _, rec := range c.Records {
				assert.Emptyf(t, rec.Errors(),
					"there should not be any errors for %s-%s", rec.Name, rec.Version)
			}
		}
	}
}

func TestCookbookRecordsByNameVersion_Len(t *testing.T) {
	crs := []*subject.CookbookRecord{
		&subject.CookbookRecord{Name: "cookbook_name"},
	}
	assert.Equal(t, 1, subject.CookbookRecordsBySortOrder.Len(crs))
}

func TestCookbookRecordsByNameVersion_Swap(t *testing.T) {
	crs := []*subject.CookbookRecord{
		&subject.CookbookRecord{Name: "cb1"},
		&subject.CookbookRecord{Name: "cb2"},
	}
	expected := []*subject.CookbookRecord{
		&subject.CookbookRecord{Name: "cb2"},
		&subject.CookbookRecord{Name: "cb1"},
	}
	subject.CookbookRecordsBySortOrder.Swap(crs, 0, 1)
	assert.Equal(t, expected, crs)
}

func TestCookbookRecordsByNameVersion_Less(t *testing.T) {
	crs := []*subject.CookbookRecord{
		&subject.CookbookRecord{Name: "cb1", Version: "1.0.1"},
		&subject.CookbookRecord{Name: "cb2", Version: "1.0.1"},
		&subject.CookbookRecord{Name: "cb2", Version: "1.0.0"},
		&subject.CookbookRecord{Name: "cb1", Version: "1.0.0"},
		&subject.CookbookRecord{Name: "CB1", Version: "1.0.0"},
		&subject.CookbookRecord{Name: "CB2", Version: "1.0.0"},
	}
	assert.True(t, subject.CookbookRecordsBySortOrder.Less(crs, 0, 1))
	assert.True(t, subject.CookbookRecordsBySortOrder.Less(crs, 0, 2))
	assert.False(t, subject.CookbookRecordsBySortOrder.Less(crs, 0, 3))
	assert.False(t, subject.CookbookRecordsBySortOrder.Less(crs, 1, 0))
	assert.False(t, subject.CookbookRecordsBySortOrder.Less(crs, 1, 2))
	assert.False(t, subject.CookbookRecordsBySortOrder.Less(crs, 1, 3))
	assert.False(t, subject.CookbookRecordsBySortOrder.Less(crs, 2, 0))
	assert.True(t, subject.CookbookRecordsBySortOrder.Less(crs, 2, 1))
	assert.False(t, subject.CookbookRecordsBySortOrder.Less(crs, 2, 3))
	assert.True(t, subject.CookbookRecordsBySortOrder.Less(crs, 3, 0))
	assert.True(t, subject.CookbookRecordsBySortOrder.Less(crs, 3, 1))
	assert.True(t, subject.CookbookRecordsBySortOrder.Less(crs, 3, 2))
	assert.False(t, subject.CookbookRecordsBySortOrder.Less(crs, 4, 3))
	assert.False(t, subject.CookbookRecordsBySortOrder.Less(crs, 5, 2))
}
