package cmd

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/chef/chef-analyze/pkg/config"
)

var reportCmd = &cobra.Command{
	Use:   "report [type]",
	Short: "Generate reports about your Chef inventory",
	Long: `Generate reports to analyze your Chef inventory, the available
report types are:
  nodes     - Display nodes report
  cookbooks - Display cookbooks report
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		cfg, err := config.FromViper()
		if err != nil {
			return err
		}

		// Allowed reports
		switch args[0] {
		case "nodes":
			if err := nodesReport(cfg); err != nil {
				return err
			}
		case "cookbooks":
			if err := cookbooksReport(cfg); err != nil {
				return err
			}

		default:
			return errors.New("Invalid report type, available reports are: [nodes|cookbooks]")
		}

		return nil
	},
}

func nodesReport(cfg *config.Config) error {
	var (
		query = map[string]interface{}{
			"name":         []string{"name"},
			"chef_version": []string{"chef_packages", "chef", "version"},
			"os":           []string{"platform"},
			"os_version":   []string{"platform_version"}}
		//using 'v' not 's' because not all fields will have values.
		formatString = "%30s   %-12v   %-15v   %-10v\n"
		pres, err    = cfg.ChefClient.Search.PartialExec("node", "*:*", query)
	)
	if err != nil {
		return errors.Wrap(err, "unable to get node(s) information")
	}

	fmt.Printf(formatString, "Node Name", "Chef Version", "OS", "OS Version")
	for _, element := range pres.Rows {
		v := element.(map[string]interface{})["data"].(map[string]interface{})
		fmt.Printf(formatString, v["name"], v["chef_version"], v["os"], v["os_version"])
	}

	return nil
}

func cookbooksReport(cfg *config.Config) error {
	cbooksList, err := cfg.ChefClient.Cookbooks.ListAvailableVersions("all")
	if err != nil {
		return errors.Wrap(err, "unable to get cookbook(s) information")
	}

	formatString := "%30s   %-12v\n"
	fmt.Printf(formatString, "Cookbook Name", "Version(s)")
	for cookbook, cbookVersions := range cbooksList {
		versionsArray := make([]string, len(cbookVersions.Versions))
		for i, details := range cbookVersions.Versions {
			versionsArray[i] = details.Version
		}
		fmt.Printf(formatString, cookbook, strings.Join(versionsArray[:], ", "))
	}

	return nil
}
