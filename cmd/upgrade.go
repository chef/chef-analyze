package cmd

import "github.com/spf13/cobra"

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Perform automatic upgrades",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, args []string) error {
		// TODO @afiune maybe here is where we perform upgrades
		return nil
	},
}
