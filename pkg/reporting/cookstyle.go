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

package reporting

import (
	"encoding/json"
	"os/exec"

	"github.com/pkg/errors"
)

type CookstyleOffense struct {
	Severity    string `json:"severity"`
	Message     string `json:"message"`
	CopName     string `json:"cop_name"`
	Corrected   bool   `json:"corrected"`
	Correctable bool   `json:"correctable"`

	Location struct {
		StartLine   int `json:"start_line"`
		StartColumn int `json:"start_column"`
		LastLine    int `json:"last_line"`
		LastColumn  int `json:"last_column"`
		Length      int `json:"length"`
		Line        int `json:"line"`
		Column      int `json:"column"`
	} `json:"location"`
}

type CookbookFile struct {
	Path     string             `json:"path"`
	Offenses []CookstyleOffense `json:"offenses"`
}

type CookstyleResult struct {
	Metadata struct {
		RubocopVersion string `json:"rubocop_version"`
		RubyEngine     string `json:"ruby_engine"`
		RubyVersion    string `json:"ruby_version"`
		RubyPatchlevel string `json:"ruby_patchlevel"`
		RubyPlatform   string `json:"ruby_platform"`
	} `json:"metadata"`
	Files []CookbookFile `json:"files"`
}

type CookstyleRunner struct {
	Opts []string
}

func (ecr *CookstyleRunner) Run(workingDir string) (*CookstyleResult, error) {
	var (
		cookstyleRes CookstyleResult
		cmd          = exec.Command("cookstyle", ecr.Opts...)
	)
	cmd.Dir = workingDir

	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// https://docs.rubocop.org/en/latest/basic_usage/#exit-codes
			// exit code of 1 is ok, it means some violations were found
			if exitError.ExitCode() != 1 {
				// we wrap the Stderr into the error message so that callers
				// don't have to cast this error as an ExitError type again
				// to access the error message that cookstle sends to Stderr
				return nil, errors.Wrap(exitError, string(exitError.Stderr))
			}
		} else {
			return nil, err
		}
	}

	err = json.Unmarshal(output, &cookstyleRes)
	if err != nil {
		return nil, err
	}

	return &cookstyleRes, nil
}

func NewCookstyleRunner() *CookstyleRunner {
	return &CookstyleRunner{
		Opts: []string{"--format", "json"},
	}
}

func RunCookstyle(workingDir string) (*CookstyleResult, error) {
	runner := NewCookstyleRunner()
	return runner.Run(workingDir)
}
