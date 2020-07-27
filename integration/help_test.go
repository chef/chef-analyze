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

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelpCommand(t *testing.T) {
	out, err, exitcode := ChefAnalyze("help")
	assert.Equal(t, "Please use the 'chef' executable to access Upgrade Lab features.\n", out.String())
	assert.Empty(t, err.String(), "STDERR should be empty")
	assert.Equal(t, 0, exitcode, "EXITCODE is not the expected one")
}

func TestHelpFlags_h(t *testing.T) {
	out, err, exitcode := ChefAnalyze("-h")
	assert.Equal(t, "Please use the 'chef' executable to access Upgrade Lab features.\n", out.String())
	assert.Empty(t, err.String(), "STDERR should be empty")
	assert.Equal(t, 0, exitcode, "EXITCODE is not the expected one")
}

func TestHelpFlags__help(t *testing.T) {
	out, err, exitcode := ChefAnalyze("--help")
	assert.Equal(t, "Please use the 'chef' executable to access Upgrade Lab features.\n", out.String())
	assert.Empty(t, err.String(), "STDERR should be empty")
	assert.Equal(t, 0, exitcode, "EXITCODE is not the expected one")
}

func TestHelpNoArgs(t *testing.T) {
	out, err, exitcode := ChefAnalyze()
	assert.Equal(t, "Please use the 'chef' executable to access Upgrade Lab features.\n", out.String())
	assert.Empty(t, err.String(), "STDERR should be empty")
	assert.Equal(t, 0, exitcode, "EXITCODE is not the expected one")
}
func TestHelpCommandDisplayHelpForCapture(t *testing.T) {
	out, err, exitcode := ChefAnalyze("help", "capture")

	var expected = `Captures a node's state as a local chef-repo, which
can then be used to converge locally.

Usage:
  chef capture NODE-NAME [flags]

Flags:
  -s, --chef-server-url string   Chef Infra Server URL
  -k, --client-key string        Chef Infra Server API client key
  -n, --client-name string       Chef Infra Server API client name
  -c, --credentials string       credentials file (default $HOME/.chef/credentials)
  -h, --help                     help for capture
  -p, --profile string           profile to use from credentials file (default "default")
  -o, --ssl-no-verify            Do not verify SSL when connecting to Chef Infra Server (default: verify)
  -d, --with-data-bags           download all data bags as part of node capture
`
	assert.Equal(t, expected, out.String())
	assert.Empty(t, err.String(), "STDERR should be empty")
	assert.Equal(t, 0, exitcode, "EXITCODE is not the expected one")
}

func TestHelpCommandDisplayHelpForReport(t *testing.T) {
	out, err, exitcode := ChefAnalyze("help", "report")
	expected := `Generate reports from a Chef Infra Server

Usage:
  chef report [command]

Available Commands:
  cookbooks   Generates a cookbook-oriented report
  nodes       Generates a nodes-oriented report

Flags:
  -a, --anonymize                replace cookbook and node names with hash values
  -s, --chef-server-url string   Chef Infra Server URL
  -k, --client-key string        Chef Infra Server API client key
  -n, --client-name string       Chef Infra Server API client name
  -c, --credentials string       credentials file (default $HOME/.chef/credentials)
  -f, --format string            output format: txt is human readable, csv is machine readable (default "txt")
  -h, --help                     help for report
  -F, --node-filter string       Search filter to apply to nodes
  -p, --profile string           profile to use from credentials file (default "default")
  -o, --ssl-no-verify            Do not verify SSL when connecting to Chef Infra Server (default: verify)

Use "chef report [command] --help" for more information about a command.
`
	assert.Equal(t, expected, out.String())
	assert.Empty(t, err.String(), "STDERR should be empty")
	assert.Equal(t, 0, exitcode, "EXITCODE is not the expected one")
}

func TestHelpCommandDisplayHelpForReportCookbooks(t *testing.T) {
	out, err, exitcode := ChefAnalyze("help", "report", "cookbooks")
	expected := `Generates a cookbook-oriented report containing details about the
upgrade compatibility errors and node cookbook usage.

The result is written to file.

Usage:
  chef report cookbooks [flags]

Flags:
  -h, --help             help for cookbooks
  -u, --only-unused      generate a report with only cookbooks that are not included in any node's runlist
  -V, --verify-upgrade   verify the upgrade compatibility of every cookbook
  -w, --workers int      maximum number of parallel workers at once (default 50)

Global Flags:
  -a, --anonymize                replace cookbook and node names with hash values
  -s, --chef-server-url string   Chef Infra Server URL
  -k, --client-key string        Chef Infra Server API client key
  -n, --client-name string       Chef Infra Server API client name
  -c, --credentials string       credentials file (default $HOME/.chef/credentials)
  -f, --format string            output format: txt is human readable, csv is machine readable (default "txt")
  -F, --node-filter string       Search filter to apply to nodes
  -p, --profile string           profile to use from credentials file (default "default")
  -o, --ssl-no-verify            Do not verify SSL when connecting to Chef Infra Server (default: verify)
`
	assert.Equal(t, expected, out.String())
	assert.Empty(t, err.String(), "STDERR should be empty")
	assert.Equal(t, 0, exitcode, "EXITCODE is not the expected one")
}

func TestHelpCommandDisplayHelpForReportNodes(t *testing.T) {
	out, err, exitcode := ChefAnalyze("help", "report", "nodes")
	var expected = `Generates a nodes-oriented report containing basic information about the node,
any applied policies, and the cookbooks used during the most recent chef-client run

Usage:
  chef report nodes [flags]

Flags:
  -h, --help   help for nodes

Global Flags:
  -a, --anonymize                replace cookbook and node names with hash values
  -s, --chef-server-url string   Chef Infra Server URL
  -k, --client-key string        Chef Infra Server API client key
  -n, --client-name string       Chef Infra Server API client name
  -c, --credentials string       credentials file (default $HOME/.chef/credentials)
  -f, --format string            output format: txt is human readable, csv is machine readable (default "txt")
  -F, --node-filter string       Search filter to apply to nodes
  -p, --profile string           profile to use from credentials file (default "default")
  -o, --ssl-no-verify            Do not verify SSL when connecting to Chef Infra Server (default: verify)
`

	assert.Equal(t, expected, out.String())
	assert.Empty(t, err.String(), "STDERR should be empty")
	assert.Equal(t, 0, exitcode, "EXITCODE is not the expected one")

}

func TestHelpCommandDisplayHelpFromUnknownCommand(t *testing.T) {
	out, err, exitcode := ChefAnalyze("help", "foo")
	// NOTE since this is an unknown command, we should display the help
	// message via STDERR and not STDOUT
	assert.Contains(t, err.String(),
		"Use \"chef [command] --help\" for more information about a command.",
		"STDERR bottom message doesn't match")
	assert.Empty(t, out.String(), "STDOUT should be empty")
	assert.Equal(t, 0, exitcode, "EXITCODE is not the expected one")
}
