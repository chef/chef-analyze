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
	"path/filepath"
	"sync"

	chef "github.com/chef/go-chef"
	"github.com/chef/go-libs/config"
	"github.com/cheggaaa/pb/v3"
	"github.com/pkg/errors"
)

const (
	analyzeCacheDir     = "cache"     // Used for $HOME/.chef-workstation/cache
	analyzeCookbooksDir = "cookbooks" // Used for $HOME/.chef-workstation/cache/cookbooks
)

// CookbooksReport maintains the overall state of the process (download, find nodes, run cookstyle)
type CookbooksReport struct {
	Records        []*CookbookRecord
	recordsMutex   sync.Mutex
	TotalCookbooks int
	onlyUnused     bool
	RunCookstyle   bool
	cookbooksDir   string
	cookbooks      CookbookInterface
	searcher       SearchInterface
	cookstyle      *CookstyleRunner
	progress       *pb.ProgressBar
}

// CookbookRecord is a single cookbook that we want to download and analyze
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

// Errors collates all known errors
func (cr CookbookRecord) Errors() []error {
	errs := make([]error, 0)
	if cr.DownloadError != nil {
		errs = append(errs, cr.DownloadError)
	}
	if cr.UsageLookupError != nil {
		errs = append(errs, cr.UsageLookupError)
	}
	if cr.CookstyleError != nil {
		errs = append(errs, cr.CookstyleError)
	}
	return errs
}

// internally used to submit items to the workers
type cookbookItem struct {
	Name    string
	Version string
}

// NumNodesAffected shortcut
func (cr *CookbookRecord) NumNodesAffected() int {
	return len(cr.Nodes)
}

// NumOffenses collects the total number of cookstyle offenses in the cookbook
func (cr *CookbookRecord) NumOffenses() int {
	i := 0
	for _, f := range cr.Files {
		i += len(f.Offenses)
	}
	return i
}

// NumCorrectable collects the number of correctable cookstyle offenses in the cookbook
func (cr *CookbookRecord) NumCorrectable() int {
	i := 0
	for _, f := range cr.Files {
		for _, o := range f.Offenses {
			if o.Correctable {
				i++
			}
		}
	}
	return i
}

// GenerateCookbooksReport is the business logic of performing the analysis
func GenerateCookbooksReport(
	cbi CookbookInterface, searcher SearchInterface,
	runCookstyle, onlyUnused bool, workers int,
) (*CookbooksReport, error) {
	wsDir, err := config.ChefWorkstationDir()
	if err != nil {
		return nil, err
	}
	cookbooksDir := filepath.Join(wsDir, analyzeCacheDir, analyzeCookbooksDir)
	//debug("%s: using cookbooks directory %s", time.Now(), cookbooksDir)

	fmt.Printf("Finding available cookbooks...") // c <- ProgressUpdate(Event: COOKBOOK_FETCH)
	// Version limit of "0" means fetch all
	results, err := cbi.ListAvailableVersions("0")
	if err != nil {
		// carrier return so we output the actual error message on a new line
		// why? because of the Printf above.
		// TODO delete this once we have a progress update
		fmt.Println(" (-)")
		return nil, errors.Wrap(err, "unable to retrieve cookbooks")
	}

	// get totals so we can accurately report progress and allocate results
	totalCookbooks := 0
	for _, versions := range results {
		totalCookbooks += len(versions.Versions)
	}
	fmt.Printf(" (%d found)\n", totalCookbooks)

	var (
		downloadCh   = make(chan cookbookItem)
		analyzeCh    = make(chan *CookbookRecord)
		doneCh       = make(chan bool)
		reportStatus = &CookbooksReport{
			Records:        make([]*CookbookRecord, 0, totalCookbooks),
			TotalCookbooks: totalCookbooks,
			cookbooks:      cbi,
			searcher:       searcher,
			cookstyle:      NewCookstyleRunner(),
			RunCookstyle:   runCookstyle,
			cookbooksDir:   cookbooksDir,
			onlyUnused:     onlyUnused,
		}
	)

	if totalCookbooks == 0 {
		fmt.Println("No cookbooks available for analysis")
		return reportStatus, nil
	}

	// determine how many workers do we need, by default, the total number of cookbooks
	numWorkers := totalCookbooks
	if totalCookbooks > workers {
		// but if the total number of cookbooks is major than
		// the maximum allowed, set it to the maximum
		numWorkers = workers
	}

	fmt.Println("Analyzing cookbooks...")
	// start and store a progress bar
	reportStatus.progress = pb.StartNew(totalCookbooks)

	// launch jobs that will be read by the workers (goroutines)
	go reportStatus.triggerJobs(results, downloadCh)

	// launch the download workers, these will process downloads and send them to the next
	// channel to be analyzed, once all messages have been red, it closes the download channel
	go reportStatus.createDownloadWorkerPool(numWorkers, downloadCh, analyzeCh)

	// launch the analyze workers, these will process analyzis by running cookstyle,
	// once all messages have been processed, it will send a single message to the done channel
	go reportStatus.createAnalyzeWorkerPool(numWorkers, analyzeCh, doneCh)

	// wait for a message in from the done channel
	// to make sure there are no more processes running
	<-doneCh
	// make sure the progress bar reports we are done
	reportStatus.progress.Finish()

	return reportStatus, nil
}

