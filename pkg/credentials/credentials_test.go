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

package credentials_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	subject "github.com/chef/chef-analyze/pkg/credentials"
)

func TestFindCredentialsFileNotFound(t *testing.T) {
	credsFile, err := subject.FindCredentialsFile()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "file 'credentials' not found")
		assert.Equal(t, "", credsFile)
	}
}

func TestNewDefault(t *testing.T) {
	createCredentialsConfig(t)
	defer os.RemoveAll(".chef") // clean up

	creds, err := subject.NewDefault()
	if assert.Nil(t, err) {
		assert.Equal(t,
			subject.DefaultProfileName, creds.ActiveProfile(),
			"the default profile doesn't match")
		assert.Equal(t,
			"foo", creds.ClientName,
			"the default client_name doesn't match")
		assert.Equal(t,
			"foo.pem", creds.ClientKey,
			"the default client_key doesn't match")
		assert.Equal(t,
			"chef-server.example.com/organizations/bubu", creds.ChefServerUrl,
			"the default chef_server_url doesn't match")
	}
}

func TestNewDefaultOtherKindOfError(t *testing.T) {
	// save real HOME, empty it and restored it after this test runs
	savedHome := os.Getenv("HOME")
	os.Setenv("HOME", "")
	defer os.Setenv("HOME", savedHome)

	creds, err := subject.NewDefault()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "unable to detect home directory")
		assert.Contains(t, err.Error(), "$HOME is not defined")
		assert.Equal(t, subject.Credentials{}, creds, "should return an empty config")
	}
}

func TestNewErrorCredsNotFound(t *testing.T) {
	creds, err := subject.New("dev")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "credentials file not found")
		assert.Contains(t, err.Error(), "default: $HOME/.chef/credentials")
		assert.Contains(t, err.Error(), "setup your local credentials config by following this documentation")
		assert.Contains(t, err.Error(), "https://docs.chef.io/knife_setup.html#knife-profiles")
		assert.Equal(t, subject.Credentials{}, creds)
	}
}

// TODO @afiune re enable this test once we understand what is happening with
// our docker images. Issue: https://github.com/chef/chef-analyze/issues/56
func _TestNewPermissionsError(t *testing.T) {
	defer os.RemoveAll(".chef") // clean up
	if err := os.MkdirAll(".chef", os.ModePerm); err != nil {
		t.Fatal(err)
	}

	// a super secure file! no one can read, write or execute. ha!
	err := ioutil.WriteFile(".chef/credentials", []byte("[default]"), 0000)
	if err != nil {
		t.Fatal(err)
	}

	creds, err := subject.New("default")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "unable to read credentials from")
		assert.Contains(t, err.Error(), ".chef/credentials")
		assert.Contains(t, err.Error(), "permission denied")
		assert.Equal(t, subject.Credentials{}, creds)
	}
}

func TestNewMalformedError(t *testing.T) {
	createMalformedCredentialsConfig(t)
	defer os.RemoveAll(".chef") // clean up

	creds, err := subject.New("default")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "unable to parse credentials file.")
		assert.Contains(t, err.Error(), "verify the format of the credentials file by following this documentation")
		assert.Contains(t, err.Error(), "cannot load TOML value")
		assert.Equal(t, subject.Credentials{}, creds)
	}
}

func TestNew(t *testing.T) {
	createCredentialsConfig(t)
	defer os.RemoveAll(".chef") // clean up

	creds, err := subject.New("dev")
	if assert.Nil(t, err) {
		assert.Equal(t,
			"dev", creds.ActiveProfile(),
			"the active profile doesn't match")
		assert.Equal(t,
			"dev", creds.ClientName,
			"the active client_name doesn't match")
		assert.Equal(t,
			"dev.pem", creds.ClientKey,
			"the active client_key doesn't match")
		assert.Equal(t,
			"chef-server.example.com/organizations/dev", creds.ChefServerUrl,
			"the active chef_server_url doesn't match")
	}
}

func TestNewNoProfiles(t *testing.T) {
	if err := os.MkdirAll(".chef", os.ModePerm); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(".chef") // clean up

	err := ioutil.WriteFile(".chef/credentials", []byte(""), 0644)
	if err != nil {
		t.Fatal(err)
	}

	creds, err := subject.New("default")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "profile not found in credentials file")
		assert.Contains(t, err.Error(), "verify the format of the credentials file by following this documentation:")
		assert.Contains(t, err.Error(), "https://docs.chef.io/knife_setup.html#knife-profiles")
		assert.Equal(t, subject.Profiles{}, creds.Profiles)
	}
}

