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

package credentials

import (
	"io/ioutil"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/chef/chef-analyze/pkg/config"
)

const (
	DefaultProfileName = "default"
	DefaultFileName    = "credentials"
)

// the credentials from a single profile
type CredsDetail struct {
	ChefServerUrl string `toml:"chef_server_url" mapstructure:"chef_server_url"`
	ClientName    string `toml:"client_name" mapstructure:"client_name"`
	ClientKey     string `toml:"client_key" mapstructure:"client_key"`
}

// a list of credentials, this type maps exactly as our .chef/credentials file
//
// example:
//
// [default]
// client_name = "foo"
// client_key = "foo.pem"
// chef_server_url = "chef-server.example.com/organizations/bubu"
//
// [dev]
// client_name = "dev"
// client_key = "dev.pem"
// chef_server_url = "chef-server.example.com/organizations/dev"
//
type Profiles map[string]CredsDetail

// main struct that holds all profiles, active profile and credentials
type Credentials struct {
	// list of all profiles available
	Profiles
	// active profile (private)
	// @afiune to change the active profile use SwitchProfile("dev")
	profile string
	// active credentials detail from the active profile
	CredsDetail
}

// returns the active profile
func (c *Credentials) ActiveProfile() string {
	return c.profile
}

// switch the active profile inside the credentials struct
func (c *Credentials) SwitchProfile(name string) error {
	if len(c.Profiles) == 0 {
		return errors.New(ProfileNotFoundErr)
	}
	if creds, ok := c.Profiles[name]; ok {
		c.CredsDetail = creds
		c.profile = name
		return nil
	}
	return errors.New(ProfileNotFoundErr)
}

// override functions can be passed to credentials.FromViper to override
// any particular setting
type OverrideFunc func(*Credentials)

// returns a Credentials instance from the current viper config
func FromViper(profile string, overrides ...OverrideFunc) (Credentials, error) {
	if viper.ConfigFileUsed() == "" {
		return Credentials{}, errors.New(CredentialsNotFoundErr)
	}

	if err := viper.ReadInConfig(); err != nil {
		//debug("Using config file:", viper.ConfigFileUsed())
		return Credentials{}, errors.Wrapf(err, "unable to read credentials from '%s'", viper.ConfigFileUsed())
	}

	// TODO @afiune add a mechanism to log debug messages
	//debug("using config file:", viper.ConfigFileUsed())

	profiles := Profiles{}
	// Unmarshall the viper config
	if err := viper.Unmarshal(&profiles); err != nil {
		return Credentials{}, errors.Wrap(err, MalformedCredentialsFileErr)
	}

	creds := Credentials{Profiles: profiles}
	if err := creds.SwitchProfile(profile); err != nil {
		return creds, err
	}

	for _, f := range overrides {
		f(&creds)
	}

	return creds, nil
}

// returns a Credentials instance using the default profile
func NewDefault() (Credentials, error) {
	return New(DefaultProfileName)
}

// returns a Credentials instance using the provided profile
func New(profile string, overrides ...OverrideFunc) (Credentials, error) {
	creds := Credentials{}

	credsFile, err := FindCredentialsFile()
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return creds, errors.New(CredentialsNotFoundErr)
		}
		return creds, err
	}

	credsBytes, err := ioutil.ReadFile(credsFile)
	if err != nil {
		return creds, errors.Wrapf(err, "unable to read credentials from '%s'", credsFile)
	}

	profiles := Profiles{}
	if _, err := toml.Decode(string(credsBytes), &profiles); err != nil {
		return creds, errors.Wrap(err, MalformedCredentialsFileErr)
	}

	creds.Profiles = profiles
	if err := creds.SwitchProfile(profile); err != nil {
		return creds, err
	}

	for _, f := range overrides {
		f(&creds)
	}

	return creds, nil
}

// finds the credentials file (default .chef/credentials) inside the current
// directory and recursively, plus inside the $HOME directory
func FindCredentialsFile() (string, error) {
	return config.FindConfigFile(DefaultFileName)
}
