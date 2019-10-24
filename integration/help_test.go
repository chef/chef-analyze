package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelpCommand(t *testing.T) {
	out, err, exitcode := ChefAnalyze("help")
	assert.Contains(t,
		out.String(),
		"Use \"chef-analyze [command] --help\" for more information about a command.",
		"STDOUT doesn't match")
	assert.Empty(t,
		err.String(),
		"STDERR should be empty")
	assert.Equal(t, 0, exitcode,
		"EXITCODE is not the expected one")
}
