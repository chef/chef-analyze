package cmd

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/chef/chef-analyze/pkg/config"
)

var (
	validReportTypes = []string{"nodes", "cookbooks"}
	reportCmd        = &cobra.Command{
		Use:   "report [type]",
		Short: "Generate reports about your Chef inventory",
		Long: `Generate reports to analyze your Chef inventory, the available
report types are:
  nodes     - Display a list of nodes with Chef Client Infra version and a list of cookbooks being used
  cookbooks - Display a list of cookbooks with their violations and a list of nodes using it
`,
		Args: reportArgsValidator,
		RunE: func(_ *cobra.Command, args []string) error {

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
)

func reportArgsValidator(_ *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("report type not specified")
	}
	if isValidReportType(args[0]) {
		return nil
	}
	return fmt.Errorf("\n  invalid report type specified: %s\n  %s", args[0], displayReportTypes())
}

func displayReportTypes() string {
	return fmt.Sprintf("available report types: %s", strings.Join(validReportTypes, ", "))
}

func isValidReportType(t string) bool {
	for _, vt := range validReportTypes {
		if vt == t {
			return true
		}
	}
	return false
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
