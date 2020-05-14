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

import chef "github.com/chef/go-chef"

// These interfaces map to chef.Client.* implementations.
// Using an interface allows us inject our own mock implementations for testing
// up to the go-chef API boundary.

// CookbookInterface for testing
type CookbookInterface interface {
	ListAvailableVersions(numVersions string) (chef.CookbookListResult, error)
	DownloadTo(name, version, localDir string) error
}

type CBAInterface interface {
	List() (chef.CBAGetResponse, error)
	DownloadTo(name, id, localDir string) error
}

type PolicyGroupInterface interface {
	List() (chef.PolicyGroupGetResponse, error)
}

type PolicyInterface interface {
	GetRevisionDetails(policyName string, revisionID string) (chef.RevisionDetailsResponse, error)
}

// SearchInterface for testing
type SearchInterface interface {
	PartialExec(idx, statement string, params map[string]interface{}) (res chef.SearchResult, err error)
}

type NodesInterface interface {
	Get(name string) (chef.Node, error)
}
type RolesInterface interface {
	Get(name string) (*chef.Role, error)
}
type EnvironmentInterface interface {
	Get(name string) (*chef.Environment, error)
}
