//
// Copyright 2020 Chef Software, Inc.
// Author: Marc Paradise <marc@chef.io>
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

// This file contains user-facing text used in command output.
// Do not modify leading/trailing newlines unless an issue in formatting
// is identified.
package cmd

const (
	// Param 1: cookbook directory in repository.
	// Param 2: newline-separated list of cookbooks
	CookbooksNotSourcedTxt = `Changes made to the following cookbooks in %s
cannot be saved upstream, though they can still be uploaded
to a Chef Server:

%s

`
	CookbookCaptureCompleteTxt = `You're ready to begin!

Start with 'kitchen converge'.  As you identify issues, you
can modify cookbooks in their original checkout locations or
in the repository's cookbooks directory and they will be picked
up on subsequent runs of 'kitchen converge'.
`
	// Param 1: repository directory
	// Param 2: newline-separated space-indented list of cookbooks
	CookbookCaptureGatherSourcesTxt = `
Repository has been created in '%s'.

Please clone or check out the following cookbooks locally
from their original sources, and provide the base path
for the checkout:

%s

If all cookbooks are not available in the same base location,
you will have a chance to provide additional locations.

If sources are not available for these cookbooks, leave this blank.

Checkout Location [none]: `

	// Param 1: newline-separated space-indented list of cookbooks
	CookbookCaptureRequestCookbookPathTxt = `
Please provide the base checkout path for the following
cookbooks, or leave blank if no more cookbooks are checked out:

%s

Checkout Location [none]: `
)
