package cmd

// Argument classification for root routing lives here as pure helpers.
// Extension contract: add new top-level classification rules in this file,
// then update cmd/argv_test.go and docs/SUBSYSTEM_CMD_ROUTING.md in the same PR.

// isReportCommandArgs returns true when args target the top-level report command.
func isReportCommandArgs(args []string) bool {
	if len(args) <= 1 {
		return false
	}
	return args[1] == "report"
}

// isTopLevelHelpArgs returns true for top-level help invocations only.
func isTopLevelHelpArgs(args []string) bool {
	if len(args) != 2 {
		return false
	}
	return args[1] == "help" || args[1] == "-h" || args[1] == "--help"
}
