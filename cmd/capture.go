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
	"path/filepath"
	"strings"

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
			creds, err := credentials.FromViper(
				infraFlags.profile,
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
			if infraFlags.noSSLverify {
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

			err = os.RemoveAll(fmt.Sprintf("%s/data_bags/example", dirName))
			if err != nil {
				return errors.Wrap(err, "Could not remove pre-created repo content: example data bag")
			}

			capturer := reporting.NewNodeCapturer(
				chefClient.Nodes, chefClient.Roles,
				chefClient.Environments, chefClient.Cookbooks,
				&reporting.ObjectWriter{RootDir: dirName},
			)
			nc := reporting.NewNodeCapture(nodeName, dirName, capturer)
			go nc.Run()
			for progress := range nc.Progress {
				switch progress {
				case reporting.FetchingNode:
					fmt.Printf(" - Capturing node object '%s'\n", nodeName)
				case reporting.FetchingCookbooks:
					fmt.Println(" - Capturing cookbooks...")
				case reporting.FetchingRoles:
					fmt.Println(" - Capturing roles...")
				case reporting.FetchingEnvironment:
					fmt.Println(" - Capturing environment...")
				case reporting.WritingKitchenConfig:
					fmt.Println(" - Writing kitchen configuration...")
				case reporting.FetchingComplete:
				}
			}
			if nc.Error != nil {
				return nc.Error
			}

			cookbooks := nc.Cookbooks
			path := requestCookbookPathFullMessage(dirName, cookbooks)

			// Try to gather sources for all cookbooks; stop when we've found
			// them all or the operator provides a blank input when asked for a new
			// path.
			for ok := true; ok; ok = (path != "" && len(cookbooks) > 0) {
				cookbooks, err = resolveCookbooks(cookbooks, dirName, path)
				if err != nil {
					return err
				}
				path = requestCookbookPath(cookbooks)
			}

			if len(cookbooks) > 0 {
				fmt.Printf(CookbooksNotSourcedTxt, fmt.Sprintf("%s/%s", dirName, "cookbooks"),
					formatCookbooks(cookbooks),
				)
			}

			fmt.Print(CookbookCaptureCompleteTxt)
			return nil
		},
	}
)

func init() {
	addInfraFlagsToCommand(captureCmd)
}
func requestCookbookPathFullMessage(repoPath string, cookbooks []reporting.NodeCookbook) string {
	fmt.Printf(CookbookCaptureGatherSourcesTxt, repoPath, formatCookbooks(cookbooks))
	var path string
	fmt.Scanf("%s\n", &path)
	return path
}

func requestCookbookPath(cookbooks []reporting.NodeCookbook) string {
	fmt.Printf(CookbookCaptureRequestCookbookPathTxt, formatCookbooks(cookbooks))
	var path string
	fmt.Scanf("%s\n", &path)
	return path
}

func formatCookbooks(cookbooks []reporting.NodeCookbook) string {
	var cbNames []string
	for _, cb := range cookbooks {
		cbNames = append(cbNames, fmt.Sprintf("  - %s (v%s)", cb.Name, cb.Version))
	}

	return strings.Join(cbNames, "\n")
}

// For a given `sourcePath`, replace copy of cookbook in repoDir/cookbooks/CB
// with a symlink to cookbooks found in the provided source.
// TODO: In the unlikely event that user directories are using FAT
//       symlinking will not work.
// TODO (future) - split terminal IO, and move this into an appropriate package.
func resolveCookbooks(cookbooks []reporting.NodeCookbook, repoDir string, sourcePath string) ([]reporting.NodeCookbook, error) {
	if repoDir == "" {
		return cookbooks, nil
	}

	var unresolved []reporting.NodeCookbook
	for _, cb := range cookbooks {
		targetPath, _ := filepath.Abs(fmt.Sprintf("%s/cookbooks/%s", repoDir, cb.Name))
		sourcePath, _ := filepath.Abs(fmt.Sprintf("%s/%s", sourcePath, cb.Name))
		if _, err := os.Stat(sourcePath); err == nil {
			fmt.Printf("  Using your checked-out cookbook: %s\n", cb.Name)
			savedTargetPath := fmt.Sprintf("%s.server", targetPath)
			err := os.Rename(targetPath, savedTargetPath)
			if err != nil {
				return nil, errors.Wrapf(err, "could not rename unsourced cookbook %s", targetPath)
			}
			os.Symlink(sourcePath, targetPath)
			if err != nil {
				return nil, errors.Wrapf(err, "Could not link %s to %s", sourcePath, targetPath)
			}

		} else {
			unresolved = append(unresolved, cb)
		}
	}

	return unresolved, nil
}
