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
	"strings"
	"sync"

	chef "github.com/chef/go-chef"
	"github.com/chef/go-libs/config"
	"github.com/pkg/errors"
)

const (
	analyzeCacheDir     = "cache"     // Used for $HOME/.chef-workstation/cache
	analyzeCookbooksDir = "cookbooks" // Used for $HOME/.chef-workstation/cache/cookbooks
)

// CookbooksReport maintains the overall state of the process (download, find nodes, run cookstyle)
type CookbooksReport struct {
	Records               []*CookbookRecord
	recordsMutex          sync.Mutex
	TotalCookbooks        int
	onlyUnused            bool
	RunCookstyle          bool
	NodeFilter            string
	cookbooksDir          string
	cookbooks             CookbookInterface
	cookbookArtifacts     CBAInterface
	searcher              SearchInterface
	cookstyle             *CookstyleRunner
	Progress              chan int
	numWorkers            int
	cookbookSearchResults chef.CookbookListResult
	CBASearchResults      []cookbookItem
	policyGroups          PolicyGroupInterface
	Policies              PolicyInterface
}

// CookbookRecord is a single cookbook that we want to download and analyze
type CookbookRecord struct {
	Name             string
	Version          string
	Identifier       string
	Files            []CookbookFile
	Nodes            []string
	path             string
	DownloadError    error
	UsageLookupError error
	CookstyleError   error
	Policy           string
	PolicyVer        string
	PolicyGroup      string
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

// Sort interface for cookbook recrods
// Order by Policy Group, Policy Name, Cookbook Name, Cookbook Version
type CookbookRecordsBySortOrder []*CookbookRecord

func (crs CookbookRecordsBySortOrder) Len() int {
	return len(crs)
}
func (crs CookbookRecordsBySortOrder) Swap(i, j int) {
	crs[i], crs[j] = crs[j], crs[i]
}

func (crs CookbookRecordsBySortOrder) Less(i, j int) bool {
	// Sort by the Policy Group
	l1, l2 := strings.ToLower(crs[i].PolicyGroup), strings.ToLower(crs[j].PolicyGroup)
	if l1 != l2 {
		return l1 < l2
	}

	// Sort by Polic Name
	l3, l4 := strings.ToLower(crs[i].Policy), strings.ToLower(crs[j].Policy)
	if l3 != l4 {
		return l3 < l4
	}

	l5, l6 := strings.ToLower(crs[i].Name), strings.ToLower(crs[j].Name)
	if l5 != l6 {
		return l5 < l6
	}

	return strings.ToLower(crs[i].Version) < strings.ToLower(crs[j].Version)
}

// internally used to submit items to the workers
type cookbookItem struct {
	Name          string
	Version       string
	CBAIdentifier string
	Policy        string
	PolicyGroup   string
	PolicyRev     string
}

func NewCookbooksReport(
	cbi CookbookInterface, cbai CBAInterface, pgi PolicyGroupInterface, poi PolicyInterface, searcher SearchInterface,
	runCookstyle bool, onlyUnused bool, workers int,
	nodeFilter string,
) (*CookbooksReport, error) {
	wsDir, err := config.ChefWorkstationDir()
	if err != nil {
		return nil, err
	}
	cookbooksDir := filepath.Join(wsDir, analyzeCacheDir, analyzeCookbooksDir)
	results, err := cbi.ListAvailableVersions("0")
	if err != nil {
		return nil, errors.Wrap(err, "unable to retrieve cookbooks")
	}
	resultsCBA, err := getCookbookArtifacts(pgi, poi)
	if err != nil {
		return nil, errors.Wrap(err, "unable to retrieve cookbook artifacts")
	}

	totalCookbooks := 0
	for _, versions := range results {
		totalCookbooks += len(versions.Versions)
	}
	totalCookbooks += len(resultsCBA)

	return &CookbooksReport{
		Records: make([]*CookbookRecord, 0, totalCookbooks),
		// We buffer this so that we don't block
		// if caller does not drain the queue as messages
		// arrive when we Run() the report:
		Progress:              make(chan int, totalCookbooks),
		TotalCookbooks:        totalCookbooks,
		RunCookstyle:          runCookstyle,
		NodeFilter:            nodeFilter,
		cookbooks:             cbi,
		cookbookArtifacts:     cbai,
		onlyUnused:            onlyUnused,
		searcher:              searcher,
		cookstyle:             NewCookstyleRunner(),
		cookbooksDir:          cookbooksDir,
		numWorkers:            workers,
		cookbookSearchResults: results,
		CBASearchResults:      resultsCBA,
		policyGroups:          pgi,
		Policies:              poi,
	}, nil
}

func (cbr *CookbooksReport) Generate() {

	var (
		downloadCh = make(chan cookbookItem, cbr.TotalCookbooks)
		analyzeCh  = make(chan *CookbookRecord, cbr.TotalCookbooks)
		doneCh     = make(chan bool)
	)

	// determine how many workers do we need, by default, the total number of cookbooks
	numWorkers := cbr.TotalCookbooks
	if cbr.TotalCookbooks > cbr.numWorkers {
		// but if the total number of cookbooks is major than
		// the maximum allowed, set it to the maximum
		numWorkers = cbr.numWorkers
	}

	// launch jobs that will be read by the workers (goroutines)
	go cbr.triggerJobs(downloadCh)

	// launch the download workers, these will process downloads and send them to the next
	// channel to be analyzed, once all messages have been read, it closes the download channel
	go cbr.createDownloadWorkerPool(numWorkers, downloadCh, analyzeCh)

	// launch the analyze workers, these will process analyzis by running cookstyle,
	// once all messages have been processed, it will send a single message to the done channel
	go cbr.createAnalyzeWorkerPool(numWorkers, analyzeCh, doneCh)

	// wait for a message in from the done channel
	// to make sure there are no more processes running
	<-doneCh
}

func (cbr *CookbooksReport) addRecord(r *CookbookRecord) {
	cbr.recordsMutex.Lock()
	defer cbr.recordsMutex.Unlock()
	cbr.Records = append(cbr.Records, r)
}

func (cbr *CookbooksReport) triggerJobs(inCh chan<- cookbookItem) {
	for cookbookName, cookbookVersions := range cbr.cookbookSearchResults {
		for _, ver := range cookbookVersions.Versions {
			inCh <- cookbookItem{Name: cookbookName, Version: ver.Version}
		}
	}
	for _, cbItem := range cbr.CBASearchResults {
		inCh <- cbItem
	}
	close(inCh)
}

func (cbr *CookbooksReport) createDownloadWorkerPool(nWorkers int, downloadCh <-chan cookbookItem, analyzeCh chan<- *CookbookRecord) {
	var wg sync.WaitGroup
	for i := 0; i < nWorkers; i++ {
		wg.Add(1)
		go func(inCh <-chan cookbookItem, outCh chan<- *CookbookRecord, wg *sync.WaitGroup) {
			for item := range inCh {
				if item.Policy != "" {
					cbState := cbr.downloadCookbookArtifact(item)
					analyzeCh <- cbState
				} else {
					cbState := cbr.downloadCookbook(item.Name, item.Version)
					analyzeCh <- cbState
				}
			}
			wg.Done()

		}(downloadCh, analyzeCh, &wg)
	}

	wg.Wait()
	close(analyzeCh)
}

func (cbr *CookbooksReport) createAnalyzeWorkerPool(nWorkers int, analyzeCh <-chan *CookbookRecord, doneCh chan<- bool) {
	var wg sync.WaitGroup
	defer close(cbr.Progress)

	for i := 0; i < nWorkers; i++ {
		wg.Add(1)
		go func(inCh <-chan *CookbookRecord, wg *sync.WaitGroup) {
			for record := range inCh {
				if record != nil {
					cbr.addRecord(record)
					if cbr.RunCookstyle {
						cbr.runCookstyleFor(record)
					}
				}
				// This is the end of the pipeline - once a work item
				// is done here, we can notify that it is complete
				cbr.Progress <- 1
			}
			wg.Done()
		}(analyzeCh, &wg)
	}

	wg.Wait()
	doneCh <- true
}

func (cbr *CookbooksReport) downloadCookbook(cookbookName, version string) *CookbookRecord {
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
		if len(nodes) > 0 {
			return nil
		}
	} else {
		if len(nodes) == 0 {
			return nil
		}
	}

	// only do the actual download if we need to analyze the cookbooks.
	if cbr.RunCookstyle {
		err = cbr.cookbooks.DownloadTo(cookbookName, version, cbr.cookbooksDir)
		if err != nil {
			cbState.DownloadError = errors.Wrapf(err, "unable to download cookbook %s", cookbookName)
		}
	}

	return cbState
}