func TestNewWithSingleOverrides(t *testing.T) {
	createCredentialsConfig(t)
	defer os.RemoveAll(".chef") // clean up

	creds, err := subject.New("dev",
		func(c *subject.Credentials) {
			c.ClientName = "caramelo"
		},
	)
	if assert.Nil(t, err) {
		assert.Equal(t,
			"dev", creds.ActiveProfile(),
			"the active profile doesn't match")
		assert.Equal(t,
			"caramelo", creds.ClientName, // overridden above
			"the overridden client_name doesn't match")
		assert.Equal(t,
			"dev.pem", creds.ClientKey,
			"the active client_key doesn't match")
		assert.Equal(t,
			"chef-server.example.com/organizations/dev", creds.ChefServerUrl,
			"the active chef_server_url doesn't match")
	}
}

func TestNewWithMultipleOverrides(t *testing.T) {
	createCredentialsConfig(t)
	defer os.RemoveAll(".chef") // clean up

	overrides := []subject.OverrideFunc{
		// overrides client_key
		func(c *subject.Credentials) { c.ClientKey = "foo.pem" },
		// switches profile, which makes it all change
		func(c *subject.Credentials) { c.SwitchProfile("dev") },
		// overrides client_name
		func(c *subject.Credentials) { c.ClientName = "bubulubu" },
	}

	// loading default profile but change to dev from above override
	creds, err := subject.New("default", overrides...)
	if assert.Nil(t, err) {
		assert.Equal(t,
			"dev", creds.ActiveProfile(),
			"the active profile doesn't match")
		assert.Equal(t,
			"bubulubu", creds.ClientName,
			"the overridden client_name doesn't match")
		assert.Equal(t,
			"dev.pem", creds.ClientKey,
			"the active client_key doesn't match")
		assert.Equal(t,
			"chef-server.example.com/organizations/dev", creds.ChefServerUrl,
			"the active chef_server_url doesn't match")
	}
}

func TestConfigSwitchProfiles(t *testing.T) {
	createCredentialsConfig(t)
	defer os.RemoveAll(".chef") // clean up

	creds, err := subject.NewDefault()
	if assert.Nil(t, err) {
		assert.Equal(t,
			subject.DefaultProfileName, creds.ActiveProfile(),
			"the active profile doesn't match")
		assert.Equal(t,
			"foo", creds.ClientName,
			"the active client_name doesn't match")
		assert.Equal(t,
			"foo.pem", creds.ClientKey,
			"the active client_key doesn't match")
		assert.Equal(t,
			"chef-server.example.com/organizations/bubu", creds.ChefServerUrl,
			"the active chef_server_url doesn't match")
	}

	// Switching to an existing profile
	err = creds.SwitchProfile("dev")
	if assert.Nil(t, err) {
		assert.Equal(t,
			"dev", creds.ActiveProfile(),
			"the new active profile doesn't match")
		assert.Equal(t,
			"dev", creds.ClientName,
			"the new active client_name doesn't match")
		assert.Equal(t,
			"dev.pem", creds.ClientKey,
			"the new active client_key doesn't match")
		assert.Equal(t,
			"chef-server.example.com/organizations/dev", creds.ChefServerUrl,
			"the new active chef_server_url doesn't match")
	}
	// Switching to an existing profile
	err = creds.SwitchProfile("not-exist")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "profile not found in credentials file")
		assert.Equal(t,
			"dev", creds.ActiveProfile(),
			"the active profile shouldn't have changed")
	}
}

func TestFromViperError(t *testing.T) {
	creds, err := subject.FromViper("default")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "credentials file not found. (default: $HOME/.chef/credentials")
		assert.Equal(t, subject.Credentials{}, creds)
	}

	// These settings are set from the cmd/ Go package, note that
	// this file was not rendered on purpose to trigger an error
	viper.SetConfigFile(".chef/credentials")
	viper.SetConfigType("toml")

	creds, err = subject.FromViper("default")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "unable to read credentials from '.chef/credentials'")
		assert.Contains(t, err.Error(), "no such file or directory")
		assert.Equal(t, subject.Credentials{}, creds)
	}
}

