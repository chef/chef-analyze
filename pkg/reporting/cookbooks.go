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
	MaxParallelWorkers = 50
)

type CookbooksStatus struct {
	Records        []*CookbookRecord
	TotalCookbooks int
	SkipUnused     bool
	Cookbooks      CookbookInterface
	Searcher       SearchInterface
	Cookstyle      *CookstyleRunner
	progress       *pb.ProgressBar
}

type CookbookRecord struct {
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

func (r *CookbookRecord) NumOffenses() int {
	i := 0
	for _, f := range r.Files {
		i += len(f.Offenses)
	}
	return i
}

func (r *CookbookRecord) NumCorrectable() int {
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

func NewCookbooks(cbi CookbookInterface, searcher SearchInterface, skipUnused bool) (*CookbooksStatus, error) {
	fmt.Printf("Finding available cookbooks...") // c <- ProgressUpdate(Event: COOKBOOK_FETCH)
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
	fmt.Printf(" (%d found)\n", totalCookbooks)

	var (
		downloadCh     = make(chan cookbookItem)
		analyzeCh      = make(chan *CookbookRecord)
		doneCh         = make(chan bool)
		cookbooksState = &CookbooksStatus{
			Records:        make([]*CookbookRecord, 0, totalCookbooks),
			TotalCookbooks: totalCookbooks,
			Cookbooks:      cbi,
			Searcher:       searcher,
			Cookstyle:      NewCookstyleRunner(),
			SkipUnused:     skipUnused,
		}
	)

	if totalCookbooks == 0 {
		fmt.Println("No cookbooks available for analysis")
		return cookbooksState, nil
	}

	// determine how many workers do we need, by default, the total number of cookbooks
	numWorkers := totalCookbooks
	if totalCookbooks > MaxParallelWorkers {
		// but if the total number of cookbooks is major than
		// the maximum allowed, set it to the maximum
		numWorkers = MaxParallelWorkers
	}

	fmt.Println("Analyzing cookbooks...")
	// start and store a progress bar
	cookbooksState.progress = pb.StartNew(totalCookbooks)

	// launch jobs that will be read by the workers (goroutines)
	go cookbooksState.triggerJobs(results, downloadCh)

	// launch the download workers, these will process downloads and send them to the next
	// channel to be analyzed, once all messages have been red, it closes the download channel
	go cookbooksState.createDownloadWorkerPool(numWorkers, downloadCh, analyzeCh)

	// launch the analyze workers, these will process analyzis by running cookstyle,
	// once all messages have been processed, it will send a single message to the done channel
	go cookbooksState.createAnalyzeWorkerPool(numWorkers, analyzeCh, doneCh)

	// wait for a message in from the done channel
	// to make sure there are no more processes running
	<-doneCh
	// make sure the progress bar reports we are done
	cookbooksState.progress.Finish()

	return cookbooksState, nil
}

func (cbs *CookbooksStatus) triggerJobs(cookbooks chef.CookbookListResult, inCh chan<- cookbookItem) {
	for cookbookName, cookbookVersions := range cookbooks {
		for _, ver := range cookbookVersions.Versions {
			inCh <- cookbookItem{cookbookName, ver.Version}
		}
	}
	//debug("%s: finished sending jobs", time.Now())
	close(inCh)
}

func (cbs *CookbooksStatus) createDownloadWorkerPool(nWorkers int, downloadCh <-chan cookbookItem, analyzeCh chan<- *CookbookRecord) {
	var wg sync.WaitGroup

	for i := 0; i < nWorkers; i++ {
		wg.Add(1)
		go func(inCh <-chan cookbookItem, outCh chan<- *CookbookRecord, wg *sync.WaitGroup) {
			for item := range inCh {
				cbs.downloadCookbook(item.Name, item.Version, outCh)
			}
			wg.Done()
		}(downloadCh, analyzeCh, &wg)
	}

	wg.Wait()
	//debug("%s: finished downloading", time.Now())
	close(analyzeCh)
}

func (cbs *CookbooksStatus) createAnalyzeWorkerPool(nWorkers int, analyzeCh <-chan *CookbookRecord, doneCh chan<- bool) {
	var wg sync.WaitGroup

	for i := 0; i < nWorkers; i++ {
		wg.Add(1)
		go func(inCh <-chan *CookbookRecord, wg *sync.WaitGroup) {
			for record := range inCh {
				cbs.Records = append(cbs.Records, record)
				cbs.runCookstyleFor(record)
			}
			wg.Done()
		}(analyzeCh, &wg)
	}

	wg.Wait()
	//debug("%s: finished analyze", time.Now())
	doneCh <- true
}

func (cbs *CookbooksStatus) downloadCookbook(cookbookName, version string, analyzeCh chan<- *CookbookRecord) {
	cbState := &CookbookRecord{Name: cookbookName,
		path:    fmt.Sprintf("%s/cookbooks/%v-%v", AnalyzeCacheDir, cookbookName, version),
		Version: version,
	}

	nodes, err := cbs.nodesUsingCookbookVersion(cookbookName, version)
	if err != nil {
		cbState.UsageLookupError = err
	}
	cbState.Nodes = nodes

	// when there are no nodes using this cookbook and the skip unused flag was provided,
	// exit without downloading and increment the progress bar since we are donw processing
	if len(nodes) == 0 && cbs.SkipUnused {
		cbs.progress.Increment()
		return
	}

	err = cbs.Cookbooks.DownloadTo(cookbookName, version, fmt.Sprintf("%s/cookbooks", AnalyzeCacheDir))
	if err != nil {
		cbState.DownloadError = err
	}

	// move to store and analyze the cookbook record
	analyzeCh <- cbState
}

func (cbs *CookbooksStatus) nodesUsingCookbookVersion(cookbook string, version string) ([]string, error) {
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

func (cbs *CookbooksStatus) runCookstyleFor(cb *CookbookRecord) {
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
