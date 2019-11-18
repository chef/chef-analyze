package reporting

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"strconv"

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

type PartialCookbookInterface interface {
	ListAvailableVersions(numVersions string) (chef.CookbookListResult, error)
	DownloadTo(name, version, localDir string) error
}

func CookbookState(cfg *Reporting, cbi PartialCookbookInterface, searcher PartialSearchInterface) ([]*CookbookStateRecord, error) {
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
		for _, _ = range versions.Versions {
			numVersions += 1
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

func downloadCookbooks(cbi PartialCookbookInterface, searcher PartialSearchInterface, cookbooks chef.CookbookListResult, numVersions int) ([]*CookbookStateRecord, error) {
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
	for _, cb := range cbStates {
		progress.Increment()
		// If we could not download the complete cookbook, we can't provide
		// an accurate set of results
		if cb.CBDownloadError != nil {
			continue
		}

		cmd := exec.Command("/opt/chef-workstation/bin/cookstyle", "--require", formatterPath, "--format", "Local::CorrectableCountFormatter")
		cmd.Dir = cb.path
		// Note that err will often be non-nil because cookstyle will exit with exit code 1 when violations are found.
		// TODO allow that error through, and log the others
		out, _ := cmd.Output()
		returns := strings.Split(string(out), " ")
		cb.Violations, _ = strconv.Atoi(returns[0])
		cb.Autocorrectable, _ = strconv.Atoi(returns[1])
	}
	progress.Finish()
}
func nodesUsingCookbookVersion(searcher PartialSearchInterface, cookbook string, version string) ([]string, error) {
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