func (cbr *CookbooksReport) downloadCookbookArtifact(item cookbookItem) *CookbookRecord {
	var (
		nodes, err       = cbr.nodesUsingPolicy(item.PolicyGroup, item.Policy, item.PolicyRev)
		cookbookLongName = fmt.Sprintf("%v-%v", item.Name, item.CBAIdentifier[0:20])
		cbState          = &CookbookRecord{
			Name:        item.Name,
			path:        filepath.Join(cbr.cookbooksDir, cookbookLongName),
			Identifier:  item.CBAIdentifier,
			Policy:      item.Policy,
			PolicyGroup: item.PolicyGroup,
			PolicyVer:   item.PolicyRev,
		}
	)
	if err != nil {
		cbState.UsageLookupError = err
	}
	cbState.Nodes = nodes
	// by default we report only cookbooks that are being used by one or more nodes,
	// but we also provide a way to report the opposite, that is, only unused cookbooks
	if cbr.onlyUnused {
		if len(nodes) > 0 {
			return nil
		}
	} else {
		if len(nodes) == 0 {
			return nil
		}
	}

	// only do the actual download if we need to analyze the cookbooks.
	if cbr.RunCookstyle {
		err = cbr.cookbookArtifacts.DownloadTo(item.Name, item.CBAIdentifier, cbr.cookbooksDir)
		if err != nil {
			cbState.DownloadError = errors.Wrapf(err, "unable to download cookbook %s", item.Name)
		}
	}

	return cbState
}

