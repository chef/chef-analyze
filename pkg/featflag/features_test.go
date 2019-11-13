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

package featflag_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/chef/chef-analyze/pkg/config"
	subject "github.com/chef/chef-analyze/pkg/featflag"
)

func TestDefaultFeatureFlags(t *testing.T) {
	assert.Equal(t,
		"(all:CHEF_FEAT_ALL) (analyze:CHEF_FEAT_ANALYZE)",
		subject.ListAll(),
		"default list of features doesn't match",
	)
}

func TestAllFeatureFlagsDisabledByDefault(t *testing.T) {
	assert.False(t, subject.ChefFeatAll.Enabled())
	assert.False(t, subject.ChefFeatAnalyze.Enabled())
}

func TestEnableChefAnalyzeThroughEnvVariable(t *testing.T) {
	defer os.Setenv("CHEF_FEAT_ANALYZE", "")

	if err := os.Setenv("CHEF_FEAT_ANALYZE", "true"); err != nil {
		t.Fatal("unable to set environment variable")
	}
	assert.True(t, subject.ChefFeatAnalyze.Enabled())
}

func TestEnableChefAnalyzeThroughGlobalChefAllFeatureFlag(t *testing.T) {
	defer os.Setenv("CHEF_FEAT_ALL", "")

	if err := os.Setenv("CHEF_FEAT_ALL", "true"); err != nil {
		t.Fatal("unable to set environment variable")
	}
	assert.True(t, subject.ChefFeatAnalyze.Enabled())
}

func TestAddFooFeatureFlag(t *testing.T) {
	// verify the list of feature flags before adding a new one
	assert.Equal(t,
		"(all:CHEF_FEAT_ALL) (analyze:CHEF_FEAT_ANALYZE)",
		subject.ListAll(),
		"default list of features doesn't match",
	)

	// defining a new Feature Flag Foo
	var chefFeatFoo = subject.New("CHEF_FEAT_FOO", "foo")

	assert.Equal(t, "CHEF_FEAT_FOO", chefFeatFoo.Env())
	assert.Equal(t, "foo", chefFeatFoo.Key())
	assert.False(t, chefFeatFoo.Enabled())
	assert.Equal(t,
		"(all:CHEF_FEAT_ALL) (analyze:CHEF_FEAT_ANALYZE) (foo:CHEF_FEAT_FOO)",
		subject.ListAll(),
		"the new flag should be displayed",
	)

	defer os.Setenv("CHEF_FEAT_FOO", "")

	if err := os.Setenv("CHEF_FEAT_FOO", "1"); err != nil {
		t.Fatal("unable to set environment variable")
	}
	assert.True(t, chefFeatFoo.Enabled())
}

func TestProtectAddingAnExistingFeatureFlag(t *testing.T) {
	// the CHEF_FEAT_ALL is a protected environment variable,
	// if a user tries to set a new feature flag with it,
	// the library should return the already registered flag
	featFlag := subject.New("CHEF_FEAT_ALL", "bubu")
	assert.Equal(t, "CHEF_FEAT_ALL", featFlag.Env())
	assert.Equal(t, "all", featFlag.Key())
	assert.True(t,
		subject.ChefFeatAll.Equals(&featFlag),
		"the CHEF_FEAT_ALL feature flags should never change",
	)

	// the same goes for the global analyze feature flag
	featFlag = subject.New("MY_ANALYZE_FEATURE", "analyze")
	assert.Equal(t, "CHEF_FEAT_ANALYZE", featFlag.Env())
	assert.Equal(t, "analyze", featFlag.Key())
	assert.True(t,
		subject.ChefFeatAnalyze.Equals(&featFlag),
		"the CHEF_FEAT_ANALYZE feature flags should never change",
	)
}

//
// tests that involve modifying the config (config.toml)
//
func TestAccessingFlagsWithNilConfig(t *testing.T) {
	subject.LoadConfig(nil)
	assert.False(t, subject.ChefFeatAll.Enabled())
	assert.False(t, subject.ChefFeatAnalyze.Enabled())
}

func TestEnableFlagsFromUserConfigToml(t *testing.T) {
	// all flags should be disabled
	subject.LoadConfig(nil)
	assert.False(t,
		subject.ChefFeatAll.Enabled(),
		"global ALL feature flag should be disabled",
	)
	assert.False(t,
		subject.ChefFeatAnalyze.Enabled(),
		"global ANALYZE feature flag should be disabled",
	)
	// even new flags created locally
	var chefFeatXYZ = subject.New("CHEF_FEAT_XYZ", "xyz")
	assert.False(t,
		chefFeatXYZ.Enabled(),
		"local XYZ feature flag should be disabled",
	)

	// initializing a configuration file
	createUserConfigToml(t)
	defer os.RemoveAll(config.DefaultChefWorkstationDirectory) // clean up

	// loading the config to the featflag library
	customConfig, err := config.New()
	if err != nil {
		t.Fatal("unable to load config.toml, check why this happened", err.Error())
	}
	subject.LoadConfig(&customConfig)

	// now things start getting interesting
	assert.False(t,
		subject.ChefFeatAll.Enabled(),
		"global ALL feature flag should continue to be disabled",
	)
	assert.True(t, // now it is enabled! (from the config)
		subject.ChefFeatAnalyze.Enabled(),
		"global ANALYZE feature flag should be enabled from loaded config.toml",
	)
	assert.False(t,
		chefFeatXYZ.Enabled(),
		"local XYZ feature flag should continue to be disabled",
	)

	// since environment variables has more priorityby, by enabling the
	// global feature flag CHEF_FEAT_ALL all features should be enabled
	defer os.Setenv("CHEF_FEAT_ALL", "")
	if err := os.Setenv("CHEF_FEAT_ALL", "gimmeall"); err != nil {
		t.Fatal("unable to set environment variable")
	}

	assert.True(t,
		subject.ChefFeatAll.Enabled(),
		"global ALL feature flag should be now enabled",
	)
	assert.True(t, // now it is enabled! (from the config)
		subject.ChefFeatAnalyze.Enabled(),
		"global ANALYZE feature flag should continue to be enabled",
	)
	assert.True(t,
		chefFeatXYZ.Enabled(),
		"local XYZ feature flag should be now enabled",
	)
}

