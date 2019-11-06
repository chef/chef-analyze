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
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	subject "github.com/chef/chef-analyze/pkg/reporting"
)

func TestReportingWithDefaults(t *testing.T) {
	createCredentialsConfig(t)
	defer os.RemoveAll(".chef") // clean up
	createConfigToml(t)
	defer os.RemoveAll(".chef-workstation") // clean up

	cfg, err := subject.NewDefault()
	if assert.Nil(t, err) {
		assert.Equal(t, "foo", cfg.ClientName)
		assert.Equal(t, ".chef/foo.pem", cfg.ClientKey)
		assert.Equal(t, "chef-server.example.com/organizations/bubu", cfg.ChefServerUrl)
		assert.Equal(t, false, cfg.NoSSLVerify)
		assert.Equal(t, true, cfg.Telemetry.Enable, "telemetry.enable is not well parsed")
		assert.Equal(t, false, cfg.Telemetry.Dev, "telemetry.dev is not well parsed")
		assert.Equal(t, "debug", cfg.Log.Level, "telemetry.dev is not well parsed")
		assert.Equal(t, "ssh", cfg.Connection.DefaultProtocol, "connection.default_protocol is not well parsed")
		assert.Equal(t, 3, len(cfg.Features), "features is not well parsed")
		assert.Equal(t, true, cfg.Features["foo"], "features is not well parsed")
	}
}

