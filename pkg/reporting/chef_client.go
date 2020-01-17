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
)

// NewChefClient creates an instance with the loaded credentials
func NewChefClient(cfg *Reporting) (*chef.Client, error) {
	// read the client key
	key, err := ioutil.ReadFile(cfg.ClientKey)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't read key '%s'", cfg.ClientKey)
	}

	// create a chef client
	client, err := chef.NewClient(&chef.Config{
		Name: cfg.ClientName,
		Key:  string(key),
		// TODO @afiune fix this upstream, if you do not add a '/' at the end of th URL
		// the client will be malformed for any further request
		BaseURL: cfg.ChefServerUrl + "/",
		SkipSSL: cfg.NoSSLVerify,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to setup an API client")
	}

	return client, nil
}
