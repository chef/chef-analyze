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

// https://www.davidkaya.com/sets-in-golang/
type versionSet map[string]struct{}

func newVersionSet(version string) *versionSet {
	return &versionSet{
		version: struct{}{},
	}
}

func (s versionSet) Add(value string) {
	_, exists := s[value]
	if !exists {
		s[value] = struct{}{}
	}
}

func (s versionSet) Array() []string {
	keys := make([]string, len(s), len(s))

	i := 0
	for k := range s {
		keys[i] = k
		i++
	}
	return keys
}

// func (s *versionSet) Remove(value string) {
// 	delete(s.m, value)
// }

// func (s *versionSet) Contains(value string) bool {
// 	_, c := s.m[value]
// 	return c
// }

func Cookbooks(cfg *Reporting, searcher PartialSearchInterface) error {
	var (
		query = map[string]interface{}{
			"cookbooks": []string{"cookbooks"},
		}
	)

	results, err := searcher.PartialExec("node", "*:*", query)
	if err != nil {
		return errors.Wrap(err, "unable to get node(s) information")
	}

	// We use len here and not pres.Rows, because when caller does not have permissions to
	// 	view all nodes in the result set, the actual returned number will be lower than
	// 	the value of Rows.

	formatString := "%30s   %-12v\n"
	fmt.Printf(formatString, "Cookbook Name", "Version(s)")
	uniqueCookbooks := make(map[string]*versionSet, 0)

	for _, element := range results.Rows {

		// cookbook version arrives as [ NAME : { version: VERSION } - we extract that here.
		v := element.(map[string]interface{})["data"].(map[string]interface{})

		if v != nil {
			if v["cookbooks"] != nil {

				// First, de-dup used cookbooks
				cookbooks := v["cookbooks"].(map[string]interface{})
				for k, v := range cookbooks {
					name := k
					version := safeStringFromMap(v.(map[string]interface{}), "version")
					if versionSet, ok := uniqueCookbooks[name]; ok {
						versionSet.Add(version)
					} else {
						uniqueCookbooks[name] = newVersionSet(version)
					}
				}
			}
		} else {
			return errors.New("No cookbooks found")
		}
	}

	for name, versions := range uniqueCookbooks {
		fmt.Printf(formatString, name, strings.Join(versions.Array(), ", "))
	}

	return nil

	// chefClient, err := NewChefClient(cfg)
	// if err != nil {
	// 	return err
	// }

	// cbooksList, err := chefClient.Cookbooks.ListAvailableVersions("all")
	// if err != nil {
	// 	return errors.Wrap(err, "unable to get cookbook(s) information")
	// }

	// formatString := "%30s   %-12v\n"
	// fmt.Printf(formatString, "Cookbook Name", "Version(s)")
	// for cookbook, cbookVersions := range cbooksList {
	// 	versionsArray := make([]string, len(cbookVersions.Versions))
	// 	for i, details := range cbookVersions.Versions {
	// 		versionsArray[i] = details.Version
	// 	}
	// 	fmt.Printf(formatString, cookbook, strings.Join(versionsArray[:], ", "))
	// }

	// return nil
}
