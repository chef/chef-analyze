package reporting

import (
	"encoding/json"
	"os/exec"
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
	}
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
	}
	Files []CookbookFile `json:"files"`
}

// Should these live here?
type CookstyleRunner interface {
	Run(workingDir string) ([]byte, error)
}

type ExecCookstyleRunner struct{}

func (er ExecCookstyleRunner) Run(workingDir string) ([]byte, error) {
	cmd := exec.Command("cookstyle", "--format", "json")
	cmd.Dir = workingDir
	return cmd.Output()
}

func RunCookstyle(workingDir string, runner CookstyleRunner) (*CookstyleResult, error) {
	output, err := runner.Run(workingDir)
	if exitError, ok := err.(*exec.ExitError); ok {
		// https://docs.rubocop.org/en/latest/basic_usage/#exit-codes
		// exit code of 1 is ok , it means some violations were found
		if exitError.ExitCode() != 1 {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	var csr CookstyleResult
	err = json.Unmarshal(output, &csr)
	if err != nil {
		return nil, err
	}

	return &csr, nil
}
