package reporting_test

import (
	"fmt"
	"testing"

	subject "github.com/chef/chef-analyze/pkg/reporting"
	"github.com/stretchr/testify/assert"
)

func TestRunCookstyle(t *testing.T) {
	desiredOutput := `{"metadata" : { "rubocop_version" : "0.0", "ruby_engine": "ruby", "ruby_version": "0.0", "ruby_patchlevel": "0", "ruby_platform": ""},
                     "files" : [
											 { "path" : "/path/to/file1.rb", "offenses" : [] },
											 { "path" : "/path/to/file2.rb", "offenses" : [] }
                     ]}`

	runner := MockCookstyleRunner{desiredOutput: []byte(desiredOutput), desiredError: nil}
	result, err := subject.RunCookstyle("workingDir", runner)

	// There is not much value in testing the functionality of json.Unmarshal
	// by verifying the struct results. that the structure deserialized
	// correctly -- json.Unmarshal is reliably tested as part of go.
	// Any successful unmarshal will have the correct type, and any failed one
	// would give us a not-nil error,
	assert.Nil(t, err)
	assert.NotNil(t, result)
}

func TestRunCookstyle_withViolations(t *testing.T) {
	// This is not ideal - we should also mock the error behavior of exit code 1
	// for a complete test, but so far I haven't been able to find a reasonable way to do that.

	desiredOutput := `{"metadata" : { "rubocop_version" : "0.0", "ruby_engine": "ruby", "ruby_version": "0.0", "ruby_patchlevel": "0", "ruby_platform": ""},
                     "files" : [
											 { "path" : "/path/to/file1.rb",
											   "offenses" : [
													 { "severity" : "high", "message" : "err1", "cop_name" : "cop1", "corrected": false, "correctable": false, "location" : { "start_line" : 1, "start_column": 1, "last_line": 1, "last_column": 1, "length": 1, "line": 1, "column" : 1 }},
													 { "severity" : "med", "message" : " err2", "cop_name" : "cop2", "corrected": false, "correctable": true, "location" : { "start_line" : 1, "start_column": 1, "last_line": 1, "last_column": 1, "length": 1, "line": 1, "column" : 1 }}
                         ]
											 },
											 { "path" : "/path/to/file2.rb",
											   "offenses" : [
													 { "severity" : "high", "message" : "err3", "cop_name" : "cop3", "corrected": false, "correctable": false, "location" : { "start_line" : 1, "start_column": 1, "last_line": 1, "last_column": 1, "length": 1, "line": 1, "column" : 1 }},
													 { "severity" : "med", "message" : " err4", "cop_name" : "cop4", "corrected": false, "correctable": true, "location" : { "start_line" : 1, "start_column": 1, "last_line": 1, "last_column": 1, "length": 1, "line": 1, "column" : 1 }}
                         ]
											 }
                     ]}`

	runner := MockCookstyleRunner{desiredOutput: []byte(desiredOutput), desiredError: nil}
	result, err := subject.RunCookstyle("workingDir", runner)
	assert.Nil(t, err)
	assert.NotNil(t, result)
}

func TestRunCookstyle_withOtherErrors(t *testing.T) {
	desiredErr := fmt.Errorf("Test error")

	runner := MockCookstyleRunner{desiredOutput: []byte(`{}`), desiredError: desiredErr}
	result, err := subject.RunCookstyle("workingDir", runner)
	assert.Nil(t, result)
	assert.Error(t, err)

}

type MockCookstyleRunner struct {
	desiredOutput []byte
	desiredError  error
}

func (mcr MockCookstyleRunner) Run(workingDir string) ([]byte, error) {
	return mcr.desiredOutput, mcr.desiredError
}
