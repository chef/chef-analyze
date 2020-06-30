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

func TestReportingWithDefaults(t *testing.T) {
	defer os.RemoveAll(createCredentialsConfig(t))
	defer os.RemoveAll(createConfigToml(t))

	cfg, err := subject.NewDefault()
	if assert.Nil(t, err) {
		assert.Equal(t, "foo", cfg.ClientName)
		// the key should contain .chef/foo.pem
		assert.Contains(t, cfg.ClientKey, "foo.pem")
		assert.Contains(t, cfg.ClientKey, ".chef")
		assert.Equal(t, "chef-server.example.com/organizations/bubu", cfg.ChefServerUrl)
		assert.Equal(t, false, cfg.NoSSLVerify)
		assert.Equal(t, true, cfg.Telemetry.Enable, "telemetry.enable is not well parsed")
		assert.Equal(t, false, cfg.Telemetry.Dev, "telemetry.dev is not well parsed")
		assert.Equal(t, "debug", cfg.Log.Level, "telemetry.dev is not well parsed")
		assert.Equal(t, "ssh", cfg.Connection.DefaultProtocol, "connection.default_protocol is not well parsed")
		assert.Equal(t, 3, len(cfg.Features), "features is not well parsed")
		assert.Equal(t, true, cfg.Features["foo"], "features is not well parsed")
		assert.Equal(t, true, cfg.Reports.Anonymize)
	}
}

func TestReportingWirhOverrides(t *testing.T) {
	defer os.RemoveAll(createCredentialsConfig(t))

	cfg, err := subject.NewDefault(
		func(c *subject.Reporting) {
			c.NoSSLVerify = true
			c.Reports.Anonymize = false
		},
	)
	if assert.Nil(t, err) {
		assert.Equal(t, "foo", cfg.ClientName)
		// the key should contain .chef/foo.pem
		assert.Contains(t, cfg.ClientKey, "foo.pem")
		assert.Contains(t, cfg.ClientKey, ".chef")
		assert.Equal(t, "chef-server.example.com/organizations/bubu", cfg.ChefServerUrl)
		assert.Equal(t, true, cfg.NoSSLVerify)
		assert.Equal(t, false, cfg.Reports.Anonymize)
	}
}

// NOTE: @afiune we don't report an error if we were unable to load the config.toml
func TestReportingNewDefaultErrorWithoutCredentials(t *testing.T) {
	cfg, err := subject.NewDefault()

	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "credentials file not found")
		assert.Contains(t, err.Error(), "default: $HOME/.chef/credentials")
		assert.Contains(t, err.Error(), "setup your local credentials config by following this documentation")
		assert.Contains(t, err.Error(), "https://docs.chef.io/knife_setup.html#knife-profiles")
		assert.Equal(t, subject.Reporting{}, cfg)
	}
}
