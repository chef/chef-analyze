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
	"github.com/pkg/errors"
)

const (
	// cached directory
	AnalyzeCacheDir = ".analyze-cache"

	// maximum number of parallel downloads at once
	MaxParallelDownloads = 100

	// maximum number of parallel cookstyle analysis at once
	MaxParallelCookstyleAnalysis = 100
)

const (
	CBStateActionStarted = iota
	CBStateActionComplete
	CBStateActionError
)
const (
	CBStateJobFetch = iota
	CBStateJobDownload
	CBStateJobCookstyle
)

type ItemUpdate struct {
	Action int
	Item   *CookbookStateRecord
}

type JobUpdate struct {
	JobEvent int
	Action   int
	Count    int
	Err      error
}

type CookbookState struct {
	Records        []*CookbookStateRecord
	TotalCookbooks int
	SkipUnused     bool
	// TODO _ should we make these private? Caller shouldn't need them.
	Cookbooks CookbookInterface
	Searcher  SearchInterface
	Cookstyle CookstyleRunner
	// Status         struct {
	JobCompletion   chan *JobUpdate
	DownloadStatus  chan *ItemUpdate
	CookstyleStatus chan *ItemUpdate
	//	}
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

func NewCookbookState(cbi CookbookInterface, searcher SearchInterface, skipUnused bool) *CookbookState {
	return &CookbookState{
		Cookbooks:       cbi,
		Searcher:        searcher,
		Cookstyle:       NewCookstyleRunner(),
		SkipUnused:      skipUnused,
		JobCompletion:   make(chan *JobUpdate),
		DownloadStatus:  make(chan *ItemUpdate),
		CookstyleStatus: make(chan *ItemUpdate),
	}
}

func (cbs *CookbookState) Run() error {

	cbs.jobStatusUpdate(CBStateJobFetch, CBStateActionStarted, 0, nil)

	results, err := cbs.Cookbooks.ListAvailableVersions("0")
	// TODO - lets pass error in the error update.
	if err != nil {
		cbs.jobStatusUpdate(CBStateJobFetch, CBStateActionError, 0, err)
		return errors.Wrap(err, "unable to retrieve cookbooks")
	}

	// get totals so we can accurately report progress and allocate results
	totalCookbooks := 0
	for _, versions := range results {
		totalCookbooks += len(versions.Versions)
	}

	cbs.Records = make([]*CookbookStateRecord, totalCookbooks, totalCookbooks)
	cbs.TotalCookbooks = totalCookbooks
	cbs.jobStatusUpdate(CBStateJobFetch, CBStateActionComplete, totalCookbooks, nil)

	cbs.downloadCookbooks(results, totalCookbooks)

	cbs.analyzeCookbooks()
	return nil
}

func (cbs *CookbookState) jobStatusUpdate(event int, action int, count int, err error) {
	cbs.JobCompletion <- &JobUpdate{event, action, count, err}
}

func (cbs *CookbookState) downloadStatusUpdate(action int, cbState *CookbookStateRecord) {
	cbs.DownloadStatus <- &ItemUpdate{action, cbState}
}

func (cbs *CookbookState) analyzeStatusUpdate(action int, cbState *CookbookStateRecord) {
	cbs.CookstyleStatus <- &ItemUpdate{action, cbState}
}

// start a progress bar, launch batches of parallel downloads and wait for all goroutines to finish
func (cbs *CookbookState) downloadCookbooks(cookbooks chef.CookbookListResult, totalCookbooks int) {
	var (
		index      int
		goroutines int
		wg         sync.WaitGroup
	)

	cbs.jobStatusUpdate(CBStateJobDownload, CBStateActionStarted, totalCookbooks, nil)
	for cookbookName, cookbookVersions := range cookbooks {
		for _, ver := range cookbookVersions.Versions {
			wg.Add(1)
			go cbs.downloadCookbook(cookbookName, ver.Version, index, &wg)
			goroutines++
			index++

			// @afiune launch batches of goroutines, say 100 downloads at once
			if goroutines >= MaxParallelDownloads {
				wg.Wait()
				goroutines = 0
			}
		}
	}
	defer cbs.jobStatusUpdate(CBStateJobDownload, CBStateActionComplete, totalCookbooks, nil)
	// make sure there are no more processes running
	wg.Wait()
}
func (cbs *CookbookState) downloadCookbook(cookbookName, version string, index int, wg *sync.WaitGroup) {

	defer wg.Done()

	cbState := &CookbookStateRecord{Name: cookbookName,
		path:    fmt.Sprintf("%s/cookbooks/%v-%v", AnalyzeCacheDir, cookbookName, version),
		Version: version,
	}
	cbs.downloadStatusUpdate(CBStateActionStarted, cbState)

	exitAction := CBStateActionComplete
	defer cbs.downloadStatusUpdate(exitAction, cbState)

	cbs.Records[index] = cbState
	nodes, err := cbs.nodesUsingCookbookVersion(cookbookName, version)
	if err != nil {
		exitAction = CBStateActionError
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
}

// start a progress bar, launch batches of parallel analysis and wait for all goroutines to finish
func (cbs *CookbookState) analyzeCookbooks() {
	var (
		goroutines int
		wg         sync.WaitGroup
	)

	cbs.jobStatusUpdate(CBStateJobCookstyle, CBStateActionStarted, len(cbs.Records), nil)
	for _, cb := range cbs.Records {
		wg.Add(1)
		go cbs.runCookstyleFor(cb, &wg)
		goroutines++

		if goroutines >= MaxParallelCookstyleAnalysis {
			// wait for the batch of goroutines to finish, then continue
			wg.Wait()
			goroutines = 0
		}
	}

	defer cbs.jobStatusUpdate(CBStateJobCookstyle, CBStateActionComplete, len(cbs.Records), nil)
	// make sure there are no more processes running
	wg.Wait()
}

func (cbs *CookbookState) runCookstyleFor(cb *CookbookStateRecord, wg *sync.WaitGroup) {
	defer wg.Done()

	cbs.analyzeStatusUpdate(CBStateActionStarted, cb)
	exitAction := CBStateActionComplete
	defer cbs.analyzeStatusUpdate(exitAction, cb)
	// an accurate set of results
	if cb.DownloadError != nil {
		return
	}

	// skip unused cookbooks
	if len(cb.Nodes) == 0 && cbs.SkipUnused {
		return
	}

	cookstyleResults, err := cbs.Cookstyle.Run(cb.path)
	if err != nil {
		exitAction = CBStateActionError
		cb.CookstyleError = err
		return
	}

	for _, file := range cookstyleResults.Files {
		cb.Files = append(cb.Files, file)
	}

}

func (cbs *CookbookState) nodesUsingCookbookVersion(cookbook string, version string) ([]string, error) {
	// Tell partial search that we only want one field back:
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
