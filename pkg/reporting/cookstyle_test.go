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

package reporting_test

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	subject "github.com/chef/chef-analyze/pkg/reporting"
)

func TestRunCookstyleError(t *testing.T) {
	result, err := subject.RunCookstyle("/foo")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "\"cookstyle\": executable file not found")
	}
	assert.Nil(t, result)
}

func TestRunCookstyleWithBinstubs(t *testing.T) {
	savedPath := setupBinstubsDir()
	defer os.Setenv("PATH", savedPath)

	dir, err := ioutil.TempDir("", "cookbook")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	result, err := subject.RunCookstyle(dir)
	assert.Nil(t, err)
	if assert.NotNil(t, result) {
		assert.Equal(t, len(result.Files), 0)
	}
}

func TestCookstyleRunnerWithBinstubs(t *testing.T) {
	savedPath := setupBinstubsDir()
	defer os.Setenv("PATH", savedPath)

	dir, err := ioutil.TempDir("", "cookbook")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	runner := subject.NewCookstyleRunner()
	assert.Equal(t,
		[]string{
			"--format", "json",
			"--only", "Chef/Deprecations,Chef/Correctness",
			"--force-default-config"},
		runner.Opts,
		"default cookstyle options need to be updated",
	)

	// manipulating the output of the binstub `cookstyle` command
	// that lives inside the `binstubs/` directory
	//
	// Emulating a syntax error
	runner.Opts = []string{"syntax-error"}
	result, err := runner.Run(dir)
	assert.Nil(t, result)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(),
			"invalid character 'e' looking for beginning of object key string")
	}

	// Emulating a known exit code: 1
	//
	// https://docs.rubocop.org/en/latest/basic_usage/#exit-codes
	// exit code of 1 is ok, it means some violations were found

	runner.Opts = []string{"known-exit-code-error"}
	result, err = runner.Run(dir)
	assert.Nil(t, err)
	if assert.NotNil(t, result) {
		assert.Equal(t, len(result.Files), 0)
	}

	// Emulating a known exit code: 1
	runner.Opts = []string{"exit-code-error", "2"}
	result, err = runner.Run(dir)
	assert.Nil(t, result)
	if assert.NotNil(t, err) {
		assert.Contains(t,
			err.Error(), "exit status 2",
			"wrong exit code, check either the binstub or the Go code for validation",
		)
		assert.Contains(t,
			err.Error(), "ERROR: something happened with cookstyle",
			"wrong Stderr, check either the binstub or the Go code for validation",
		)
	}
}
