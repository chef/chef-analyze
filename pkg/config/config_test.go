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

package config_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	subject "github.com/chef/chef-analyze/pkg/config"
)

func TestNewConfigNotFoundError(t *testing.T) {
	cfg, err := subject.New()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "config.toml file not found. (default: $HOME/.chef-workstation/config.toml)")
		assert.Contains(t, err.Error(), "setup your local configuration file by following this documentation:")
		assert.Contains(t, err.Error(), "https://www.chef.sh/docs/reference/config/")
		assert.Equal(t, subject.Config{}, cfg, "should return an empty config")
	}
}

// TODO @afiune re enable this test once we understand what is happening with
// our docker images. Issue: https://github.com/chef/chef-analyze/issues/56
func _TestNewConfigUnableToReadPermissionsError(t *testing.T) {
	defer os.RemoveAll(subject.DefaultChefWorkstationDirectory) // clean up
	if err := os.MkdirAll(".chef-workstation", os.ModePerm); err != nil {
		t.Fatal(err)
	}

	// a super secure file! no one can read, write or execute. ha!
	err := ioutil.WriteFile(".chef-workstation/config.toml", []byte(""), 0000)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := subject.New()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "unable to read config.toml from")
		assert.Contains(t, err.Error(), ".chef-workstation/config.toml")
		assert.Contains(t, err.Error(), "permission denied")
		assert.Equal(t, subject.Config{}, cfg, "should return an empty config")
	}
}

func TestNewConfigOtherKindOfError(t *testing.T) {
	// save real HOME, empty it and restored it after this test runs
	savedHome := os.Getenv("HOME")
	os.Setenv("HOME", "")
	defer os.Setenv("HOME", savedHome)

	cfg, err := subject.New()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "unable to detect home directory")
		assert.Contains(t, err.Error(), "$HOME is not defined")
		assert.Equal(t, subject.Config{}, cfg, "should return an empty config")
	}
}

func TestNewMalformedConfigError(t *testing.T) {
	createUserConfigTomlMalformed(t)
	createAppConfigTomlMalformed(t)
	defer os.RemoveAll(subject.DefaultChefWorkstationDirectory) // clean up

	cfg, err := subject.New()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "unable to parse config.toml file.")
		assert.Contains(t, err.Error(), "verify the format of the configuration file by following this documentation")
		assert.Contains(t, err.Error(), "https://www.chef.sh/docs/reference/config/")
		assert.Contains(t, err.Error(), "(last key parsed 'chef.cookbook_repo_paths'): expected a comma or array terminator")
		assert.Equal(t, subject.Config{}, cfg, "should return an empty config")
	}
}

func TestNewWithNoAppConfig(t *testing.T) {
	createUserConfigToml(t)
	defer os.RemoveAll(subject.DefaultChefWorkstationDirectory) // clean up

	cfg, err := subject.New()
	if assert.Nil(t, err) {
		assertSameInformationThatDoesntChangeOften(t, &cfg)

		// test parsing the features settings
		assert.Equal(t, 3, len(cfg.Features), "features is not well parsed")
		assert.Equal(t, true, cfg.Features["foo"], "features is not well parsed")
		assert.Equal(t, true, cfg.Features["bar"], "features is not well parsed")
		assert.Equal(t, false, cfg.Features["xyz"], "features is not well parsed")
	}
}

func TestNewOverrideFuncs(t *testing.T) {
	createUserConfigToml(t)
	defer os.RemoveAll(subject.DefaultChefWorkstationDirectory) // clean up

	cfg, err := subject.New()
	if assert.Nil(t, err) {
		assert.Equal(t, true, cfg.Telemetry.Enable, "telemetry.enable is not well parsed")
		assert.Equal(t, false, cfg.Telemetry.Dev, "telemetry.dev is not well parsed")
		assert.Equal(t, "debug", cfg.Log.Level, "telemetry.dev is not well parsed")
		assert.Equal(t, "ssh", cfg.Connection.DefaultProtocol, "connection.default_protocol is not well parsed")
		assert.Equal(t, 3, len(cfg.Features), "features is not well parsed")
		assert.Equal(t, true, cfg.Features["foo"], "features is not well parsed")
	}

	cfg, err = subject.New(
		// switching telemetry off
		func(c *subject.Config) { c.Telemetry.Enable = false },
		// activating dev mode
		func(c *subject.Config) { c.Telemetry.Dev = true },
		// lower log level
		func(c *subject.Config) { c.Log.Level = "info" },
		// change default protocol
		func(c *subject.Config) { c.Connection.DefaultProtocol = "winrm" },
		// activate a new feature
		func(c *subject.Config) { c.Features["new"] = true },
		// deactivate a new feature
		func(c *subject.Config) { c.Features["foo"] = false },
	)
	if assert.Nil(t, err) {
		assert.Equal(t, false, cfg.Telemetry.Enable, "telemetry.enable was not overwritten")
		assert.Equal(t, true, cfg.Telemetry.Dev, "telemetry.dev was not overwritten")
		assert.Equal(t, "info", cfg.Log.Level, "telemetry.dev was not overwritten")
		assert.Equal(t, "winrm", cfg.Connection.DefaultProtocol, "connection.default_protocol was not overwritten")
		assert.Equal(t, 4, len(cfg.Features), "features was not overwritten")
		assert.Equal(t, false, cfg.Features["foo"], "features was not overwritten")
		assert.Equal(t, true, cfg.Features["new"], "features was not overwritten")
	}
}

