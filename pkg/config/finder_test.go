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

func TestFindConfigFileNotFoundErr(t *testing.T) {
	cfgFile, err := subject.FindConfigFile("config.toml")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "file 'config.toml' not found")
		assert.Empty(t, cfgFile, "no config file should be found")
	}
}

func TestFindConfigFileNoHomeDirectoryErr(t *testing.T) {
	// save real HOME, empty it and restored it after this test runs
	savedHome := os.Getenv("HOME")
	os.Setenv("HOME", "")
	defer os.Setenv("HOME", savedHome)

	cfgFile, err := subject.FindConfigFile("config.toml")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "unable to detect home directory")
		assert.Contains(t, err.Error(), "$HOME is not defined")
		assert.Empty(t, cfgFile, "no config file should be found")
	}
}

func TestFindConfigFileInsideHomeDirectory(t *testing.T) {
	mockedHomeDir, err := ioutil.TempDir("", "mocked-home-dir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(mockedHomeDir) // clean up

	chefWorkstationDir := filepath.Join(mockedHomeDir, ".chef-workstation")
	if err := os.MkdirAll(chefWorkstationDir, os.ModePerm); err != nil {
		t.Fatal(err)
	}

	configToml := filepath.Join(chefWorkstationDir, "config.toml")
	if err := ioutil.WriteFile(configToml, []byte("data"), 0666); err != nil {
		t.Fatal(err)
	}

	// save real HOME, set a mocked home directory and restored it after this test runs
	savedHome := os.Getenv("HOME")
	os.Setenv("HOME", mockedHomeDir)
	defer os.Setenv("HOME", savedHome)

	cfgFile, err := subject.FindConfigFile("config.toml")
	if assert.Nil(t, err) {
		assert.Equal(t, configToml, cfgFile, "mocked config file should be found")
	}
}

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

func TestFindChefWorkstationConfigFileExists(t *testing.T) {
	createConfigToml(t)
	defer os.RemoveAll(".chef-workstation") // clean up

	configToml, err := subject.FindChefWorkstationConfigFile()
	if assert.Nil(t, err) {
		assert.NotEmpty(t, configToml, "the config.toml should be found")
	}
}