func (cbr *CookbooksReport) addRecord(r *CookbookRecord) {
	cbr.recordsMutex.Lock()
	defer cbr.recordsMutex.Unlock()
	cbr.Records = append(cbr.Records, r)
}

func (cbr *CookbooksReport) triggerJobs(cookbooks chef.CookbookListResult, inCh chan<- cookbookItem) {
	for cookbookName, cookbookVersions := range cookbooks {
		for _, ver := range cookbookVersions.Versions {
			inCh <- cookbookItem{cookbookName, ver.Version}
		}
	}
	//debug("%s: finished sending jobs", time.Now())
	close(inCh)
}

func (cbr *CookbooksReport) createDownloadWorkerPool(nWorkers int, downloadCh <-chan cookbookItem, analyzeCh chan<- *CookbookRecord) {
	var wg sync.WaitGroup

	for i := 0; i < nWorkers; i++ {
		wg.Add(1)
		go func(inCh <-chan cookbookItem, outCh chan<- *CookbookRecord, wg *sync.WaitGroup) {
			for item := range inCh {
				cbr.downloadCookbook(item.Name, item.Version, outCh)
			}
			wg.Done()
		}(downloadCh, analyzeCh, &wg)
	}

	wg.Wait()
	//debug("%s: finished downloading", time.Now())
	close(analyzeCh)
}

func (cbr *CookbooksReport) createAnalyzeWorkerPool(nWorkers int, analyzeCh <-chan *CookbookRecord, doneCh chan<- bool) {
	var wg sync.WaitGroup

	for i := 0; i < nWorkers; i++ {
		wg.Add(1)
		go func(inCh <-chan *CookbookRecord, wg *sync.WaitGroup) {
			for record := range inCh {
				cbr.addRecord(record)

				if cbr.RunCookstyle {
					cbr.runCookstyleFor(record)
				}
			}
			wg.Done()
		}(analyzeCh, &wg)
	}

	wg.Wait()
	//debug("%s: finished analyze", time.Now())
	doneCh <- true
}

func (cbr *CookbooksReport) downloadCookbook(cookbookName, version string, analyzeCh chan<- *CookbookRecord) {

	var (
		nodes, err       = cbr.nodesUsingCookbookVersion(cookbookName, version)
		cookbookLongName = fmt.Sprintf("%v-%v", cookbookName, version)
		cbState          = &CookbookRecord{
			Name:    cookbookName,
			path:    filepath.Join(cbr.cookbooksDir, cookbookLongName),
			Version: version,
		}
	)
	if err != nil {
		cbState.UsageLookupError = err
	}
	cbState.Nodes = nodes

	// by default we report only cookbooks that are being used by one or more nodes,
	// but we also provide a way to report the opposite, that is, only unused cookbooks
	if cbr.onlyUnused {
		// report only unused cookbooks
		if len(nodes) > 0 {
			cbr.progress.Increment()
			return
		}
	} else {
		// report only cookbooks being used
		if len(nodes) == 0 {
			cbr.progress.Increment()
			return
		}
	}

	// do we need to analyze the cookbooks
	if cbr.RunCookstyle {
		err = cbr.cookbooks.DownloadTo(cookbookName, version, cbr.cookbooksDir)
		if err != nil {
			cbState.DownloadError = errors.Wrapf(err, "unable to download cookbook %s", cookbookName)
		}
	}

	// move to store and analyze the cookbook record
	analyzeCh <- cbState
}

func (cbr *CookbooksReport) nodesUsingCookbookVersion(cookbook string, version string) ([]string, error) {
	query := map[string]interface{}{
		"name": []string{"name"},
	}

	// TODO add pagination
	pres, err := cbr.searcher.PartialExec("node", fmt.Sprintf("cookbooks_%s_version:%s", cookbook, version), query)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get cookbook usage information")
	}

	// If the error is unrelated to returning any results we want to parse them and return them.
	results := make([]string, 0, len(pres.Rows))
	for _, element := range pres.Rows {
		v := element.(map[string]interface{})["data"].(map[string]interface{})
		if v != nil {
			results = append(results, safeStringFromMap(v, "name"))
		}
	}

	return results, nil
}

func (cbr *CookbooksReport) runCookstyleFor(cb *CookbookRecord) {
	defer cbr.progress.Increment()

	// an accurate set of results
	if cb.DownloadError != nil {
		return
	}

	cookstyleResults, err := cbr.cookstyle.Run(cb.path)
	if err != nil {
		cb.CookstyleError = err
		return
	}

	for _, file := range cookstyleResults.Files {
		cb.Files = append(cb.Files, file)
	}
}
