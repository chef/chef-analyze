package reporting

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/chef/chef-analyze/pkg/config"
)

func Nodes(cfg *config.Config) error {
	var (
		query = map[string]interface{}{
			"name":         []string{"name"},
			"chef_version": []string{"chef_packages", "chef", "version"},
			"os":           []string{"platform"},
			"os_version":   []string{"platform_version"}}
		//using 'v' not 's' because not all fields will have values.
		formatString = "%30s   %-12v   %-15v   %-10v\n"
	)
	chefClient, err := NewChefClient(cfg)
	if err != nil {
		return err
	}

	pres, err := chefClient.Search.PartialExec("node", "*:*", query)
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
