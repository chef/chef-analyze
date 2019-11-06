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

package config

import (
	"io/ioutil"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

// abstraction of the chef-workstation configuration file (config.toml)
type Config struct {
	Telemetry     telemetrySettings `toml:"telemetry" mapstructure:"telemetry"`
	Log           logSettings       `toml:"log" mapstructure:"log"`
	Cache         cacheSettings     `toml:"cache" mapstructure:"cache"`
	Chef          chefSettings      `toml:"chef" mapstructure:"chef"`
	Updates       updatesSettings   `toml:"updates" mapstructure:"updates"`
	DataCollector dcSettings        `toml:"data_collector" mapstructure:"data_collector"`
	Connection    connSettings      `toml:"connection" mapstructure:"connection"`
	Features      map[string]bool   `toml:"features" mapstructure:"features"`
	Dev           devSettings       `toml:"dev" mapstructure:"dev"`
}

type telemetrySettings struct {
	Enable bool `toml:"enable" mapstructure:"enable"`
	Dev    bool `toml:"dev" mapstructure:"dev"`
}

type logSettings struct {
	Level    string `toml:"level" mapstructure:"level"`
	Location string `toml:"location" mapstructure:"location"`
}

type cacheSettings struct {
	Path string `toml:"path" mapstructure:"path"`
}

type chefSettings struct {
	TrustedCertsDir   string   `toml:"trusted_certs_dir" mapstructure:"trusted_certs_dir"`
	CookbookRepoPaths []string `toml:"cookbook_repo_paths" mapstructure:"cookbook_repo_paths"`
}

type updatesSettings struct {
	Enable          bool   `toml:"enable" mapstructure:"enable"`
	Channel         string `toml:"channel" mapstructure:"channel"`
	IntervalMinutes int    `toml:"interval_minutes" mapstructure:"interval_minutes"`
}

type dcSettings struct {
	Url   string `toml:"url" mapstructure:"url"`
	Token string `toml:"token" mapstructure:"token"`
}

type connSettings struct {
	DefaultProtocol string           `toml:"default_protocol" mapstructure:"default_protocol"`
	DefaultUser     string           `toml:"default_user" mapstructure:"default_user"`
	WinRM           protocolSettings `toml:"winrm" mapstructure:"winrm"`
	SSH             protocolSettings `toml:"ssh" mapstructure:"ssh"`
}

type protocolSettings struct {
	SSL       bool `toml:"ssl" mapstructure:"ssl"`
	SSLVerify bool `toml:"ssl_verify" mapstructure:"ssl_verify"`
}

type devSettings struct {
	Spinner bool `toml:"spinner" mapstructure:"spinner"`
}

// override functions to override any particular setting
type OverrideFunc func(*Config)

// returns a new Config instance
func New(overrides ...OverrideFunc) (Config, error) {
	cfg := Config{}

	configToml, err := FindChefWorkstationConfigFile()
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return cfg, errors.New(ConfigTomlNotFoundErr)
		}
		return cfg, err
	}

	configBytes, err := ioutil.ReadFile(configToml)
	if err != nil {
		return cfg, errors.Wrapf(err, "unable to read config.toml from '%s'", configToml)
	}

	if _, err := toml.Decode(string(configBytes), &cfg); err != nil {
		return cfg, errors.Wrap(err, MalformedConfigTomlFileErr)
	}

	for _, f := range overrides {
		f(&cfg)
	}

	return cfg, nil
}
