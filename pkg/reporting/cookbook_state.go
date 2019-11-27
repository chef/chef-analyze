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

const (
	// cached directory
	AnalyzeCacheDir = ".analyze-cache"

	// maximum number of parallel workers at once
	MaxParallelWorkers = 100
)

type CookbookState struct {
	Records        []*CookbookStateRecord
	TotalCookbooks int
	SkipUnused     bool
	Cookbooks      CookbookInterface
	Searcher       SearchInterface
	Cookstyle      CookstyleRunner
	progress       *pb.ProgressBar
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

// internally used to submit items to the workers
type cookbookItem struct {
	Name    string
	Version string
}

func (r *CookbookStateRecord) NumOffenses() int {
	i := 0
	for _, f := range r.Files {
		i += len(f.Offenses)
	}
	return i
}

func (r *CookbookStateRecord) NumCorrectable() int {
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

func NewCookbookState(cbi CookbookInterface, searcher SearchInterface, skipUnused bool) (*CookbookState, error) {
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
		Records:        make([]*CookbookStateRecord, 0, totalCookbooks),
		TotalCookbooks: totalCookbooks,
		Cookbooks:      cbi,
		Searcher:       searcher,
		Cookstyle:      NewCookstyleRunner(),
		SkipUnused:     skipUnused,
	}

	fmt.Println("Downloading cookbooks...")
	cookbookState.downloadCookbooks(results)

	// update total cookbooks since we could have reduced the list by the skip-unused flag
	cookbookState.TotalCookbooks = len(cookbookState.Records)

	fmt.Println("Analyzing cookbooks...")
	cookbookState.analyzeCookbooks()

	return cookbookState, nil
}

// start a progress bar, launch batches of parallel downloads and wait for all goroutines to finish
func (cbs *CookbookState) downloadCookbooks(cookbooks chef.CookbookListResult) {
	var (
		progress = pb.StartNew(cbs.TotalCookbooks)
		inCh     = make(chan cookbookItem, MaxParallelWorkers)
		outCh    = make(chan *CookbookStateRecord, MaxParallelWorkers)
		doneCh   = make(chan bool)
	)

	// store the progress bar
	cbs.progress = progress

	// this goroutine is in charge to store the records of the CookbookState
	go cbs.recordFromCh(outCh, doneCh)

	// launch jobs that will be read by the workers (goroutines)
	go cbs.triggerDownloadJobs(cookbooks, inCh)

	cbs.createDownloadWorkerPool(MaxParallelWorkers, inCh, outCh)

	// make sure there are no more processes running
	<-doneCh
	progress.Finish()
}

func (cbs *CookbookState) triggerDownloadJobs(cookbooks chef.CookbookListResult, inCh chan cookbookItem) {
	for cookbookName, cookbookVersions := range cookbooks {
		for _, ver := range cookbookVersions.Versions {
			inCh <- cookbookItem{cookbookName, ver.Version}
		}
	}
	close(inCh)
}

func (cbs *CookbookState) createDownloadWorkerPool(nWorkers int, inCh chan cookbookItem, outCh chan *CookbookStateRecord) {
	var wg sync.WaitGroup

	for i := 0; i < nWorkers; i++ {
		wg.Add(1)
		go cbs.downloadWorker(inCh, outCh, &wg)
	}

	wg.Wait()
	close(outCh)
}

func (cbs *CookbookState) downloadWorker(inCh chan cookbookItem, outCh chan *CookbookStateRecord, wg *sync.WaitGroup) {
	defer wg.Done()

	for cookbook := range inCh {
		cbs.downloadCookbook(cookbook.Name, cookbook.Version, outCh)
	}
}

func (cbs *CookbookState) recordFromCh(outCh chan *CookbookStateRecord, doneCh chan bool) {
	for csr := range outCh {
		cbs.Records = append(cbs.Records, csr)
	}
	doneCh <- true
}

func (cbs *CookbookState) downloadCookbook(cookbookName, version string, outCh chan *CookbookStateRecord) {
	defer cbs.progress.Increment()

	cbState := &CookbookStateRecord{Name: cookbookName,
		path:    fmt.Sprintf("%s/cookbooks/%v-%v", AnalyzeCacheDir, cookbookName, version),
		Version: version,
	}

	nodes, err := cbs.nodesUsingCookbookVersion(cookbookName, version)
	if err != nil {
		cbState.UsageLookupError = err
	}
	cbState.Nodes = nodes

	// when there are no nodes using this cookbook and the skip unused flag was provided, exit without downloading
	if len(nodes) == 0 && cbs.SkipUnused {
		return
	}

	err = cbs.Cookbooks.DownloadTo(cookbookName, version, fmt.Sprintf("%s/cookbooks", AnalyzeCacheDir))
	if err != nil {
		cbState.DownloadError = err
	}

	// store record
	outCh <- cbState
}

func (cbs *CookbookState) nodesUsingCookbookVersion(cookbook string, version string) ([]string, error) {
	query := map[string]interface{}{
		"name": []string{"name"},
	}

	// TODO add pagination
	pres, err := cbs.Searcher.PartialExec("node", fmt.Sprintf("cookbooks_%s_version:%s", cookbook, version), query)
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

// start a progress bar, launch batches of parallel analysis and wait for all goroutines to finish
func (cbs *CookbookState) analyzeCookbooks() {
	var (
		progress = pb.StartNew(cbs.TotalCookbooks)
		inCh     = make(chan *CookbookStateRecord, MaxParallelWorkers)
	)

	// store the progress bar
	cbs.progress = progress

	// launch jobs that will be read by the workers (goroutines)
	go cbs.triggerAnalyzeJobs(inCh)

	cbs.createAnalyzeWorkerPool(MaxParallelWorkers, inCh)

	progress.Finish()
}

func (cbs *CookbookState) triggerAnalyzeJobs(inCh chan *CookbookStateRecord) {
	for _, cb := range cbs.Records {
		inCh <- cb
	}
	close(inCh)
}

func (cbs *CookbookState) createAnalyzeWorkerPool(nWorkers int, inCh chan *CookbookStateRecord) {
	var wg sync.WaitGroup

	for i := 0; i < nWorkers; i++ {
		wg.Add(1)
		go cbs.analyzeWorker(inCh, &wg)
	}

	wg.Wait()
}

func (cbs *CookbookState) analyzeWorker(inCh chan *CookbookStateRecord, wg *sync.WaitGroup) {
	defer wg.Done()

	for record := range inCh {
		cbs.runCookstyleFor(record)
	}
}

func (cbs *CookbookState) runCookstyleFor(cb *CookbookStateRecord) {
	defer cbs.progress.Increment()

	// an accurate set of results
	if cb.DownloadError != nil {
		return
	}

	cookstyleResults, err := cbs.Cookstyle.Run(cb.path)
	if err != nil {
		cb.CookstyleError = err
		return
	}

	for _, file := range cookstyleResults.Files {
		cb.Files = append(cb.Files, file)
	}
}