func TestNewPlusWSAppConfig(t *testing.T) {
	createUserConfigToml(t)
	createAppConfigToml(t)
	defer os.RemoveAll(subject.DefaultChefWorkstationDirectory) // clean up

	cfg, err := subject.New()
	if assert.Nil(t, err) {
		// @afiune telemetry data is inverted inside the WS App Config (Tray) but it doesn't
		// matter since it is being overwritten by the User Config :nice:
		// NOTE: tests are the same, they don't change
		assertSameInformationThatDoesntChangeOften(t, &cfg)

		// test parsing of features settings
		// NOTE: @afiune here is where it gets interesting, the user has 3 features and the
		// app config has 2 new extra features, that is why the features here are 5
		assert.Equal(t, 5, len(cfg.Features), "features are not well parsed")
		assert.Equal(t, true, cfg.Features["foo"], "features are not well parsed")
		assert.Equal(t, true, cfg.Features["bar"], "features are not well parsed")
		assert.Equal(t, false, cfg.Features["xyz"], "features (from the app config) are not well parsed")
		assert.Equal(t, true, cfg.Features["abc"], "features (from the app config) are not well parsed")
		assert.Equal(t, true, cfg.Features["ws_app_feature"], "features (from the app config) are not well parsed")
	}
}

// this test is exactly the same as TestNewWithNoAppConfig()
func TestUser(t *testing.T) {
	createUserConfigToml(t)
	defer os.RemoveAll(subject.DefaultChefWorkstationDirectory) // clean up

	cfg, err := subject.User()
	if assert.Nil(t, err) {
		assertSameInformationThatDoesntChangeOften(t, &cfg)
		assert.Equal(t, 3, len(cfg.Features), "features is not well parsed")
		assert.Equal(t, true, cfg.Features["foo"], "features is not well parsed")
		assert.Equal(t, true, cfg.Features["bar"], "features is not well parsed")
		assert.Equal(t, false, cfg.Features["xyz"], "features is not well parsed")
	}
}

func TestApp(t *testing.T) {
	createAppConfigToml(t)
	defer os.RemoveAll(subject.DefaultChefWorkstationDirectory) // clean up

	cfg, err := subject.App()
	if assert.Nil(t, err) {
		// here we can't use assertSameInformationThatDoesntChangeOften(t, &cfg)
		// because mainly all the user config is not loaded, so this is only loading
		// the app config which is very short, only telemetry, updates and features

		// test parsing of telemetry settings
		assert.Equal(t, false, cfg.Telemetry.Enable, "telemetry.enable is not well parsed")
		assert.Equal(t, true, cfg.Telemetry.Dev, "telemetry.dev is not well parsed")

		// test parsing of updates settings
		assert.Equal(t, "stable", cfg.Updates.Channel, "updates.channel is not well parsed")

		// test parsing of features settings
		assert.Equal(t, 4, len(cfg.Features), "features is not well parsed")
		assert.Equal(t, true, cfg.Features["ws_app_feature"], "features is not well parsed")
		assert.Equal(t, true, cfg.Features["abc"], "features is not well parsed")
		assert.Equal(t, true, cfg.Features["xyz"], "features is not well parsed")
		assert.Equal(t, false, cfg.Features["foo"], "features is not well parsed")

		// NOTE: any other config should be empty, for example the data_collector config
		assert.Equal(t, "", cfg.DataCollector.Url, "data_collector.url is not well parsed")
		assert.Equal(t, "", cfg.DataCollector.Token, "data_collector.token is not well parsed")
	}
}

func TestUserOverrideFuncs(t *testing.T) {
	createUserConfigToml(t)
	defer os.RemoveAll(subject.DefaultChefWorkstationDirectory) // clean up

	cfg, err := subject.User()
	if assert.Nil(t, err) {
		assert.Equal(t, true, cfg.Telemetry.Enable, "telemetry.enable is not well parsed")
	}

	cfg, err = subject.User(
		// switching telemetry off
		func(c *subject.Config) { c.Telemetry.Enable = false },
	)
	if assert.Nil(t, err) {
		assert.Equal(t, false, cfg.Telemetry.Enable, "telemetry.enable was not overwritten")
	}
}

