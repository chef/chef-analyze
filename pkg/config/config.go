package config

import (
	"io/ioutil"
	"os"
	"path"

	chef "github.com/chef/go-chef"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// TODO @afiune be able to use .chef/credentials config

type Config struct {
	ChefServerUrl string `toml:"chef_server_url" mapstructure:"chef_server_url"`
	ClientName    string `toml:"client_name" mapstructure:"client_name"`
	ClientKey     string `toml:"client_key" mapstructure:"client_key"`
	ChefClient    *chef.Client
}

type Profiles map[string]Config

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
		return errors.Wrap(err, "Unable setup Chef client")
	}

	// store client inside the config for reusability
	c.ChefClient = client

	return nil
}

// FromViper returns a Config instance from the current viper config
func FromViper() (*Config, error) {
	if viper.ConfigFileUsed() == "" {
		return nil, errors.New(ConfigNotFoundErr)
	}

	if err := viper.ReadInConfig(); err != nil {
		//debug("Using config file:", viper.ConfigFileUsed())
		return nil, errors.Wrapf(err, "unable to read config from %s", viper.ConfigFileUsed())
	}

	// TODO @afiune add a mechanism to log debug messages
	//debug("using config file:", viper.ConfigFileUsed())

	cfg := &Config{}
	// Unmarshall the viper config into the server Config
	// By default we support only the 'default' profile
	// TODO Add a profile flag just like knife https://docs.chef.io/knife_setup.html#knife-profiles
	if err := viper.UnmarshalKey("default", cfg); err != nil {
		return nil, errors.Wrap(err, "failed to parse config")
	}

	if err := cfg.CreateClient(); err != nil {
		// this error is already wrapped with useful information
		return cfg, err
	}

	return cfg, nil
}

func FindConfigFile() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "unable to detect current directory")
	}

	var previous string
	for {
		// @afiune do we want to find other files like knife.rb?
		credsFile := path.Join(cwd, ".chef", "credentials")
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

	// Find config in home directory
	// @afiune what about using github.com/mitchellh/go-homedir
	// => homedir.Dir()
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "unable to detect home directory")
	}
	credsFile := path.Join(home, ".chef", "credentials")
	if _, err := os.Stat(credsFile); err == nil {
		// creds file found
		return credsFile, nil
	}

	// @afiune tell the user the paths we tried to find it?
	return "", errors.New("unable to locate config file")
}
