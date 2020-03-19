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

package reporting_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	// store previous home directory
	oldHome := os.Getenv("HOME")

	// create a temporary directory and use it as a mocked home directory
	tmpHome, err := ioutil.TempDir("", "mock-home")
	if err != nil {
		panic("unable to create a temporary directory")
	}
	os.Setenv("HOME", tmpHome)

	// run each test
	exitCode := m.Run()

	// setup the previous home directory
	os.Setenv("HOME", oldHome)

	// remove the temporary home directory
	os.RemoveAll(tmpHome)

	os.Exit(exitCode)
}

// when calling this function, make sure to wrap it inside a defer statement
// ```go
// defer os.RemoveAll(createCredentialsConfig(t))
// ```
func createCredentialsConfig(t *testing.T) string {

	// for security to NOT delete a real .chef/ directory on our home directories
	// lets verify if the directory exist, if so, panic and inform the developer
	var (
		home     = os.Getenv("HOME")
		homeChef = filepath.Join(home, ".chef")
	)
	if _, err := os.Stat(homeChef); err == nil {
		// Oh oh, exists!
		panic(`

!!! ATTENTION !!!

You might be using the func createCredentialsConfig() without mocking the
HOME directory inside the Go tests. Please, verify that the test package
has a TestMain() func that mocks the HOME directory. (look at setup_test.go)

    `)
	}

	if err := os.MkdirAll(homeChef, os.ModePerm); err != nil {
		t.Fatal(err)
	}

	var (
		credsPath    = filepath.Join(homeChef, "credentials")
		fooPemPath   = filepath.Join(homeChef, "foo.pem")
		devPemPath   = filepath.Join(homeChef, "dev.pem")
		emptyPemPath = filepath.Join(homeChef, "empty.pem")
		creds        = []byte(`[default]
client_name = "foo"
client_key = "` + fooPemPath + `"
chef_server_url = "chef-server.example.com/organizations/bubu"

[dev]
client_name = "dev"
client_key = "` + devPemPath + `"
chef_server_url = "chef-server.example.com/organizations/dev"

[empty]
client_name = "empty"
client_key = "` + emptyPemPath + `"
chef_server_url = "chef-server.example.com/organizations/empty"

[inlined]
client_name = "inlined"
client_key ="""
` + string(key()) + `
"""
chef_server_url = "chef-server.example.com/organizations/inlined"
`)
	)
	err := ioutil.WriteFile(credsPath, creds, 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(fooPemPath, key(), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(devPemPath, key(), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(emptyPemPath, []byte(""), 0644)
	if err != nil {
		t.Fatal(err)
	}

	return homeChef
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

// when calling this function, make sure to wrap it inside a defer statement
// ```go
// defer os.RemoveAll(createConfigToml(t))
// ```
func createConfigToml(t *testing.T) string {

	// for security to NOT delete a real .chef-workstation/ directory on our
	// home directories lets verify if the directory exist, if so, panic and
	// inform the developer
	var (
		home         = os.Getenv("HOME")
		homeWS       = filepath.Join(home, ".chef-workstation")
		wsConfigToml = filepath.Join(homeWS, "config.toml")
	)
	if _, err := os.Stat(homeWS); err == nil {
		// Oh oh, exists!
		panic(`

!!! ATTENTION !!!

You might be using the func createConfigToml() without mocking the
HOME directory inside the Go tests. Please, verify that the test package
has a TestMain() func that mocks the HOME directory. (look at setup_test.go)

    `)
	}

	if err := os.MkdirAll(homeWS, os.ModePerm); err != nil {
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
	err := ioutil.WriteFile(wsConfigToml, creds, 0666)
	if err != nil {
		t.Fatal(err)
	}

	return homeWS
}

// setups up the binstubs directory at the front of the PATH environment
// environment and returns the previous value so that callers can set
// it back to its original value
//
// example:
//
// ```go
// savedPath := setupBinstubsDir()
// defer os.Setenv("PATH", savedPath)
// ```
func setupBinstubsDir() string {
	pathEnv := os.Getenv("PATH")
	newPathEnv := binstubsDir() + ":" + pathEnv
	os.Setenv("PATH", newPathEnv)
	return pathEnv
}

// returns the binstubs/ directory path
func binstubsDir() string {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return filepath.Join(pwd, "..", "..", "binstubs")
}