func TestReportingWirhOverrides(t *testing.T) {
	createCredentialsConfig(t)
	defer os.RemoveAll(".chef") // clean up

	cfg, err := subject.NewDefault(
		func(c *subject.Reporting) {
			c.NoSSLVerify = true
		},
	)
	if assert.Nil(t, err) {
		assert.Equal(t, "foo", cfg.ClientName)
		assert.Equal(t, ".chef/foo.pem", cfg.ClientKey)
		assert.Equal(t, "chef-server.example.com/organizations/bubu", cfg.ChefServerUrl)
		assert.Equal(t, true, cfg.NoSSLVerify)
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

// when calling this function, make sure to add the defer clean up as the below example
// ```go
// createCredentialsConfig(t)
// defer os.RemoveAll(".chef") // clean up
// ```
func createCredentialsConfig(t *testing.T) {
	if err := os.MkdirAll(".chef", os.ModePerm); err != nil {
		t.Fatal(err)
	}

	creds := []byte(`[default]
client_name = "foo"
client_key = ".chef/foo.pem"
chef_server_url = "chef-server.example.com/organizations/bubu"

[dev]
client_name = "dev"
client_key = ".chef/dev.pem"
chef_server_url = "chef-server.example.com/organizations/dev"
`)
	err := ioutil.WriteFile(".chef/credentials", creds, 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(".chef/foo.pem", key(), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(".chef/dev.pem", key(), 0644)
	if err != nil {
		t.Fatal(err)
	}
}

// mocked key
func key() []byte {
	return []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAuip6wlbkvuHvFl4m3Fzpqk3eSCBOudfubLdVZb8bWa0UM++d
fipPJadjDkrf1tlvKhjjoUF1XTizhI1xholJTRm92AjNi+rLWuCcn3Zih58xHTrl
S6+V8VHzmUB9iEDc7iyCdkXabDmoKhvu6CgagrFMN2maBjX7toaTOsAeGahVz3bv
inO/V61ea1an2TCgJ/Ovp77728kmPYQgQAceewelxgx/FwsroPVS0XFlBaRme8bg
Lc2VFJWeID27905ecPXQOZCfm5H2aCfpmzNIwHZ9WRpvOnpQxOXYYsimRGiAhwYA
zkC/FoCgny6ohMjhisibfHI6g7N+Y/iaWBImBwIDAQABAoIBAQCCS+sBi+mrw9wf
zqPqRclxXfC+kIYpQn1ob+SAQwJ0gFQMiZ+0Rw6ALyiAP11tNV+9mg/vtC3doirb
Elgrrni0UtjxlC+wxxOvNlfIsAYEICIy8B6+G1WZwh752w5BSAyZUmO5PejDKJOP
bV+H81GiuU671dhskmnrdUMksoQeteU0Xc1Uc8tYEAp7u/5KEA7WUCF5KiDvp5VC
5YSHC+slrIi0D+2xyWrPMFNqASN05Ke4yXhFEl9tuwiXUFSWahVlyY+mbsfE7xCX
GKPXjXpRFDIg/3EOdL0Cy1U6UUAFkYvYfU08UpQpxNA+ffgUcXbUfbLNcaCHfRl0
L/cPXEmhAoGBAO1/Q6M3sfNtyTMSaRoYxX+k8eV4A2qGy8B1rWv/oNV8NLAdbRBL
9E8QPk4TLSnTf9tK34JqGpUXHktYRvjFasB43V+LGhmrcflQ1YEeGodm3iiBZAuE
7KUX6x71/n3pHiOBea32hDBrGd2MLsX31Nwzo3e152JxLwZifpl48VsxAoGBAMir
cijumQLLIyB1urn1lQqtJm/Cd8jXypwkv6m0qD3bl5/shOEUTqnuGOnXJk2AcfNG
uv2TmP07gjz8JGTqInYgVHZrPd5Cu16TUl6TVezLu0Mu0yjpqlaSar4gTx1UT5ic
annHVa784VQnS/85WzVgldqvPSNsGhlDyfwkf9a3AoGBAJiUQmgByBmUVsaw9UUG
1RuEZMP/rnIp14z2DUxtFm8RNOhQf1kQ8ww4a07Nkx5j+qhwGdg3Qoy2JYhSVoZM
jqDJBa/0Nfh35Ok/vWsOZAzJUcDEH/omk8Ic87kYYT+THQHClOHmllZk+GEVRpd4
+Q/fPQ4Tl2vvOz7m2F7RDH6BAoGAS3oA5FhqANz7B1iAtTUjq/JYhKy2dTqFIJnJ
5UDoDuwraaGCkU4cEFpX0Ix2AayQL5qo9nuvjX/2io2j+rj94URjwG6xxImBBB+R
WbU9GmW+t5RDJB5PTWSg9YYde8Ccd6BNhCRvm/PNpONq+EJQhhEgDDLhYhNk9Z/D
tyzbUJ0CgYB9XAHT44ZbMVF6mF9egdx8HMQg0KgjUL8mAi5gvlbZjLyxpS0nnLhz
mjnO50tNmJ5OE3QGOEMr9Tzs6Sj7g7pv1McWtCpuvgAg8R4qL0a7MJ8qGSK29RP8
rfCdVg0GCAEjF9Dx+rdn4rWwP/8yJ5I61pWUDgmhKXyki94nPZt6Zg==
-----END RSA PRIVATE KEY-----`)
}

// when calling this function, make sure to add the defer clean up as the below example
// ```go
// createConfigToml(t)
// defer os.RemoveAll(".chef-workstation") // clean up
// ```
func createConfigToml(t *testing.T) {
	if err := os.MkdirAll(".chef-workstation", os.ModePerm); err != nil {
		t.Fatal(err)
	}

	creds := []byte(`
[telemetry]
enable = true
dev = false

[log]
level="debug"
location="/path/to/chef-workstation.log"

[cache]
path="/path/to/.cache/chef-workstation"

[connection]
default_protocol="ssh"
default_user="bubu"

[connection.ssh]
ssl=true
ssl_verify=true

[connection.winrm]
ssl=false
ssl_verify=false

[chef]
trusted_certs_dir="/path/to/mytrustedcerts"
cookbook_repo_paths = [
  "/path/to/cookbooks",
  "/var/chef/cookbooks"
]

[updates]
enable = true
channel = "current"
interval_minutes = 60

[data_collector]
url="https://1.1.1.1/data-collector/v0/"
token="ABCDEF0123456789"

[features]
foo = true
bar = true
xyz = false

[dev]
spinner = true
`)
	err := ioutil.WriteFile(".chef-workstation/config.toml", creds, 0666)
	if err != nil {
		t.Fatal(err)
	}
}