func TestAppOverrideFuncs(t *testing.T) {
	createAppConfigToml(t)
	defer os.RemoveAll(subject.DefaultChefWorkstationDirectory) // clean up

	cfg, err := subject.App()
	if assert.Nil(t, err) {
		assert.Equal(t, false, cfg.Telemetry.Enable, "telemetry.enable is not well parsed")
	}

	cfg, err = subject.App(
		// switching telemetry on
		func(c *subject.Config) { c.Telemetry.Enable = true },
	)
	if assert.Nil(t, err) {
		assert.Equal(t, true, cfg.Telemetry.Enable, "telemetry.enable was not overwritten")
	}
}

// not that we actually need this, but we might :wink:
func TestMergeAppConfig(t *testing.T) {
	createAppConfigToml(t)
	defer os.RemoveAll(subject.DefaultChefWorkstationDirectory) // clean up

	cfg := subject.Config{}
	// values that doesn't exist in the application config
	cfg.DataCollector.Url = "example.com"
	cfg.DataCollector.Token = "token"

	// value that is set in the application config to "stable"
	// we set it to current and the merge should override it
	cfg.Updates.Channel = "current"

	err := cfg.MergeAppConfig()

	if assert.Nil(t, err) {
		// test parsing of telemetry settings
		assert.Equal(t, false, cfg.Telemetry.Enable, "telemetry.enable is not well parsed")
		assert.Equal(t, true, cfg.Telemetry.Dev, "telemetry.dev is not well parsed")

		// test parsing of updates settings
		assert.Equal(t, "stable", cfg.Updates.Channel, "updates.channel is not well parsed")

		// test parsing of features settings
		assert.Equal(t, 4, len(cfg.Features), "features is not well parsed")
		assert.Equal(t, true, cfg.Features["ws_app_feature"], "features is not well parsed")
		assert.Equal(t, true, cfg.Features["abc"], "features is not well parsed")
		assert.Equal(t, true, cfg.Features["xyz"], "features is not well parsed")
		assert.Equal(t, false, cfg.Features["foo"], "features is not well parsed")

		// NOTE: any other config should be empty, for example the data_collector config
		assert.Equal(t, "example.com", cfg.DataCollector.Url, "data_collector.url is not well parsed")
		assert.Equal(t, "token", cfg.DataCollector.Token, "data_collector.token is not well parsed")
	}
}

func TestAppMalformedConfigError(t *testing.T) {
	createAppConfigTomlMalformed(t)
	defer os.RemoveAll(subject.DefaultChefWorkstationDirectory) // clean up

	cfg, err := subject.App()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "unable to parse .app-managed-config.toml file.")
		assert.Contains(t, err.Error(), "there must be a problem with the Chef Workstation App, verify the format of the configuration by following this documentation")
		assert.Contains(t, err.Error(), "https://www.chef.sh/docs/reference/config/")
		assert.Contains(t, err.Error(), "toml: cannot load TOML value of type string into a Go boolea")
	}
}

