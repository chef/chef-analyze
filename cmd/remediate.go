package cmd

import "github.com/spf13/cobra"

var remediateCmd = &cobra.Command{
	Use:   "remediate",
	Short: "Perform automatic remediations",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, args []string) error {
		// TODO @afiune maybe here is where we perform remediations
		return nil
	},
}
