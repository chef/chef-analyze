package config

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	DefaultProfileName         = "default"
	DefaultChefDirectory       = ".chef"
	DefaultCredentialsFileName = "credentials"
)

// the credentials from a single profile
type Credentials struct {
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
type Profiles map[string]Credentials

// main config that holds all profiles, active profile and credentials
// plus a Chef API client to connect to a Chef Infra Server
type Config struct {
	// list of all profiles available
	Profiles
	// active profile (private)
	// @afiune to change the active profile use ActivateProfile("dev")
	profile string
	// active Credentials from the active profile
	Credentials
}

// switch the active profile inside the config
func (c *Config) ActivateProfile(name string) error {
	if len(c.Profiles) == 0 {
		return errors.New(ProfileNotFoundErr)
	}
	if creds, ok := c.Profiles[name]; ok {
		c.Credentials = creds
		c.profile = name
		return nil
	}
	return errors.New(ProfileNotFoundErr)
}

// override functions can be passed to config.FromViper to override
// any configuration setting
type OverrideFunc func(*Config)

// returns a Config instance from the current viper config
func FromViper(profile string, overrides ...OverrideFunc) (*Config, error) {
	if viper.ConfigFileUsed() == "" {
		return nil, errors.New(CredentialsNotFoundErr)
	}

	if err := viper.ReadInConfig(); err != nil {
		//debug("Using config file:", viper.ConfigFileUsed())
		return nil, errors.Wrapf(err, "unable to read config from '%s'", viper.ConfigFileUsed())
	}

	// TODO @afiune add a mechanism to log debug messages
	//debug("using config file:", viper.ConfigFileUsed())

	profiles := Profiles{}
	// Unmarshall the viper config
	if err := viper.Unmarshal(&profiles); err != nil {
		return nil, errors.Wrap(err, MalformedCredentialsFileErr)
	}

	cfg := &Config{Profiles: profiles}
	if err := cfg.ActivateProfile(profile); err != nil {
		return cfg, err
	}

	for _, f := range overrides {
		f(cfg)
	}

	return cfg, nil
}

// returns a Config instance using the default profile
func NewDefault() (*Config, error) {
	return New(DefaultProfileName)
}

// returns a Config instance using the provided profile
func New(profile string, overrides ...OverrideFunc) (*Config, error) {
	credsFile, err := FindCredentialsFile()
	if err != nil {
		// @afiune wrap it with useful message
		return nil, err
	}

	credsBytes, err := ioutil.ReadFile(credsFile)
	if err != nil {
		// @afiune wrap it with useful message
		return nil, err
	}

	profiles := Profiles{}
	if _, err := toml.Decode(string(credsBytes), &profiles); err != nil {
		// @afiune wrap it with useful message
		return nil, err
	}

	cfg := &Config{Profiles: profiles}
	if err := cfg.ActivateProfile(profile); err != nil {
		return cfg, err
	}

	for _, f := range overrides {
		f(cfg)
	}

	return cfg, nil
}

// finds the credentials file (default .chef/credentials) inside the current
// directory and recursively, plus inside the $HOME directory
func FindCredentialsFile() (string, error) {
	return FindConfigFile(DefaultCredentialsFileName)
}

// @afiune do we want to find other files like knife.rb and config.rb?
//func FindKnifeRbFile() (string, error) {
//return FindConfigFile("knife.rb")
//}
//func FindConfigRbFile() (string, error) {
//return FindConfigFile("config.rb")
//}

// finds the provided configuration file inside the current
// directory and recursively, plus inside the $HOME directory
func FindConfigFile(name string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "unable to detect current directory")
	}

	var previous string
	for {
		configFile := filepath.Join(cwd, DefaultChefDirectory, name)
		if configFile == previous {
			break
		}

		//debug("searching config in: %s\n", configFile)

		if _, err := os.Stat(configFile); err == nil {
			// config file found
			return configFile, nil
		}

		// save the config file as previous, why? we need a way
		// to stop the for loop from going in an infinite loop
		previous = configFile
		// go down the directory tree
		cwd = filepath.Join(cwd, "..")
	}

	// if we were unable to find the config file inside the current
	// directory and recursively, then try inside the home directory
	// @afiune what about using github.com/mitchellh/go-homedir
	// => homedir.Dir()
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "unable to detect home directory")
	}
	configFile := filepath.Join(home, DefaultChefDirectory, name)
	if _, err := os.Stat(configFile); err == nil {
		// config file found
		return configFile, nil
	}

	// @afiune tell the user the paths we tried to find it?
	return "", errors.New(CredentialsNotFoundErr)
}