// when calling this function, make sure to add the defer clean up as the below example
// ```go
// createUserConfigToml(t)
// defer os.RemoveAll(subject.DefaultChefWorkstationDirectory) // clean up
// ```
func createUserConfigToml(t *testing.T) {
	if err := os.MkdirAll(subject.DefaultChefWorkstationDirectory, os.ModePerm); err != nil {
		t.Fatal(err)
	}

	var (
		cfgData = []byte(`
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
		userConfigFile = filepath.Join(
			subject.DefaultChefWorkstationDirectory,
			subject.DefaultChefWSUserConfigFile,
		)
	)
	err := ioutil.WriteFile(userConfigFile, cfgData, 0666)
	if err != nil {
		t.Fatal(err)
	}
}

// when calling this function, make sure to add the defer clean up as the below example
// ```go
// createUserConfigTomlMalformed(t)
// defer os.RemoveAll(subject.DefaultChefWorkstationDirectory) // clean up
// ```
func createUserConfigTomlMalformed(t *testing.T) {
	if err := os.MkdirAll(subject.DefaultChefWorkstationDirectory, os.ModePerm); err != nil {
		t.Fatal(err)
	}

	var (
		// @afiune the typo is inside the 'cookbook_repo_paths': missing a comma
		cfgData = []byte(`
[telemetry]
enable = true
dev = false

[chef]
trusted_certs_dir="/path/to/mytrustedcerts"
cookbook_repo_paths = [
  "/path/to/cookbooks"
  "/var/chef/cookbooks"
]

[updates]
enable = true
channel = "current"
interval_minutes = 60
`)
		userConfigFile = filepath.Join(
			subject.DefaultChefWorkstationDirectory,
			subject.DefaultChefWSUserConfigFile,
		)
	)
	err := ioutil.WriteFile(userConfigFile, cfgData, 0666)
	if err != nil {
		t.Fatal(err)
	}
}

// when calling this function, make sure to add the defer clean up as the below example
// ```go
// createAppConfigToml(t)
// defer os.RemoveAll(subject.DefaultChefWorkstationDirectory) // clean up
// ```
func createAppConfigToml(t *testing.T) {
	if err := os.MkdirAll(subject.DefaultChefWorkstationDirectory, os.ModePerm); err != nil {
		t.Fatal(err)
	}

	var (
		featuresList = []byte(`
[telemetry]
enable = false
dev = true

[updates]
channel = "stable"

[features]
ws_app_feature = true
abc = true
xyz = true
foo = false
`)
		appConfigFile = filepath.Join(
			subject.DefaultChefWorkstationDirectory,
			subject.DefaultChefWSAppConfigFile,
		)
	)
	err := ioutil.WriteFile(appConfigFile, featuresList, 0666)
	if err != nil {
		t.Fatal(err)
	}
}

// when calling this function, make sure to add the defer clean up as the below example
// ```go
// createAppConfigTomlMalformed(t)
// defer os.RemoveAll(subject.DefaultChefWorkstationDirectory) // clean up
// ```
func createAppConfigTomlMalformed(t *testing.T) {
	if err := os.MkdirAll(subject.DefaultChefWorkstationDirectory, os.ModePerm); err != nil {
		t.Fatal(err)
	}

	var (
		featuresList = []byte(`
[telemetry]
enable = false
dev = true

[updates]
channel = "stable"

[features]
foo = "wrong"
`)
		appConfigFile = filepath.Join(
			subject.DefaultChefWorkstationDirectory,
			subject.DefaultChefWSAppConfigFile,
		)
	)
	err := ioutil.WriteFile(appConfigFile, featuresList, 0666)
	if err != nil {
		t.Fatal(err)
	}
}

func assertSameInformationThatDoesntChangeOften(t *testing.T, cfg *subject.Config) {
	// test parsing of telemetry settings
	assert.Equal(t, true, cfg.Telemetry.Enable, "telemetry.enable is not well parsed")
	assert.Equal(t, false, cfg.Telemetry.Dev, "telemetry.dev is not well parsed")

	// test parsing of log settings
	assert.Equal(t, "debug", cfg.Log.Level, "log.level is not well parsed")
	assert.Equal(t, "/path/to/chef-workstation.log", cfg.Log.Location, "log.location is not well parsed")

	// test parsing of cache settings
	assert.Equal(t, "/path/to/.cache/chef-workstation", cfg.Cache.Path, "cache.path is not well parsed")

	// test parsing of connection settings
	assert.Equal(t, "ssh", cfg.Connection.DefaultProtocol, "connection.default_protocol is not well parsed")
	assert.Equal(t, "bubu", cfg.Connection.DefaultUser, "connection.default_user is not well parsed")
	assert.Equal(t, true, cfg.Connection.SSH.SSL, "connection.ssh.ssl is not well parsed")
	assert.Equal(t, true, cfg.Connection.SSH.SSLVerify, "connection.ssh.ssl_verify is not well parsed")
	assert.Equal(t, false, cfg.Connection.WinRM.SSL, "connection.winrm.ssl is not well parsed")
	assert.Equal(t, false, cfg.Connection.WinRM.SSLVerify, "connection.winrm.ssl_verify is not well parsed")

	// test parsing of chef settings
	assert.Equal(t, "/path/to/mytrustedcerts", cfg.Chef.TrustedCertsDir, "chef.trusted_certs_dir is not well parsed")
	assert.Equal(t, 2, len(cfg.Chef.CookbookRepoPaths), "chef.cookbook_repo_paths is not well parsed")
	assert.Contains(t, cfg.Chef.CookbookRepoPaths, "/var/chef/cookbooks", "chef.cookbook_repo_paths is not well parsed")
	assert.Contains(t, cfg.Chef.CookbookRepoPaths, "/path/to/cookbooks", "chef.cookbook_repo_paths is not well parsed")

	// test parsing of updates settings
	assert.Equal(t, true, cfg.Updates.Enable, "updates.enable is not well parsed")
	assert.Equal(t, "current", cfg.Updates.Channel, "updates.channel is not well parsed")
	assert.Equal(t, 60, cfg.Updates.IntervalMinutes, "updates.interval_minutes is not well parsed")

	// test parsing of data-collector settings
	assert.Equal(t, "https://1.1.1.1/data-collector/v0/", cfg.DataCollector.Url, "data_collector.url is not well parsed")
	assert.Equal(t, "ABCDEF0123456789", cfg.DataCollector.Token, "data_collector.token is not well parsed")

	// test parsing of dev settings
	assert.Equal(t, true, cfg.Dev.Spinner, "dev.spinner is not well parsed")
}
