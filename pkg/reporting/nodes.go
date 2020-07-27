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

package reporting

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// CookbookVersion name and version (sortable by name)
type CookbookVersion struct {
	Name    string
	Version string
}

// Sort interface the sorts cookbooks by name and version.
type CookbookByNameVersion []CookbookVersion

func (cnv CookbookByNameVersion) Len() int {
	return len(cnv)
}
func (cnv CookbookByNameVersion) Swap(i, j int) {
	cnv[i], cnv[j] = cnv[j], cnv[i]
}
func (cnv CookbookByNameVersion) Less(i, j int) bool {
	l1, l2 := strings.ToLower(cnv[i].Name), strings.ToLower(cnv[j].Name)
	if l1 != l2 {
		return l1 < l2
	}
	return strings.ToLower(cnv[i].Version) < strings.ToLower(cnv[j].Version)
}

func (cbv *CookbookVersion) String() string {
	return fmt.Sprintf("%s(%s)", cbv.Name, cbv.Version)
}

// Sort interface the sorts node records by name.
type NodeRecordsByName []*NodeReportItem

func (nrs NodeRecordsByName) Len() int {
	return len(nrs)
}
func (nrs NodeRecordsByName) Swap(i, j int) {
	nrs[i], nrs[j] = nrs[j], nrs[i]
}
func (nrs NodeRecordsByName) Less(i, j int) bool {
	return strings.ToLower(nrs[i].Name) < strings.ToLower(nrs[j].Name)
}

// NodeReportItem data object for nodes
type NodeReportItem struct {
	Name             string
	ChefVersion      string
	OS               string
	OSVersion        string
	CookbookVersions []CookbookVersion
	PolicyGroup      string
	Policy           string
	PolicyRev        string
	Anonymize        bool
}

// OSVersionPretty looks nice
func (nri *NodeReportItem) OSVersionPretty() string {
	// this data seems to be all or none,
	// you'll have both OS/Version fields, or neither.
	if nri.OS != "" {
		return fmt.Sprintf("%s v%s", nri.OS, nri.OSVersion)
	}

	return ""
}

// GetPolicyGroup returns Policy Group or "no group"
func (nri *NodeReportItem) GetPolicyGroup() string {

	if nri.PolicyGroup == "" {
		return "no group"
	}

	return nri.PolicyGroup
}

// HasPolicyGroup returns true or false for Policy Group presenece
func (nri *NodeReportItem) HasPolicyGroup() bool {

	if nri.PolicyGroup == "" {
		return false
	}

	return true
}

// GetPolicy returns Policy Name or "no policy"
func (nri *NodeReportItem) GetPolicy() string {

	if nri.Policy == "" {
		return "no policy"
	}

	return nri.Policy
}

// GetPolicyWithRev returns formatted Policy name with revision
func (nri *NodeReportItem) GetPolicyWithRev() string {

	if nri.Policy == "" {
		return "no policy"
	}

	return fmt.Sprintf("%s (rev %s)", nri.Policy, nri.PolicyRev)
}

// CookbooksList transforms to an easily printable []string
func (nri *NodeReportItem) CookbooksList() []string {
	var cookbooks = make([]string, 0, len(nri.CookbookVersions))

	for _, v := range nri.CookbookVersions {
		cookbooks = append(cookbooks, v.String())
	}

	return cookbooks
}

// GenerateNodesReport generate a nodes report
func GenerateNodesReport(client *ChefAnalyzeClient, filter string, anonymize bool) ([]*NodeReportItem, error) {
	var (
		query = map[string]interface{}{
			"name":            []string{"name"},
			"chef_version":    []string{"chef_packages", "chef", "version"},
			"os":              []string{"platform"},
			"os_version":      []string{"platform_version"},
			"cookbooks":       []string{"cookbooks"},
			"policy_group":    []string{"policy_group"},
			"policy_name":     []string{"policy_name"},
			"policy_revision": []string{"policy_revision"},
		}
	)
	if filter == "" {
		filter = "*:*"
	}
	pres, err := client.Search.PartialExec("node", filter, query)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get node(s) information")
	}

	// We use len here and not pres.Rows, because when caller does not have permissions to
	// 	view all nodes in the result set, the actual returned number will be lower than
	// 	the value of Rows.

	results := make([]*NodeReportItem, 0, len(pres.Rows))
	for _, element := range pres.Rows {

		// cookbook version arrives as [ NAME : { version: VERSION } - we extract that here.
		v := element.(map[string]interface{})["data"].(map[string]interface{})

		if v != nil {
			item := &NodeReportItem{
				Name:        safeStringFromMap(v, "name"),
				OS:          safeStringFromMap(v, "os"),
				OSVersion:   safeStringFromMap(v, "os_version"),
				ChefVersion: safeStringFromMap(v, "chef_version"),
				PolicyGroup: safeStringFromMap(v, "policy_group"),
				Policy:      safeStringFromMap(v, "policy_name"),
				PolicyRev:   safeStringFromMap(v, "policy_revision"),
			}

			if anonymize {
				item.Anonymize = true
				item.Name = hashString(item.Name)
				item.PolicyGroup = hashString(item.PolicyGroup)
				item.Policy = hashString(item.Policy)
			}

			if v["cookbooks"] != nil {
				cookbooks := v["cookbooks"].(map[string]interface{})
				item.CookbookVersions = make([]CookbookVersion, 0, len(cookbooks))
				for k, v := range cookbooks {
					cbv := CookbookVersion{Name: k, Version: safeStringFromMap(v.(map[string]interface{}), "version")}
					if anonymize {
						cbv.Name = hashString(k)
					} else {
						cbv.Name = k
					}
					item.CookbookVersions = append(item.CookbookVersions, cbv)
				}
			}
			results = append(results, item)
		}
	}
	return results, nil
}

// This returns the value referenced by `key` in `values`. If value is nil,
// it returns an empty string; otherwise it returns the original string.
// Will probably be more generally useful exposed as public in the reporting package.
func safeStringFromMap(values map[string]interface{}, key string) string {
	if values[key] == nil {
		return ""
	}
	return values[key].(string)
}
