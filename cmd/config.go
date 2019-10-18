package cmd

import "github.com/spf13/cobra"

var (
	configCmd = &cobra.Command{
		Use:   "config",
		Short: "Manage your local Chef configuration (default: $HOME/.chef/credentials)",
	}
	configVerifyCmd = &cobra.Command{
		Use:   "verify",
		Short: "Verify your Chef configuration",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, args []string) error {
			// TODO @afiune verify the config
			return nil
		},
	}
	configInitCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize a local Chef configuration",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, args []string) error {
			// TODO @afiune help user initialize a local config file
			return nil
		},
	}
)

func init() {
	// adds the verify command as a sub-command of the config command
	// => chef-analyze config verify
	configCmd.AddCommand(configVerifyCmd)
	// adds the verify command as a sub-command of the config command
	// => chef-analyze config init
	configCmd.AddCommand(configInitCmd)
}
