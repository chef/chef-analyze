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
	"sync"

	chef "github.com/chef/go-chef"
	"github.com/cheggaaa/pb/v3"
	"github.com/pkg/errors"
)

type CookbookState struct {
	Records        []*CookbookStateRecord
	TotalCookbooks int
	Cookbooks      CookbookInterface
	Searcher       SearchInterface
	Cookstyle      ExecCookstyleRunner
}

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

func NewCookbookState(cbi CookbookInterface, searcher SearchInterface, runner ExecCookstyleRunner, includeUnboundCookbooks bool) (*CookbookState, error) {
	fmt.Println("Finding available cookbooks...") // c <- ProgressUpdate(Event: COOKBOOK_FETCH)
	// Version limit of "0" means fetch all
	results, err := cbi.ListAvailableVersions("0")
	if err != nil {
		return nil, errors.Wrap(err, "unable to retrieve cookbooks")
	}

	// get totals so we can accurately report progress and allocate results
	totalCookbooks := 0
	for _, versions := range results {
		totalCookbooks += len(versions.Versions)
	}

	cookbookState := &CookbookState{
		Records:        make([]*CookbookStateRecord, totalCookbooks, totalCookbooks),
		TotalCookbooks: totalCookbooks,
		Cookbooks:      cbi,
		Searcher:       searcher,
		Cookstyle:      runner,
	}

	fmt.Println("Downloading cookbooks...")
	cookbookState.downloadCookbooks(results)

	fmt.Println("Analyzing cookbooks...")
	cookbookState.runCookstyle()

	return cookbookState, nil
}

// start a progress bar, launch all downloads in parallel and wait for all goroutines to finish
func (cbs *CookbookState) downloadCookbooks(cookbooks chef.CookbookListResult) {
	var (
		index    int
		wg       sync.WaitGroup
		progress = pb.StartNew(cbs.TotalCookbooks)
	)

	wg.Add(cbs.TotalCookbooks)

	for cookbookName, cookbookVersions := range cookbooks {
		for _, ver := range cookbookVersions.Versions {
			// @afiune we might want to launch batches of goroutines, say 100 downloads at once
			go cbs.downloadCookbook(cookbookName, ver.Version, progress, index, &wg)
			index++
		}
	}

	wg.Wait()

	progress.Finish()
}

func (cbs *CookbookState) downloadCookbook(cookbookName, version string, progress *pb.ProgressBar, index int, wg *sync.WaitGroup) {
	defer wg.Done()

	cbState := &CookbookStateRecord{Name: cookbookName,
		path:    fmt.Sprintf(".analyze-cache/cookbooks/%v-%v", cookbookName, version),
		Version: version,
	}

	nodes, err := nodesUsingCookbookVersion(cbs.Searcher, cookbookName, version)
	if err != nil {
		cbState.UsageLookupError = err
	} else {
		cbState.Nodes = nodes
	}

	err = cbs.Cookbooks.DownloadTo(cookbookName, version, ".analyze-cache/cookbooks")
	if err != nil {
		cbState.DownloadError = err
	}

	cbs.Records[index] = cbState
	progress.Increment()
}

func (cbs *CookbookState) runCookstyle() {
	progress := pb.StartNew(cbs.TotalCookbooks)

	for _, cb := range cbs.Records {
		progress.Increment()
		// an accurate set of results
		if cb.DownloadError != nil {
			continue
		}
		cookstyleResults, err := RunCookstyle(cb.path, cbs.Cookstyle)
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

func nodesUsingCookbookVersion(searcher SearchInterface, cookbook string, version string) ([]string, error) {
	var (
		query = map[string]interface{}{
			"name": []string{"name"},
		}
	)

	// TODO add pagination
	pres, err := searcher.PartialExec("node", fmt.Sprintf("cookbooks_%s_version:%s", cookbook, version), query)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get cookbook usage information")
	}

	// @afiune what is this includeUnboundCookbooks?
	results := make([]string, 0, len(pres.Rows))
	for _, element := range pres.Rows {
		v := element.(map[string]interface{})["data"].(map[string]interface{})
		if v != nil {
			results = append(results, safeStringFromMap(v, "name"))
		}
	}

	return results, nil
}
