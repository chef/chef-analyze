package config_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	subject "github.com/chef/chef-analyze/pkg/config"
)

func TestFindConfigFileExists(t *testing.T) {
	if err := os.MkdirAll(".chef", os.ModePerm); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(".chef") // clean up

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	tmpfile, err := ioutil.TempFile(".chef", "my-config")
	if err != nil {
		t.Fatal(err)
	}

	var (
		tmpfileBase = filepath.Base(tmpfile.Name())
		tmpfileAbs  = filepath.Join(cwd, tmpfile.Name())
	)

	cfgFile, err := subject.FindConfigFile(tmpfileBase)
	if assert.Nil(t, err) {
		assert.Equal(t, tmpfileAbs, cfgFile)
	}
}

func TestFindConfigFileNotFound(t *testing.T) {
	cfgFile, err := subject.FindConfigFile("config.toml")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "file 'config.toml' not found")
		assert.Equal(t, "", cfgFile)
	}
}

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

	cfg, err := subject.NewDefault()
	if assert.Nil(t, err) {
		assert.Equal(t,
			subject.DefaultProfileName, cfg.ActiveProfile(),
			"the default profile doesn't match")
		assert.Equal(t,
			"foo", cfg.ClientName,
			"the default client_name doesn't match")
		assert.Equal(t,
			"foo.pem", cfg.ClientKey,
			"the default client_key doesn't match")
		assert.Equal(t,
			"chef-server.example.com/organizations/bubu", cfg.ChefServerUrl,
			"the default chef_server_url doesn't match")
	}
}

func TestNewErrorCredsNotFound(t *testing.T) {
	cfg, err := subject.New("dev")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "file 'credentials' not found")
		assert.Nil(t, cfg)
	}
}

func TestNew(t *testing.T) {
	createCredentialsConfig(t)
	defer os.RemoveAll(".chef") // clean up

	cfg, err := subject.New("dev")
	if assert.Nil(t, err) {
		assert.Equal(t,
			"dev", cfg.ActiveProfile(),
			"the active profile doesn't match")
		assert.Equal(t,
			"dev", cfg.ClientName,
			"the active client_name doesn't match")
		assert.Equal(t,
			"dev.pem", cfg.ClientKey,
			"the active client_key doesn't match")
		assert.Equal(t,
			"chef-server.example.com/organizations/dev", cfg.ChefServerUrl,
			"the active chef_server_url doesn't match")
	}
}

func TestNewWithSingleOverrides(t *testing.T) {
	createCredentialsConfig(t)
	defer os.RemoveAll(".chef") // clean up

	cfg, err := subject.New("dev",
		func(c *subject.Config) {
			c.ClientName = "caramelo"
		},
	)
	if assert.Nil(t, err) {
		assert.Equal(t,
			"dev", cfg.ActiveProfile(),
			"the active profile doesn't match")
		assert.Equal(t,
			"caramelo", cfg.ClientName, // overridden above
			"the overridden client_name doesn't match")
		assert.Equal(t,
			"dev.pem", cfg.ClientKey,
			"the active client_key doesn't match")
		assert.Equal(t,
			"chef-server.example.com/organizations/dev", cfg.ChefServerUrl,
			"the active chef_server_url doesn't match")
	}
}

func TestNewWithMultipleOverrides(t *testing.T) {
	createCredentialsConfig(t)
	defer os.RemoveAll(".chef") // clean up

	overrides := []subject.OverrideFunc{
		// overrides client_key
		func(c *subject.Config) { c.ClientKey = "foo.pem" },
		// switches profile, which makes it all change
		func(c *subject.Config) { c.SwitchProfile("dev") },
		// overrides client_name
		func(c *subject.Config) { c.ClientName = "bubulubu" },
	}

	// loading default profile but change to dev from above override
	cfg, err := subject.New("default", overrides...)
	if assert.Nil(t, err) {
		assert.Equal(t,
			"dev", cfg.ActiveProfile(),
			"the active profile doesn't match")
		assert.Equal(t,
			"bubulubu", cfg.ClientName,
			"the overridden client_name doesn't match")
		assert.Equal(t,
			"dev.pem", cfg.ClientKey,
			"the active client_key doesn't match")
		assert.Equal(t,
			"chef-server.example.com/organizations/dev", cfg.ChefServerUrl,
			"the active chef_server_url doesn't match")
	}
}