func TestEnableFlagsFromAppConfigToml(t *testing.T) {
	// all flags should be disabled
	subject.LoadConfig(nil)
	assert.False(t,
		subject.ChefFeatAll.Enabled(),
		"global ALL feature flag should be disabled",
	)
	assert.False(t,
		subject.ChefFeatAnalyze.Enabled(),
		"global ANALYZE feature flag should be disabled",
	)
	// even new flags created locally
	var chefFeatXYZ = subject.New("CHEF_FEAT_XYZ", "xyz")
	assert.False(t,
		chefFeatXYZ.Enabled(),
		"local XYZ feature flag should be disabled",
	)

	// initializing a configuration file
	createAppConfigToml(t)
	defer os.RemoveAll(config.DefaultChefWorkstationDirectory) // clean up

	// loading the config to the featflag library
	customConfig, err := config.New()
	if err != nil {
		t.Fatal("unable to load config.toml, check why this happened", err.Error())
	}
	subject.LoadConfig(&customConfig)

	// now things start getting interesting
	assert.False(t,
		subject.ChefFeatAll.Enabled(),
		"global ALL feature flag should continue to be disabled",
	)
	assert.True(t, // now it is enabled! (from the config)
		subject.ChefFeatAnalyze.Enabled(),
		"global ANALYZE feature flag should be enabled from loaded config.toml",
	)
	assert.False(t,
		chefFeatXYZ.Enabled(),
		"local XYZ feature flag should continue to be disabled",
	)

	// since environment variables has more priorityby, by enabling the
	// global feature flag CHEF_FEAT_ALL all features should be enabled
	defer os.Setenv("CHEF_FEAT_ALL", "")
	if err := os.Setenv("CHEF_FEAT_ALL", "gimmeall"); err != nil {
		t.Fatal("unable to set environment variable")
	}

	assert.True(t,
		subject.ChefFeatAll.Enabled(),
		"global ALL feature flag should be now enabled",
	)
	assert.True(t, // now it is enabled! (from the config)
		subject.ChefFeatAnalyze.Enabled(),
		"global ANALYZE feature flag should continue to be enabled",
	)
	assert.True(t,
		chefFeatXYZ.Enabled(),
		"local XYZ feature flag should be now enabled",
	)
}

// when calling this function, make sure to add the defer clean up as the below example
// ```go
// createUserConfigToml(t)
// defer os.RemoveAll(config.DefaultChefWorkstationDirectory) // clean up
// ```
func createUserConfigToml(t *testing.T) {
	if err := os.MkdirAll(config.DefaultChefWorkstationDirectory, os.ModePerm); err != nil {
		t.Fatal(err)
	}

	var (
		featuresList = []byte(`
[features]
all = false
analyze = true
xyz = false
`)
		userConfigFile = filepath.Join(
			config.DefaultChefWorkstationDirectory,
			config.DefaultChefWSUserConfigFile,
		)
	)
	err := ioutil.WriteFile(userConfigFile, featuresList, 0666)
	if err != nil {
		t.Fatal(err)
	}
}

// when calling this function, make sure to add the defer clean up as the below example
// ```go
// createAppConfigToml(t)
// defer os.RemoveAll(config.DefaultChefWorkstationDirectory) // clean up
// ```
func createAppConfigToml(t *testing.T) {
	if err := os.MkdirAll(config.DefaultChefWorkstationDirectory, os.ModePerm); err != nil {
		t.Fatal(err)
	}

	var (
		featuresList = []byte(`
[features]
all = false
analyze = true
xyz = false
`)
		userConfigFile = filepath.Join(
			config.DefaultChefWorkstationDirectory,
			config.DefaultChefWSUserConfigFile,
		)
		appConfigFile = filepath.Join(
			config.DefaultChefWorkstationDirectory,
			config.DefaultChefWSAppConfigFile,
		)
	)
	err := ioutil.WriteFile(appConfigFile, featuresList, 0666)
	if err != nil {
		t.Fatal(err)
	}

	// create an empty user config to avoid failures on loading the config
	err = ioutil.WriteFile(userConfigFile, []byte(``), 0666)
	if err != nil {
		t.Fatal(err)
	}
}
