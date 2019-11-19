package reporting

import chef "github.com/chef/go-chef"

// These interfaces map to chef.Client.* implementations.
// Using an interface allows us inject our own mock implementations for testing
// up to the go-chef API boundary.

type CookbookInterface interface {
	ListAvailableVersions(numVersions string) (chef.CookbookListResult, error)
	DownloadTo(name, version, localDir string) error
}

type SearchInterface interface {
	PartialExec(idx, statement string, params map[string]interface{}) (res chef.SearchResult, err error)
}
