package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsReportCommandArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{name: "no args", args: []string{"chef-analyze"}, want: false},
		{name: "non-report subcommand", args: []string{"chef-analyze", "capture"}, want: false},
		{name: "report subcommand", args: []string{"chef-analyze", "report"}, want: true},
		{name: "report with extra args", args: []string{"chef-analyze", "report", "nodes"}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, isReportCommandArgs(tt.args))
		})
	}
}

func TestIsTopLevelHelpArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{name: "no args", args: []string{"chef-analyze"}, want: false},
		{name: "help word", args: []string{"chef-analyze", "help"}, want: true},
		{name: "short help", args: []string{"chef-analyze", "-h"}, want: true},
		{name: "long help", args: []string{"chef-analyze", "--help"}, want: true},
		{name: "subcommand help", args: []string{"chef-analyze", "report", "--help"}, want: false},
		{name: "other command", args: []string{"chef-analyze", "config"}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, isTopLevelHelpArgs(tt.args))
		})
	}
}
