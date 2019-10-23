package reporting

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/chef/chef-analyze/pkg/config"
)

func Cookbooks(cfg *config.Config) error {
	chefClient, err := NewChefClient(cfg)
	if err != nil {
		return err
	}

	cbooksList, err := chefClient.Cookbooks.ListAvailableVersions("all")
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
