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

package cmd

//
// The intend of this file is to have a single place where we can easily
// visualize the list of all error messages that we present to users.
//

const (
	// MissingMinimumParametersErr user feedback
	MissingMinimumParametersErr = `
  E001

  there are missing parameters for this tool to work, provide a credentials file:
    --credentials string

  or the following required flags:
    --client_key string
    --client_name string
    --chef_server_url string

`

	RepositoryAlreadyExistsE002 = `
  E002

  The repository already exists in %s.

  To re-run capture for node %s, delete this directory.

`
	// Arg: 1 - feature flag name
	//      2 - env var name
	//      3 - feature flag name
	AnalyzeNotEnabledE003 = `
  E003

	'%s' is experimental and in development. You can temporarily
	enable it by setting the following environment variable:

	  %s=true

	Or permanently enable it by modifying $HOME/.chef-workstation/config.toml:

	  [features]
		%s = true
`
)