func TestConfigSwitchProfiles(t *testing.T) {
	createCredentialsConfig(t)
	defer os.RemoveAll(".chef") // clean up

	cfg, err := subject.NewDefault()
	if assert.Nil(t, err) {
		assert.Equal(t,
			subject.DefaultProfileName, cfg.ActiveProfile(),
			"the active profile doesn't match")
		assert.Equal(t,
			"foo", cfg.ClientName,
			"the active client_name doesn't match")
		assert.Equal(t,
			"foo.pem", cfg.ClientKey,
			"the active client_key doesn't match")
		assert.Equal(t,
			"chef-server.example.com/organizations/bubu", cfg.ChefServerUrl,
			"the active chef_server_url doesn't match")
	}

	// Switching to an existing profile
	err = cfg.SwitchProfile("dev")
	if assert.Nil(t, err) {
		assert.Equal(t,
			"dev", cfg.ActiveProfile(),
			"the new active profile doesn't match")
		assert.Equal(t,
			"dev", cfg.ClientName,
			"the new active client_name doesn't match")
		assert.Equal(t,
			"dev.pem", cfg.ClientKey,
			"the new active client_key doesn't match")
		assert.Equal(t,
			"chef-server.example.com/organizations/dev", cfg.ChefServerUrl,
			"the new active chef_server_url doesn't match")
	}
	// Switching to an existing profile
	err = cfg.SwitchProfile("not-exist")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "profile not found in credentials file")
		assert.Equal(t,
			"dev", cfg.ActiveProfile(),
			"the active profile shouldn't have changed")
	}
}

func TestFromViperError(t *testing.T) {
	cfg, err := subject.FromViper("default")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "credentials file not found. (default: $HOME/.chef/credentials")
		assert.Nil(t, cfg)
	}

	// These settings are set from the cmd/ Go package, note that
	// this file was not rendered on purpose to trigger an error
	viper.SetConfigFile(".chef/credentials")
	viper.SetConfigType("toml")

	t.Fatal("foo")
	cfg, err = subject.FromViper("default")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "unable to read config from '.chef/credentials'")
		assert.Contains(t, err.Error(), "no such file or directory")
		assert.Nil(t, cfg)
	}
}

func TestFromViper(t *testing.T) {
	createCredentialsConfig(t)
	defer os.RemoveAll(".chef") // clean up

	// These settings are set from the cmd/ Go package
	viper.SetConfigFile(".chef/credentials")
	viper.SetConfigType("toml")

	cfg, err := subject.FromViper("default")
	if assert.Nil(t, err) {
		assert.Equal(t,
			subject.DefaultProfileName, cfg.ActiveProfile(),
			"the default profile doesn't match")
		assert.Equal(t,
			"foo", cfg.ClientName,
			"the default client_name doesn't match")
		assert.Equal(t,
			"foo.pem", cfg.ClientKey,
			"the default client_key doesn't match")
		assert.Equal(t,
			"chef-server.example.com/organizations/bubu", cfg.ChefServerUrl,
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
		func(c *subject.Config) { c.ClientKey = "bar.pem" },
		// overrides client_name
		func(c *subject.Config) { c.ClientName = "bubulubu" },
	}

	cfg, err := subject.FromViper("default", overrides...)
	if assert.Nil(t, err) {
		assert.Equal(t,
			subject.DefaultProfileName, cfg.ActiveProfile(),
			"the default profile doesn't match")
		assert.Equal(t,
			"bubulubu", cfg.ClientName,
			"the overridden client_name doesn't match")
		assert.Equal(t,
			"bar.pem", cfg.ClientKey,
			"the overridden client_key doesn't match")
		assert.Equal(t,
			"chef-server.example.com/organizations/bubu", cfg.ChefServerUrl,
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
