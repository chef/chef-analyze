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
	"os"

	chef "github.com/chef/go-chef"
	"github.com/cheggaaa/pb/v3"
	"github.com/pkg/errors"
)

type CookbookStateRecord struct {
	Name             string
	Version          string
	Files            []CookbookFile
	Nodes            []string
	path             string
	DownloadError    error
	UsageLookupError error
	CookstyleError   error
}

func (r CookbookStateRecord) NumOffenses() int {
	i := 0
	for _, f := range r.Files {
		i += len(f.Offenses)
	}
	return i
}

func (r CookbookStateRecord) NumCorrectable() int {
	i := 0
	for _, f := range r.Files {
		for _, o := range f.Offenses {
			if o.Correctable {
				i++
			}
		}
	}
	return i
}

func CookbookState(cfg *Reporting, cbi CookbookInterface, searcher SearchInterface, includeUnboundCookbooks bool) ([]*CookbookStateRecord, error) {
	fmt.Println("Finding available cookbooks...") // c <- ProgressUpdate(Event: COOKBOOK_FETCH)
	// Version limit of "0" means fetch all
	results, err := cbi.ListAvailableVersions("0")
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
	cbStates, err := downloadCookbooks(cbi, searcher, results, numVersions, includeUnboundCookbooks)
	if err != nil {
		return cbStates, err
	}

	fmt.Println("Analyzing cookbooks...")
	runCookstyle(cbStates)

	return cbStates, nil
}

func downloadCookbooks(cbi CookbookInterface, searcher SearchInterface, cookbooks chef.CookbookListResult, numVersions int, includeUnboundCookbooks bool) ([]*CookbookStateRecord, error) {
	var (
		query = map[string]interface{}{
			"name": []string{"name"},
		}
		cbState *CookbookStateRecord
	)

	// Ultimately, progress tracking and output lives with caller. While we work out
	// those details, it lives here.
	progress := pb.StartNew(numVersions)
	cbStates := make([]*CookbookStateRecord, 0, numVersions)

	// Second pass: let's pack them up into returnable form.
	for cookbookName, cookbookVersions := range cookbooks {
		for _, cookbookVersion := range cookbookVersions.Versions {
			dirName := fmt.Sprintf(".analyze-cache/cookbooks/%v-%v", cookbookName, cookbookVersion.Version)
			cbState = &CookbookStateRecord{Name: cookbookName,
				path:    dirName,
				Version: cookbookVersion.Version,
			}

			// TODO add pagination
			pres, err := searcher.PartialExec("node", fmt.Sprintf("cookbooks_%s_version:%s", cookbookName, cookbookVersion.Version), query)
			if err != nil {
				cbState.UsageLookupError = err
				cbStates = append(cbStates, cbState)
				progress.Increment()
				continue
			}
			if numNodes := len(pres.Rows); numNodes > 0 || includeUnboundCookbooks {
				if _, err := os.Stat("./" + dirName); os.IsNotExist(err) {
					// TODO - cbi.DownloadTo returns path so we don't have to know how it builds them
					err := cbi.DownloadTo(cookbookName, cookbookVersion.Version, ".analyze-cache/cookbooks")
					if err != nil {
						cbState.DownloadError = err
						cbStates = append(cbStates, cbState)
						continue
					}
				}
				nodesUsing := make([]string, 0, numNodes)
				for _, element := range pres.Rows {
					v := element.(map[string]interface{})["data"].(map[string]interface{})
					if v != nil {
						nodesUsing = append(nodesUsing, safeStringFromMap(v, "name"))
					}
				}
				cbState.Nodes = nodesUsing
				cbStates = append(cbStates, cbState)
				progress.Increment()
			} else {
				progress.Increment()
				continue
			}
		}
	}
	progress.Finish()
	return cbStates, nil
}

func runCookstyle(cbStates []*CookbookStateRecord) {
	progress := pb.StartNew(len(cbStates))
	runner := ExecCommandRunner{}
	for _, cb := range cbStates {
		progress.Increment()
		// an accurate set of results
		if cb.DownloadError != nil {
			continue
		}
		cookstyleResults, err := RunCookstyle(cb.path, runner)
		if err != nil {
			cb.CookstyleError = err
			continue
		}

		for _, file := range cookstyleResults.Files {
			cb.Files = append(cb.Files, file)
		}

	}
	progress.Finish()
}
