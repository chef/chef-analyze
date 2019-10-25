package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelpCommand(t *testing.T) {
	out, err, exitcode := ChefAnalyze("help")
	verifyGlobalFlagsFromOutput(t, out.String())
	assert.Contains(t,
		out.String(),
		"Use \"chef-analyze [command] --help\" for more information about a command.",
		"STDOUT bottom message doesn't match")
	assert.Empty(t,
		err.String(),
		"STDERR should be empty")
	assert.Equal(t, 0, exitcode,
		"EXITCODE is not the expected one")
}

func TestHelpFlags_h(t *testing.T) {
	out, err, exitcode := ChefAnalyze("-h")
	verifyGlobalFlagsFromOutput(t, out.String())
	assert.Contains(t,
		out.String(),
		"Use \"chef-analyze [command] --help\" for more information about a command.",
		"STDOUT bottom message doesn't match")
	assert.Empty(t,
		err.String(),
		"STDERR should be empty")
	assert.Equal(t, 0, exitcode,
		"EXITCODE is not the expected one")
}

func TestHelpFlags__help(t *testing.T) {
	out, err, exitcode := ChefAnalyze("--help")
	verifyGlobalFlagsFromOutput(t, out.String())
	assert.Contains(t,
		out.String(),
		"Use \"chef-analyze [command] --help\" for more information about a command.",
		"STDOUT bottom message doesn't match")
	assert.Empty(t,
		err.String(),
		"STDERR should be empty")
	assert.Equal(t, 0, exitcode,
		"EXITCODE is not the expected one")
}

func TestHelpCommandDisplayHelpFromCommand(t *testing.T) {
	out, err, exitcode := ChefAnalyze("help", "report")
	verifyGlobalFlagsFromOutput(t, out.String())
	assert.Contains(t,
		out.String(),
		"chef-analyze report [command]",
		"STDOUT missing help about the report sub-command")
	assert.Empty(t,
		err.String(),
		"STDERR should be empty")
	assert.Equal(t, 0, exitcode,
		"EXITCODE is not the expected one")
}

func TestHelpCommandDisplayHelpFromUnknownCommand(t *testing.T) {
	out, err, exitcode := ChefAnalyze("help", "foo")
	// NOTE since this is an unknown command, we should display the help
	// message via STDERR and not STDOUT
	verifyGlobalFlagsFromOutput(t, err.String())
	assert.Contains(t,
		err.String(),
		"Use \"chef-analyze [command] --help\" for more information about a command.",
		"STDERR bottom message doesn't match")
	assert.Empty(t,
		out.String(),
		"STDOUT should be empty")
	assert.Equal(t, 0, exitcode,
		"EXITCODE is not the expected one")
}

// verify global flags from either STDOUT or STDERR
// this is just to NOT duplicate code in multiple tests
func verifyGlobalFlagsFromOutput(t *testing.T, stdout string) {
	assert.Contains(t, stdout,
		"--chef_server_url",
		"STDOUT chef_server_url flag doesn't exist")
	assert.Contains(t, stdout,
		"--client_key",
		"STDOUT client_key flag doesn't exist")
	assert.Contains(t, stdout,
		"--client_name",
		"STDOUT client_name flag doesn't exist")
	assert.Contains(t, stdout,
		"--credentials",
		"STDOUT credentials flag doesn't exist")
	assert.Contains(t, stdout,
		"--help",
		"STDOUT help flag doesn't exist")
	assert.Contains(t, stdout,
		"--profile",
		"STDOUT profile flag doesn't exist")
}
