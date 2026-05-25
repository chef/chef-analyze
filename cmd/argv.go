package cmd

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
