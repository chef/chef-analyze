package config

import (
	"io/ioutil"
	"os"
	"path"

	chef "github.com/chef/go-chef"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	DefaultProfileName         = "default"
	DefaultChefDirectory       = ".chef"
	DefaultCredentialsFileName = "credentials"
)

type Credentials struct {
	ChefServerUrl string `toml:"chef_server_url" mapstructure:"chef_server_url"`
	ClientName    string `toml:"client_name" mapstructure:"client_name"`
	ClientKey     string `toml:"client_key" mapstructure:"client_key"`
}

type Profiles map[string]Credentials

type Config struct {
	Credentials
	Profile    string
	ChefClient *chef.Client
}

// creates a new Chef Client instance by stablishing a connection
// with the loaded Config
func (c *Config) CreateClient() error {
	// read the client key
	key, err := ioutil.ReadFile(c.ClientKey)
	if err != nil {
		return errors.Wrapf(err, "couldn't read key '%s'", c.ClientKey)
	}

	// create a chef client
	client, err := chef.NewClient(&chef.Config{
		Name: c.ClientName,
		Key:  string(key),
		// TODO @afiune fix this upstream, if you do not add a '/' at the end of th URL
		// the client will be malformed for any further request
		BaseURL: c.ChefServerUrl + "/",
	})
	if err != nil {
		return errors.Wrap(err, "unable setup a Chef client")
	}

	// store client inside the config for reusability
	c.ChefClient = client

	return nil
}

// overrides the active credentials from the viper bound flags,
// @afiune it is safe to call this function multiple times
func (c *Config) OverrideCredentialsFromViper() {
	if viper.GetString("client_name") != "" {
		c.ClientName = viper.GetString("client_name")
	}
	if viper.GetString("client_key") != "" {
		c.ClientKey = viper.GetString("client_key")
	}
	if viper.GetString("chef_server_url") != "" {
		c.ChefServerUrl = viper.GetString("chef_server_url")
	}
}

func Default() *Config {
	return &Config{
		Credentials: Credentials{},
		Profile:     DefaultProfileName,
	}
}

// returns a Config instance from the current viper config
func FromViper() (*Config, error) {
	var cfg *Config
	if viper.GetString("client_name") != "" &&
		viper.GetString("client_key") != "" &&
		viper.GetString("chef_server_url") != "" {
		cfg = Default()
		cfg.OverrideCredentialsFromViper()
	} else {

		if viper.ConfigFileUsed() == "" {
			return nil, errors.New(CredentialsNotFoundErr)
		}

		if err := viper.ReadInConfig(); err != nil {
			//debug("Using config file:", viper.ConfigFileUsed())
			return nil, errors.Wrapf(err, "unable to read config from '%s'", viper.ConfigFileUsed())
		}

		// TODO @afiune add a mechanism to log debug messages
		//debug("using config file:", viper.ConfigFileUsed())

		creds := Credentials{}
		// Unmarshall the viper config into the server Config
		// By default we support only the 'default' profile
		// TODO Add a profile flag just like knife https://docs.chef.io/knife_setup.html#knife-profiles
		if err := viper.UnmarshalKey(DefaultProfileName, &creds); err != nil {
			return nil, errors.Wrap(err, MalformedCredentialsFileErr)
		}

		cfg = &Config{
			Credentials: creds,
			Profile:     DefaultProfileName,
		}
		cfg.OverrideCredentialsFromViper()
	}

	if err := cfg.CreateClient(); err != nil {
		// this error is already wrapped with useful information
		return cfg, err
	}

	return cfg, nil
}

// find the configuration file (default .chef/credentials) inside the current
// directory and recursively, plus inside the $HOME directory
func FindConfigFile() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "unable to detect current directory")
	}

	var previous string
	for {
		// @afiune do we want to find other files like knife.rb?
		credsFile := path.Join(cwd, DefaultChefDirectory, DefaultCredentialsFileName)
		if credsFile == previous {
			break
		}
		if _, err := os.Stat(credsFile); err == nil {
			// creds file found
			return credsFile, nil
		}

		// save the creds file as previous, why? we need a way
		// to stop the for loop from going in an infinite loop
		previous = credsFile
		// go down the directory tree
		cwd = path.Join(cwd, "..")
	}

	// if we were unable to find the config file inside the current
	// directory and recursively, then try inside the home directory
	// @afiune what about using github.com/mitchellh/go-homedir
	// => homedir.Dir()
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "unable to detect home directory")
	}
	credsFile := path.Join(home, DefaultChefDirectory, DefaultCredentialsFileName)
	if _, err := os.Stat(credsFile); err == nil {
		// creds file found
		return credsFile, nil
	}

	// @afiune tell the user the paths we tried to find it?
	return "", errors.New(CredentialsNotFoundErr)
}
