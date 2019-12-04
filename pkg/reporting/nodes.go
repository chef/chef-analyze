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

type CookbookVersion struct {
	Name    string
	Version string
}

func (cbv *CookbookVersion) String() string {
	return fmt.Sprintf("%s(%s)", cbv.Name, cbv.Version)
}

type NodeReportItem struct {
	Name             string
	ChefVersion      string
	OS               string
	OSVersion        string
	CookbookVersions []CookbookVersion
}

func (nri *NodeReportItem) Array() []string {
	var cookbooks []string
	for _, v := range nri.CookbookVersions {
		cookbooks = append(cookbooks, v.String())
	}
	var chefVersion string
	if nri.ChefVersion == "" {
		chefVersion = "-"
	} else {
		chefVersion = nri.ChefVersion
	}
	// This data seems to be all or none - you'll have both OS/Version fields,
	// or neither.
	var osInfo string
	if nri.OS == "" {
		osInfo = "-"
	} else {
		osInfo = fmt.Sprintf("%s v%s", nri.OS, nri.OSVersion)
	}

	var cbInfo string
	if len(cookbooks) == 0 {
		cbInfo = "-"
	} else {
		cbInfo = strings.Join(cookbooks, " ")
	}

	return []string{nri.Name, chefVersion, osInfo, cbInfo}
}

// NOTE - we no longer need cfg. I'm not sure that this is best - I like having a single
//        cfg which includes the client, but did not want to create a full mock interface for
//        chef.client here - that belongs in go-chef, where it can be maintained alongside
//        any interface changes that originate there.
func Nodes(cfg *Reporting, searcher SearchInterface) ([]NodeReportItem, error) {
	var (
		query = map[string]interface{}{
			"name":         []string{"name"},
			"chef_version": []string{"chef_packages", "chef", "version"},
			"os":           []string{"platform"},
			"os_version":   []string{"platform_version"},
			"cookbooks":    []string{"cookbooks"},
		}
	)

	pres, err := searcher.PartialExec("node", "*:*", query)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get node(s) information")
	}

	// We use len here and not pres.Rows, because when caller does not have permissions to
	// 	view all nodes in the result set, the actual returned number will be lower than
	// 	the value of Rows.

	results := make([]NodeReportItem, 0, len(pres.Rows))
	for _, element := range pres.Rows {

		// cookbook version arrives as [ NAME : { version: VERSION } - we extract that here.
		v := element.(map[string]interface{})["data"].(map[string]interface{})

		if v != nil {
			item := NodeReportItem{
				Name:        safeStringFromMap(v, "name"),
				OS:          safeStringFromMap(v, "os"),
				OSVersion:   safeStringFromMap(v, "os_version"),
				ChefVersion: safeStringFromMap(v, "chef_version"),
			}

			if v["cookbooks"] != nil {
				cookbooks := v["cookbooks"].(map[string]interface{})
				item.CookbookVersions = make([]CookbookVersion, 0, len(cookbooks))
				for k, v := range cookbooks {
					cbv := CookbookVersion{Name: k, Version: safeStringFromMap(v.(map[string]interface{}), "version")}
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
