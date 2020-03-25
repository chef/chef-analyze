//
// Copyright 2020 Chef Software, Inc.
// Author: Marc Paradise <marc@chef.io>
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

package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/chef/chef-analyze/pkg/reporting"
	"github.com/chef/go-libs/credentials"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	// behavior:
	//  - repo dir: if it does not exist, we will create it by invoking chef generate repo node-NAME-repo.
	//  - capture json object definitions for node, env, roles. Save them to the chef-repo.
	//  - capture cookbooks and versions from automatic attributes. Save each cookbook @ version to the chef-repo.
	// Possible options (future)
	// --repo-dir -> root dir of repo to use.  Will update existing if one exists.
	// Thoughts: this may make it harder to coordinate, now user will have to track telling kitchen about
	//           repo location? Or is that something we'd capture in the kitchen.yml we generate eventually?
	//           to t
	//     Default: ./node-NAME-repo

	captureCmd = &cobra.Command{
		Use:   "capture NODE-NAME",
		Short: "Capture a node's state into a local chef-repo",
		Args:  cobra.ExactArgs(1),
		Long: `Captures a node's state as a local chef-repo, which
can then be used to converge locally.`,
		RunE: func(_ *cobra.Command, args []string) error {
			// TODO - support multiple node names?
			creds, err := credentials.FromViper(
				// TODO this isn't just reportsflags any longer.
				reportsFlags.profile,
				overrideCredentials(),
			)
			if err != nil {
				return err
			}

			nodeName := args[0]
			err = createOutputDirectories()
			if err != nil {
				return err
			}

			cfg := &reporting.Reporting{Credentials: creds}
			if reportsFlags.noSSLverify {
				cfg.NoSSLVerify = true
			}

			chefClient, err := reporting.NewChefClient(cfg)
			if err != nil {
				return err
			}

			repoName := fmt.Sprintf("node-%s-repo", nodeName)
			// TODO - future iteration - give option to set the destination path.
			dirName := fmt.Sprintf("./%s", repoName)

			// abort if it exists, have them remove it first.
			_, err = os.Stat(dirName)
			if err == nil {
				fmt.Printf(RepositoryAlreadyExistsE002, dirName, nodeName)

				return nil

			} else {
				if !os.IsNotExist(err) {
					return err
				}
			}

			fmt.Println(" - Setting up local repository")
			cmd := exec.Command("chef", "generate", "repo", repoName)
			_, err = cmd.Output()
			if err != nil {
				return err
			}

			// Some files are created that we don't need in our repo, let's remove them
			err = os.RemoveAll(fmt.Sprintf("%s/cookbooks/example", dirName))
			if err != nil {
				return errors.Wrap(err, "Could not remove pre-created repo content: example cookbook")
			}

			err = os.RemoveAll(fmt.Sprintf("%s/cookbooks/example", dirName))
			if err != nil {
				return errors.Wrap(err, "Could not remove pre-created repo content: example cookbook")
			}

			nc := reporting.NewNodeCapture(nodeName, dirName,
				chefClient.Nodes, chefClient.Roles,
				chefClient.Environments, chefClient.Cookbooks,
				&reporting.ObjectWriter{RootDir: dirName},
			)
			go nc.Run()
			for progress := range nc.Progress {
				switch progress {
				case reporting.FetchingNode:
					fmt.Printf(" - Capturing node '%s'\n", nodeName)
				case reporting.FetchingCookbooks:
					fmt.Println(" - Capturing cookbooks...")
				case reporting.FetchingRoles:
					fmt.Println(" - Capturing roles...")
				case reporting.FetchingEnvironment:
					fmt.Println(" - Capturing environment...")
				case reporting.FetchingComplete:
				}
			}
			if nc.Error != nil {
				return nc.Error
			}

			fmt.Printf("Repository has been created in '%s'\n", dirName)
			return nil
		},
	}
)
