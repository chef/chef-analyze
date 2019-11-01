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

import (
	"io/ioutil"

	chef "github.com/chef/go-chef"
	"github.com/pkg/errors"

	"github.com/chef/chef-analyze/pkg/config"
)

// creates a new Chef Client instance by stablishing a connection
// with the loaded Config
func NewChefClient(c *config.Config) (*chef.Client, error) {
	// read the client key
	key, err := ioutil.ReadFile(c.ClientKey)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't read key '%s'", c.ClientKey)
	}

	// create a chef client
	client, err := chef.NewClient(&chef.Config{
		Name: c.ClientName,
		Key:  string(key),
		// TODO @afiune fix this upstream, if you do not add a '/' at the end of th URL
		// the client will be malformed for any further request
		BaseURL: c.ChefServerUrl + "/",
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable setup a Chef client")
	}

	return client, nil
}