func (cbr *CookbooksReport) nodesUsingCookbookVersion(cookbook string, version string) ([]string, error) {
	query := map[string]interface{}{
		"name": []string{"name"},
	}

	// TODO add pagination
	nodeFilter := fmt.Sprintf("cookbooks_%s_version:%s", cookbook, version)
	if cbr.NodeFilter != "" {
		nodeFilter = fmt.Sprintf("%s AND %s", nodeFilter, cbr.NodeFilter)
	}

	pres, err := cbr.searcher.PartialExec("node", nodeFilter, query)
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

func (cbr *CookbooksReport) nodesUsingPolicy(policyGroup string, policyName string, policyRev string) ([]string, error) {

	query := map[string]interface{}{
		"name": []string{"name"},
	}

	nodeFilter := fmt.Sprintf("policy_name:%s AND policy_group:%s AND policy_revision:%s", policyName, policyGroup, policyRev)
	if cbr.NodeFilter != "" {
		nodeFilter = fmt.Sprintf("%s AND %s", nodeFilter, cbr.NodeFilter)
	}

	pres, err := cbr.searcher.PartialExec("node", nodeFilter, query)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get policy usage information for nodes")
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

func getCookbookArtifacts(policyGroups PolicyGroupInterface, Policies PolicyInterface) ([]cookbookItem, error) {

	cbaResults := []cookbookItem{}

	policyGroupList, err := policyGroups.List()
	if err != nil {
		return nil, errors.Wrap(err, "unable to retrieve policy groups")
	}

	for pg, v := range policyGroupList {
		for p, pv := range v.Policies {
			for _, rv := range pv {
				rvDetail, err := Policies.GetRevisionDetails(p, rv)
				if err != nil {
					return nil, errors.Wrap(err, "unable to retrieve cookbook artifacts for policy revisions")
				}
				for ck, cv := range rvDetail.CookbookLocks {
					cbaResults = append(cbaResults, cookbookItem{Name: ck,
						CBAIdentifier: cv.Identifier,
						Policy:        p,
						PolicyGroup:   pg,
						PolicyRev:     rv})
				}
			}
		}
	}

	return cbaResults, nil
}
