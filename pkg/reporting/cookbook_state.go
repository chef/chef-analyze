package reporting

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"strconv"

	chef "github.com/chef/go-chef"
	"github.com/cheggaaa/pb/v3"
	"github.com/pkg/errors"
)

type CookbookStateRecord struct {
	Name            string
	Version         string
	Violations      int
	Autocorrectable int
	Nodes           []string
	path            string
}

type PartialCookbookInterface interface {
	ListAvailableVersions(numVersions string) (chef.CookbookListResult, error)
	DownloadTo(name, version, localDir string) error
}

func CookbookState(cfg *Reporting, cbi PartialCookbookInterface, searcher PartialSearchInterface) ([]*CookbookStateRecord, error) {
	fmt.Println("Getting available cookbooks.") // c <- ProgressUpdate(Event: COOKBOOK_FETCH)
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
	runCookstyle(cbStates, len(cbStates))

	return cbStates, nil
}
func downloadCookbooks(cbi PartialCookbookInterface, searcher PartialSearchInterface, cookbooks chef.CookbookListResult, numVersions int) ([]*CookbookStateRecord, error) {
	progress := pb.StartNew(numVersions) // Ultimately, progress tracking and output lives with caller.

	cbStates := make([]*CookbookStateRecord, 0, numVersions)

	// Second pass: let's pack them up into returnable form.
	for cookbookName, cookbookVersions := range cookbooks {
		for _, ver := range cookbookVersions.Versions {
			err := cbi.DownloadTo(cookbookName, ver.Version, "analyze-cache")
			progress.Increment()
			if err != nil {
				// c <- ProgressUpdate{Event: DOWNLOAD_FAILED, Error: err}
				numVersions -= 1 // We use this later to make progress accurate for our second pass
				defer log.Print(err)
			} else { // TODO - lets have cbi.DownloadTo can return the localDir/COOKBOOK-VERSION path
				//        since it doesn't put it directly where we tell it to . Alternatively, it
				//        can require the entire path from the caller and not try to append cb/version
				nodesUsing, _ := nodesUsingCookbookVersion(searcher, cookbookName, ver.Version)
				dlRecord := CookbookStateRecord{Name: cookbookName,
					path:    fmt.Sprintf("analyze-cache/%v-%v", cookbookName, ver.Version),
					Version: ver.Version,
					Nodes:   nodesUsing,
				}
				// Now we need to find which nodes are using this cookbook. We'll probably want
				// to separate this for paralellism - it's one REST API query per cookbook, and we don't need
				// the results until we're ready to return from this function.
				cbStates = append(cbStates, &dlRecord)
			}
		}
	}
	progress.Finish()
	return cbStates, nil
}

func runCookstyle(cbStates []*CookbookStateRecord, numVersions int) {
	progress := pb.StartNew(numVersions)
	for _, cb := range cbStates {
		cmd := exec.Command("/opt/chef-workstation/bin/cookstyle", "--require", "../../formatter.rb", "--format", "Local::CorrectableCountFormatter")
		progress.Increment()
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
