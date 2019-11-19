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
	"log"
	"os"

	chef "github.com/chef/go-chef"
	"github.com/cheggaaa/pb/v3"
	"github.com/pkg/errors"
)

type CookbookStateRecord struct {
	Name             string
	Version          string
	Violations       int
	Autocorrectable  int
	Nodes            []string
	path             string
	CBDownloadError  error
	UsageLookupError error
}

func CookbookState(cfg *Reporting, cbi CookbookInterface, searcher PartialSearchInterface, includeUnboundCookbooks bool) ([]*CookbookStateRecord, error) {
	fmt.Println("Finding available cookbooks...") // c <- ProgressUpdate(Event: COOKBOOK_FETCH)
	// Version limit of "0" means fetch all
	results, err := cbi.ListAvailableVersions("0")
	if includeUnboundCookbooks == false {
		boundCookbooks := make(chef.CookbookListResult, len(results))
		for cookbookName, cookbookVersions := range results {
			for _, cookbookVersion := range cookbookVersions.Versions {
				// filter cookbooks down to only ones that are assigned to a cookbook
				var (
					query = map[string]interface{}{
						"nodes": []string{"nodes"},
					}
				)

				pres, err := searcher.PartialExec("node", fmt.Sprintf("cookbooks_%s_version:%s", cookbookName, cookbookVersion.Version), query)
				if err != nil {
					return nil, errors.Wrap(err, "unable to get cookbook usage information")
				}
				// If there are no nodes with this cookbook, remove that version
				if l := len(pres.Rows); l > 0 {
					fmt.Printf("%d nodes found for %s(%s)\n", l, cookbookName, cookbookVersion.Version)
					if cb, found := boundCookbooks[cookbookName]; found {
						cb.Versions = append(boundCookbooks[cookbookName].Versions, chef.CookbookVersion{Url: cookbookVersion.Url, Version: cookbookVersion.Version})
					} else {
						boundCookbooks[cookbookName] = chef.CookbookVersions{Url: cookbookVersions.Url, Versions: []chef.CookbookVersion{chef.CookbookVersion{Url: cookbookVersion.Url, Version: cookbookVersion.Version}}}
					}
				} else {
					fmt.Printf("No nodes found for %s(%s)\n", cookbookName, cookbookVersion.Version)
				}
			}
		}
		results = boundCookbooks
	}

	if err != nil {
		return nil, errors.Wrap(err, "Unable to retrieve cookbooks")
	}

	// First pass: get totals so we can accurately report progress
	//             and allocate results
	numVersions := 0
	for _, versions := range results {
		for range versions.Versions {
			numVersions++
		}
	}

	fmt.Println("Downloading cookbooks...")
	cbStates, err := downloadCookbooks(cbi, searcher, results, numVersions)
	if err != nil {
		return cbStates, err
	}

	fmt.Println("Analyzing cookbooks...")
	formatterPath, err := createTempFileWithContent(correctableCountCookstyleFormatterRb)
	if err != nil {
		return cbStates, err
	}
	defer os.Remove(formatterPath)

	runCookstyle(cbStates, formatterPath)

	return cbStates, nil
}

func downloadCookbooks(cbi CookbookInterface, searcher SearchInterface, cookbooks chef.CookbookListResult, numVersions int) ([]*CookbookStateRecord, error) {
	var cbState *CookbookStateRecord
	// Ultimately, progress tracking and output lives with caller. While we work out
	// those details, it lives here.
	progress := pb.StartNew(numVersions)
	cbStates := make([]*CookbookStateRecord, numVersions, numVersions)
	var index int
	// Second pass: let's pack them up into returnable form.
	for cookbookName, cookbookVersions := range cookbooks {
		for _, ver := range cookbookVersions.Versions {
			cbState = &CookbookStateRecord{Name: cookbookName,
				path:    fmt.Sprintf(".analyze-cache/cookbooks/%v-%v", cookbookName, ver.Version),
				Version: ver.Version,
			}

			nodes, err := nodesUsingCookbookVersion(searcher, cookbookName, ver.Version)
			if err != nil {
				cbState.UsageLookupError = err
			} else {
				cbState.Nodes = nodes
			}

			err = cbi.DownloadTo(cookbookName, ver.Version, ".analyze-cache/cookbooks")
			if err != nil {
				cbState.CBDownloadError = err
			}
			cbStates[index] = cbState
			index++

			progress.Increment()
		}
	}
	progress.Finish()
	return cbStates, nil
}

func runCookstyle(cbStates []*CookbookStateRecord, formatterPath string) {
	progress := pb.StartNew(len(cbStates))
	runner := ExecCommandRunner{}
	for _, cb := range cbStates {
		progress.Increment()
		// If we could not download the complete cookbook, we can't provide
		// an accurate set of results
		if cb.CBDownloadError != nil {
			continue
		}
		cookstyleResults, err := RunCookstyle(cb.path, runner)
		if err != nil {
			// TODO - unlikely, but if it can actually happen we should capture this
			//        potential failure in results too, eg cb.ViolationCheckError = err
			log.Print(err)
			continue
		}
		for _, file := range cookstyleResults.Files {
			for _, offense := range file.Offenses {

				cb.Violations += 1
				if offense.Correctable {
					cb.Autocorrectable += 1
				}
			}
		}
	}
	progress.Finish()
}
func nodesUsingCookbookVersion(searcher SearchInterface, cookbook string, version string) ([]string, error) {
	var (
		query = map[string]interface{}{
			"name": []string{"name"},
		}
	)

	pres, err := searcher.PartialExec("node", fmt.Sprintf("cookbooks_%s_version:%s", cookbook, version), query)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get cookbook usage information")
	}
	results := make([]string, 0, len(pres.Rows))
	for _, element := range pres.Rows {
		v := element.(map[string]interface{})["data"].(map[string]interface{})
		if v != nil {
			results = append(results, safeStringFromMap(v, "name"))
		}
	}

	return results, nil
}
