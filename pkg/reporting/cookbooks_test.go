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
	"errors"
	"fmt"
	"os"
	"testing"

	chef "github.com/chef/go-chef"
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
	c, err := subject.StartPipeline(
		newMockCookbook(chef.CookbookListResult{}, nil, nil),
		makeMockSearch("[]", nil),
		false,
		false,
		Workers,
	)
	assert.Nil(t, err)
	if assert.NotNil(t, c) {
		assert.Empty(t, c.Records)
		assert.Equal(t, 0, c.TotalCookbooks)
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
	c, err := subject.StartPipeline(
		newMockCookbook(cookbookList, nil, nil),
		makeMockSearch(mockedNodesSearchRows(), nil), // nodes are found
		false,
		true, // display only unused cookbooks
		Workers,
	)
	assert.Nil(t, err)
	if assert.NotNil(t, c) {
		assert.Empty(t, c.Records)
		assert.Equal(t, 4, c.TotalCookbooks)
	}

	c, err = subject.StartPipeline(
		newMockCookbook(cookbookList, nil, nil),
		makeMockSearch(mockedNodesSearchRows(), nil), // nodes are found
		false,
		false, // display only used cookbooks
		Workers,
	)
	assert.Nil(t, err)
	if assert.NotNil(t, c) {
		assert.Equal(t, 4, c.TotalCookbooks)
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

				assert.NotEmpty(t, rec.Nodes, "nodes should be found")
			}

			assert.Equal(t, 3, foo, "unexpected number of cookbooks foo")
			assert.Equal(t, 1, bar, "unexpected number of cookbooks bar")
		}
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
	c, err := subject.StartPipeline(
		newMockCookbook(cookbookList, nil, nil),
		makeMockSearch("[]", nil), // no nodes are returned
		false,
		true, // display only unused cookbooks
		Workers,
	)
	assert.Nil(t, err)
	if assert.NotNil(t, c) {
		assert.Equal(t, 4, c.TotalCookbooks)
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

				assert.Empty(t, rec.Nodes, "no nodes should be found")
			}

			assert.Equal(t, 3, foo, "unexpected number of cookbooks foo")
			assert.Equal(t, 1, bar, "unexpected number of cookbooks bar")
		}
	}

	c, err = subject.StartPipeline(
		newMockCookbook(cookbookList, nil, nil),
		makeMockSearch("[]", nil), // no nodes are returned
		false,
		false, // display only used cookbooks
		Workers,
	)
	assert.Nil(t, err)
	if assert.NotNil(t, c) {
		assert.Empty(t, c.Records)
		assert.Equal(t, 4, c.TotalCookbooks)
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
	c, err := subject.StartPipeline(
		newMockCookbook(cookbookList, nil, nil),
		makeMockSearch(mockedNodesSearchRows(), nil),
		true,
		false,
		Workers,
	)
	assert.Nil(t, err)
	if assert.NotNil(t, c) {
		assert.Equal(t, 4, c.TotalCookbooks)
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
	c, err := subject.StartPipeline(
		newMockCookbook(chef.CookbookListResult{}, errors.New("i/o timeout"), nil),
		makeMockSearch("[]", nil),
		false,
		false,
		Workers,
	)
	if assert.NotNil(t, err) {
		assert.Equal(t, "unable to retrieve cookbooks: i/o timeout", err.Error())
	}
	assert.Nil(t, c)
}

func TestCookbooks_ListAvailableVersionsError(t *testing.T) {
	cookbookList := chef.CookbookListResult{}
	c, err := subject.StartPipeline(
		newMockCookbook(cookbookList, errors.New("list error"), nil),
		nil,
		false,
		false,
		Workers,
	)
	assert.Nil(t, c)
	assert.EqualError(t, err, "unable to retrieve cookbooks: list error")
}

func TestCookbooks_NoneAvailable(t *testing.T) {

	c, err := subject.StartPipeline(
		newMockCookbook(chef.CookbookListResult{}, nil, nil),
		makeMockSearch(mockedEmptyNodesSearchRows(), nil),
		false,
		false,
		Workers,
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
	c, err := subject.StartPipeline(
		newMockCookbook(cookbookList, nil, errors.New("download error")),
		makeMockSearch(mockedNodesSearchRows(), nil),
		true, // run cookstyle, which means that we will download the cookbooks
		false,
		Workers,
	)
	assert.Nil(t, err)
	if assert.NotNil(t, c) {
		assert.Equal(t, 1, c.TotalCookbooks)
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
	c, err := subject.StartPipeline(
		newMockCookbook(cookbookList, nil, nil),
		makeMockSearch(mockedNodesSearchRows(), errors.New("lookup error")),
		false,
		// Because the usage error makes it look like no nodes are attached to this cookbook
		// we need to return only unused cookbooks
		true,
		Workers,
	)
	assert.Nil(t, err)
	if assert.NotNil(t, c) {
		assert.Equal(t, 1, c.TotalCookbooks)
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
	c, err := subject.NewCookbooks(
		newMockCookbook(cookbookList, nil, nil),
		makeMockSearch(mockedNodesSearchRows(), nil),
		true,
		false,
		Workers,
	)
	assert.Nil(t, err)
	if assert.NotNil(t, c) {
		assert.Equal(t, 4, c.TotalCookbooks)
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

	totalCookbooks := 500
	cookbookList := chef.CookbookListResult{}
	for i := 0; i < totalCookbooks; i++ {
		cookbookList[fmt.Sprintf("foo%d", i)] = chef.CookbookVersions{
			Versions: []chef.CookbookVersion{
				chef.CookbookVersion{Version: "0.1.0"},
				chef.CookbookVersion{Version: "0.2.0"},
				chef.CookbookVersion{Version: "0.3.0"},
			},
		}
	}

	assert.Equal(t, totalCookbooks, len(cookbookList))

	c, err := subject.NewCookbooks(
		newMockCookbook(cookbookList, nil, nil),
		makeMockSearch(mockedNodesSearchRows(), nil),
		true,
		false,
		Workers,
	)
	assert.Nil(t, err)
	if assert.NotNil(t, c) {
		// we should have a total of three times the number of totalCookbooks
		// since every cookbook contains three versions
		assert.Equal(t, totalCookbooks*3, c.TotalCookbooks)
		if assert.Equal(t, totalCookbooks*3, len(c.Records)) {
			for _, rec := range c.Records {
				assert.Emptyf(t, rec.Errors(),
					"there should not be any errors for %s-%s", rec.Name, rec.Version)
			}
		}
	}
}
