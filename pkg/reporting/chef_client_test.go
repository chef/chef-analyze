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

package reporting_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	subject "github.com/chef/chef-analyze/pkg/reporting"
)

func TestNewChefClientWithDefaults(t *testing.T) {
	defer os.RemoveAll(createCredentialsConfig(t))
	cfg, err := subject.NewDefault()
	if assert.Nil(t, err) {
		client, err := subject.NewChefClient(&cfg)
		assert.Nil(t, err)
		assert.NotNil(t, client)
	}
}

func TestNewChefClientWithouAnySettings(t *testing.T) {
	client, err := subject.NewChefClient(&subject.Reporting{})
	assert.Nil(t, client)
	if assert.NotNil(t, err) {
		// first error
		assert.Contains(t, err.Error(), "couldn't read key")
	}
}

func TestNewChefClientCredsWithEmptyKey(t *testing.T) {
	defer os.RemoveAll(createCredentialsConfig(t))

	cfg, err := subject.NewDefault()
	if assert.Nil(t, err) {
		assert.NotNil(t, cfg)

		// activate a profile that has an empty key pem
		err := cfg.SwitchProfile("empty")
		assert.Nil(t, err)

		client, err := subject.NewChefClient(&cfg)
		assert.Nil(t, client)
		if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), "unable to setup an API client")
			assert.Contains(t, err.Error(), "private key block size invalid")
		}
	}
}
