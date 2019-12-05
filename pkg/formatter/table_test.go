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

	subject "github.com/chef/chef-analyze/pkg/formatter"
	"github.com/chef/chef-analyze/pkg/reporting"
	"github.com/stretchr/testify/assert"
)

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
