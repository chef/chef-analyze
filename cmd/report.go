package cmd

import (
	"github.com/spf13/cobra"

	"github.com/chef/chef-analyze/pkg/config"
	"github.com/chef/chef-analyze/pkg/reporting"
)

var (
	reportCmd = &cobra.Command{
		Use:   "report",
		Short: "Generate reports about your Chef inventory",
	}
	reportCookbooksCmd = &cobra.Command{
		Use:   "cookbooks",
		Short: "Generates a cookbook oriented report",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.FromViper(
				cmd.Flag("profile").Value.String(),
				overrideCredentials(),
			)
			if err != nil {
				return err
			}

			return reporting.Cookbooks(cfg)
		},
	}
	reportNodesCmd = &cobra.Command{
		Use:   "nodes",
		Short: "Generates a nodes oriented report",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.FromViper(
				cmd.Flag("profile").Value.String(),
				overrideCredentials(),
			)
			if err != nil {
				return err
			}

			return reporting.Nodes(cfg)
		},
	}
)

func init() {
	// adds the cookbooks command as a sub-command of the report command
	// => chef-analyze report cookbooks
	reportCmd.AddCommand(reportCookbooksCmd)
	// adds the nodes command as a sub-command of the report command
	// => chef-analyze report nodes
	reportCmd.AddCommand(reportNodesCmd)
}
