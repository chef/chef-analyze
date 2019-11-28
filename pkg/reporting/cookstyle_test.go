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

package reporting_test

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
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
		[]string{"--format", "json"}, runner.Opts,
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

// returns the binstubs/ directory path
func binstubsDir() string {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return filepath.Join(pwd, "..", "..", "binstubs")
}

// setups up the binstubs directory at the front of the PATH environment
// environment and returns the previous value so that callers can set
// it back to its original value
//
// example:
//
// ```go
// savedPath := setupBinstubsDir()
// defer os.Setenv("PATH", savedPath)
// ```
func setupBinstubsDir() string {
	pathEnv := os.Getenv("PATH")
	newPathEnv := binstubsDir() + ":" + pathEnv
	os.Setenv("PATH", newPathEnv)
	return pathEnv
}

// TODO @afiune I think this should be integration tests instead, I will come back to this
//func TestRunCookstyle(t *testing.T) {
//desiredOutput := `{"metadata" : { "rubocop_version" : "0.0", "ruby_engine": "ruby", "ruby_version": "0.0", "ruby_patchlevel": "0", "ruby_platform": ""},
//"files" : [
//{ "path" : "/path/to/file1.rb", "offenses" : [] },
//{ "path" : "/path/to/file2.rb", "offenses" : [] }
//]}`

//runner := MockCookstyleRunner{desiredOutput: []byte(desiredOutput), desiredError: nil}
//result, err := subject.RunCookstyle("/tmp")

// There is not much value in testing the functionality of json.Unmarshal
// by verifying the struct results. that the structure deserialized
// correctly -- json.Unmarshal is reliably tested as part of go.
// Any successful unmarshal will have the correct type, and any failed one
// would give us a not-nil error,
//assert.NotNil(t, err)
//assert.Nil(t, result)
//}

//func TestRunCookstyle_withViolations(t *testing.T) {
// This is not ideal - we should also mock the error behavior of exit code 1
// for a complete test, but so far I haven't been able to find a reasonable way to do that.

//desiredOutput := `{"metadata" : { "rubocop_version" : "0.0", "ruby_engine": "ruby", "ruby_version": "0.0", "ruby_patchlevel": "0", "ruby_platform": ""},
//"files" : [
//{ "path" : "/path/to/file1.rb",
//"offenses" : [
//{ "severity" : "high", "message" : "err1", "cop_name" : "cop1", "corrected": false, "correctable": false, "location" : { "start_line" : 1, "start_column": 1, "last_line": 1, "last_column": 1, "length": 1, "line": 1, "column" : 1 }},
//{ "severity" : "med", "message" : " err2", "cop_name" : "cop2", "corrected": false, "correctable": true, "location" : { "start_line" : 1, "start_column": 1, "last_line": 1, "last_column": 1, "length": 1, "line": 1, "column" : 1 }}
//]
//},
//{ "path" : "/path/to/file2.rb",
//"offenses" : [
//{ "severity" : "high", "message" : "err3", "cop_name" : "cop3", "corrected": false, "correctable": false, "location" : { "start_line" : 1, "start_column": 1, "last_line": 1, "last_column": 1, "length": 1, "line": 1, "column" : 1 }},
//{ "severity" : "med", "message" : " err4", "cop_name" : "cop4", "corrected": false, "correctable": true, "location" : { "start_line" : 1, "start_column": 1, "last_line": 1, "last_column": 1, "length": 1, "line": 1, "column" : 1 }}
//]
//}
//]}`

//runner := MockCookstyleRunner{desiredOutput: []byte(desiredOutput), desiredError: nil}
//result, err := subject.RunCookstyle("/tmp")
//assert.Nil(t, err)
//assert.NotNil(t, result)
//}

//func TestRunCookstyle_withOtherErrors(t *testing.T) {
//desiredErr := fmt.Errorf("Test error")

//runner := MockCookstyleRunner{desiredOutput: []byte(`{}`), desiredError: desiredErr}
//result, err := subject.RunCookstyle("/tmp")
//assert.Nil(t, result)
//assert.Error(t, err)

//}

//type MockCookstyleRunner struct {
//desiredOutput []byte
//desiredError  error
//}

//func (mcr MockCookstyleRunner) Run(workingDir string) ([]byte, error) {
//return mcr.desiredOutput, mcr.desiredError
//}