func TestFromViperMalformedError(t *testing.T) {
	createMalformedCredentialsConfig(t)
	defer os.RemoveAll(".chef") // clean up

	// These settings are set from the cmd/ Go package
	viper.SetConfigFile(".chef/credentials")
	viper.SetConfigType("toml")

	creds, err := subject.FromViper("default")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "unable to parse credentials file.")
		assert.Contains(t, err.Error(), "verify the format of the credentials file by following this documentation")
		assert.Contains(t, err.Error(), "error(s) decoding")
		assert.Equal(t, subject.Credentials{}, creds)
	}
}

func TestFromViperNoProfiles(t *testing.T) {
	if err := os.MkdirAll(".chef", os.ModePerm); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(".chef") // clean up

	err := ioutil.WriteFile(".chef/credentials", []byte(""), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// These settings are set from the cmd/ Go package
	viper.SetConfigFile(".chef/credentials")
	viper.SetConfigType("toml")

	creds, err := subject.FromViper("default")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "profile not found in credentials file")
		assert.Contains(t, err.Error(), "verify the format of the credentials file by following this documentation:")
		assert.Contains(t, err.Error(), "https://docs.chef.io/knife_setup.html#knife-profiles")
		assert.Equal(t, subject.Profiles{}, creds.Profiles)
	}
}

func TestFromViper(t *testing.T) {
	createCredentialsConfig(t)
	defer os.RemoveAll(".chef") // clean up

	// These settings are set from the cmd/ Go package
	viper.SetConfigFile(".chef/credentials")
	viper.SetConfigType("toml")

	creds, err := subject.FromViper("default")
	if assert.Nil(t, err) {
		assert.Equal(t,
			subject.DefaultProfileName, creds.ActiveProfile(),
			"the default profile doesn't match")
		assert.Equal(t,
			"foo", creds.ClientName,
			"the default client_name doesn't match")
		assert.Equal(t,
			"foo.pem", creds.ClientKey,
			"the default client_key doesn't match")
		assert.Equal(t,
			"chef-server.example.com/organizations/bubu", creds.ChefServerUrl,
			"the default chef_server_url doesn't match")
	}
}

func TestFromViperWithMultipleOverrides(t *testing.T) {
	createCredentialsConfig(t)
	defer os.RemoveAll(".chef") // clean up

	// These settings are set from the cmd/ Go package
	viper.SetConfigFile(".chef/credentials")
	viper.SetConfigType("toml")

	overrides := []subject.OverrideFunc{
		// overrides client_key
		func(c *subject.Credentials) { c.ClientKey = "bar.pem" },
		// overrides client_name
		func(c *subject.Credentials) { c.ClientName = "bubulubu" },
	}

	creds, err := subject.FromViper("default", overrides...)
	if assert.Nil(t, err) {
		assert.Equal(t,
			subject.DefaultProfileName, creds.ActiveProfile(),
			"the default profile doesn't match")
		assert.Equal(t,
			"bubulubu", creds.ClientName,
			"the overridden client_name doesn't match")
		assert.Equal(t,
			"bar.pem", creds.ClientKey,
			"the overridden client_key doesn't match")
		assert.Equal(t,
			"chef-server.example.com/organizations/bubu", creds.ChefServerUrl,
			"the default chef_server_url doesn't match")
	}
}

// when calling this function, make sure to add the defer clean up as the below example
//
// createCredentialsConfig(t)
// defer os.RemoveAll(".chef") // clean up
//
func createCredentialsConfig(t *testing.T) {
	if err := os.MkdirAll(".chef", os.ModePerm); err != nil {
		t.Fatal(err)
	}

	creds := []byte(`[default]
client_name = "foo"
client_key = "foo.pem"
chef_server_url = "chef-server.example.com/organizations/bubu"

[dev]
client_name = "dev"
client_key = "dev.pem"
chef_server_url = "chef-server.example.com/organizations/dev"
`)
	err := ioutil.WriteFile(".chef/credentials", creds, 0644)
	if err != nil {
		t.Fatal(err)
	}
}

// when calling this function, make sure to add the defer clean up as the below example
//
// createMalformedCredentialsConfig(t)
// defer os.RemoveAll(".chef") // clean up
//
func createMalformedCredentialsConfig(t *testing.T) {
	if err := os.MkdirAll(".chef", os.ModePerm); err != nil {
		t.Fatal(err)
	}

	// @afiune very intense typo to pop the error inside mapstructure
	creds := []byte(`[default]
client_name = true
client_key = [ "foo", "bar" ]
chef_server_url = "chef-server.example.com/organizations/bubu"
`)
	err := ioutil.WriteFile(".chef/credentials", creds, 0644)
	if err != nil {
		t.Fatal(err)
	}
}
