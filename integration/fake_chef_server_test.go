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

package integration

import (
	"fmt"
	"net/http"
)

// TODO @afiune populate this response with some real data
func nodeSearch(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "{}\n")
}

// TODO @afiune populate this response with some real data
func cookbooksList(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "{}\n")
}

// TODO @tball populate this response with some real data
func policyGroups(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "{}\n")
}

// start a HTTP server listening on localhost:80 to create a fake Chef Server
func startFakeChefServer() {
	// TODO @afiune I think we probably need to have a way to define the responses
	// that the mocked Chef Server will return, but for now we will hardcode them.
	// Maybe we can read a toml file with routes or something similar.
	http.HandleFunc(
		fmt.Sprintf("/organizations/%s/search/node", DefaultChefServerOrganization),
		nodeSearch,
	)
	http.HandleFunc(
		fmt.Sprintf("/organizations/%s/cookbooks", DefaultChefServerOrganization),
		cookbooksList,
	)
	http.HandleFunc(
		fmt.Sprintf("/organizations/%s/policy_groups", DefaultChefServerOrganization),
		policyGroups,
	)
	// @afiune we use port 80 to use "HTTP" instead of "HTTPS" to avoid signing requests
	http.ListenAndServe(":80", nil)
}

func init() {
	go startFakeChefServer()
}
