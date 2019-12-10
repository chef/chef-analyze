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

// PipelineStatus maintains the overall state of the process (download, find nodes, run cookstyle)
type PipelineStatus struct {
	Records        []*CookbookRecord
	RecordsMutex   sync.Mutex
	TotalCookbooks int
	OnlyUnused     bool
	RunCookstyle   bool
	CookbooksDir   string
	Cookbooks      CookbookInterface
	Searcher       SearchInterface
	Cookstyle      *CookstyleRunner
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

func (r *CookbookRecord) NumNodesAffected() int {
	return len(r.Nodes)
}

// NumOffenses collects the total number of cookstyle offenses in the cookbook
func (r *CookbookRecord) NumOffenses() int {
	i := 0
	for _, f := range r.Files {
		i += len(f.Offenses)
	}
	return i
}

// NumCorrectable collects the number of correctable cookstyle offenses in the cookbook
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

// StartPipeline is the business logic of performing the analysis
func StartPipeline(
	cbi CookbookInterface, searcher SearchInterface,
	runCookstyle, onlyUnused bool, workers int,
) (*CookbooksStatus, error) {
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
		downloadCh     = make(chan cookbookItem)
		analyzeCh      = make(chan *CookbookRecord)
		doneCh         = make(chan bool)
		pipelineStatus = &PipelineStatus{
			Records:        make([]*CookbookRecord, 0, totalCookbooks),
			TotalCookbooks: totalCookbooks,
			Cookbooks:      cbi,
			Searcher:       searcher,
			Cookstyle:      NewCookstyleRunner(),
			RunCookstyle:   runCookstyle,
			CookbooksDir:   cookbooksDir,
			OnlyUnused:     onlyUnused,
		}
	)

	if totalCookbooks == 0 {
		fmt.Println("No cookbooks available for analysis")
		return pipelineStatus, nil
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
	pipelineStatus.progress = pb.StartNew(totalCookbooks)

	// launch jobs that will be read by the workers (goroutines)
	go pipelineStatus.triggerJobs(results, downloadCh)

	// launch the download workers, these will process downloads and send them to the next
	// channel to be analyzed, once all messages have been red, it closes the download channel
	go pipelineStatus.createDownloadWorkerPool(numWorkers, downloadCh, analyzeCh)

	// launch the analyze workers, these will process analyzis by running cookstyle,
	// once all messages have been processed, it will send a single message to the done channel
	go pipelineStatus.createAnalyzeWorkerPool(numWorkers, analyzeCh, doneCh)

	// wait for a message in from the done channel
	// to make sure there are no more processes running
	<-doneCh
	// make sure the progress bar reports we are done
	pipelineStatus.progress.Finish()

	return pipelineStatus, nil
}

func (p *PipelineStatus) addRecord(r *CookbookRecord) {
	p.RecordsMutex.Lock()
	defer p.RecordsMutex.Unlock()
	p.Records = append(p.Records, r)
}

func (p *PipelineStatus) triggerJobs(cookbooks chef.CookbookListResult, inCh chan<- cookbookItem) {
	for cookbookName, cookbookVersions := range cookbooks {
		for _, ver := range cookbookVersions.Versions {
			inCh <- cookbookItem{cookbookName, ver.Version}
		}
	}
	//debug("%s: finished sending jobs", time.Now())
	close(inCh)
}

func (p *PipelineStatus) createDownloadWorkerPool(nWorkers int, downloadCh <-chan cookbookItem, analyzeCh chan<- *CookbookRecord) {
	var wg sync.WaitGroup

	for i := 0; i < nWorkers; i++ {
		wg.Add(1)
		go func(inCh <-chan cookbookItem, outCh chan<- *CookbookRecord, wg *sync.WaitGroup) {
			for item := range inCh {
				p.downloadCookbook(item.Name, item.Version, outCh)
			}
			wg.Done()
		}(downloadCh, analyzeCh, &wg)
	}

	wg.Wait()
	//debug("%s: finished downloading", time.Now())
	close(analyzeCh)
}

func (p *PipelineStatus) createAnalyzeWorkerPool(nWorkers int, analyzeCh <-chan *CookbookRecord, doneCh chan<- bool) {
	var wg sync.WaitGroup

	for i := 0; i < nWorkers; i++ {
		wg.Add(1)
		go func(inCh <-chan *CookbookRecord, wg *sync.WaitGroup) {
			for record := range inCh {
				p.addRecord(record)

				if p.RunCookstyle {
					p.runCookstyleFor(record)
				}
			}
			wg.Done()
		}(analyzeCh, &wg)
	}

	wg.Wait()
	//debug("%s: finished analyze", time.Now())
	doneCh <- true
}

func (p *PipelineStatus) downloadCookbook(cookbookName, version string, analyzeCh chan<- *CookbookRecord) {

	var (
		nodes, err       = p.nodesUsingCookbookVersion(cookbookName, version)
		cookbookLongName = fmt.Sprintf("%v-%v", cookbookName, version)
		cbState          = &CookbookRecord{
			Name:    cookbookName,
			path:    filepath.Join(cbs.CookbooksDir, cookbookLongName),
			Version: version,
		}
	)
	if err != nil {
		cbState.UsageLookupError = err
	}
	cbState.Nodes = nodes

	// by default we report only cookbooks that are being used by one or more nodes,
	// but we also provide a way to report the opposite, that is, only unused cookbooks
	if p.OnlyUnused {
		// report only unused cookbooks
		if len(nodes) > 0 {
			p.progress.Increment()
			return
		}
	} else {
		// report only cookbooks being used
		if len(nodes) == 0 {
			p.progress.Increment()
			return
		}
	}

	// do we need to analyze the cookbooks
	if cbs.RunCookstyle {
		err = p.Cookbooks.DownloadTo(cookbookName, version, fmt.Sprintf("%s/cookbooks", AnalyzeCacheDir))
		if err != nil {
			cbState.DownloadError = errors.Wrapf(err, "unable to download cookbook %s", cookbookName)
		}
	}

	// move to store and analyze the cookbook record
	analyzeCh <- cbState
}

func (p *PipelineStatus) nodesUsingCookbookVersion(cookbook string, version string) ([]string, error) {
	query := map[string]interface{}{
		"name": []string{"name"},
	}

	// TODO add pagination
	pres, err := p.Searcher.PartialExec("node", fmt.Sprintf("cookbooks_%s_version:%s", cookbook, version), query)
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

func (p *PipelineStatus) runCookstyleFor(cb *CookbookRecord) {
	defer p.progress.Increment()

	// an accurate set of results
	if cb.DownloadError != nil {
		return
	}

	cookstyleResults, err := p.Cookstyle.Run(cb.path)
	if err != nil {
		cb.CookstyleError = err
		return
	}

	for _, file := range cookstyleResults.Files {
		cb.Files = append(cb.Files, file)
	}
}
