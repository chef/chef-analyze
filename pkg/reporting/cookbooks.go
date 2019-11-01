//
// Copyright 2019 Chef Software, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

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

func DemoFunction() error {
	fmt.Println("New function that has test! :(")
	return nil
}
