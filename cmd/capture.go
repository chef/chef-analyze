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
	chef "github.com/chef/go-chef"
	"github.com/chef/go-libs/credentials"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	// behavior:
	//  - repo dir: if it does not exist, we will create it by invoking chef generate repo node-NAME-repo.
	//  - capture json object definitions for node,  and (env, roles) or (policy_group, active-policy-revision) .
	//    Save them to the chef-repo.
	//  - for policy-managed nodes, capture cookbook-artifacts from server based on policy
	//    and policy_group assigned to node. Save each artifact to the chef-repo.
	//  - for non-policy, capture cookbooks and versions from automatic attributes. Save each
	//    cookbook @ version to the chef-repo.
	//  - Offer user chance to symlink in cookbooks/cookbook artifacts from one or more locations
	//    on the Workstation system.

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
			repoDirName := fmt.Sprintf("./%s", repoName)

			// abort if it exists, have them remove it first.
			_, err = os.Stat(repoDirName)
			if err == nil {
				fmt.Printf(RepositoryAlreadyExistsE002, repoDirName, nodeName)
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
			err = os.RemoveAll(fmt.Sprintf("%s/cookbooks/example", repoDirName))
			if err != nil {
				return errors.Wrap(err, "Could not remove pre-created repo content: example cookbook")
			}

			err = os.RemoveAll(fmt.Sprintf("%s/data_bags/example", repoDirName))
			if err != nil {
				return errors.Wrap(err, "Could not remove pre-created repo content: example data bag")
			}

			capturer := reporting.NewNodeCapturer(
				chefClient.Nodes, chefClient.Roles,
				chefClient.Environments, chefClient.Cookbooks,
				chefClient.DataBags,
				chefClient.PolicyGroups,
				chefClient.Policies,
				chefClient.CookbookArtifacts,
				&reporting.ObjectWriter{RootDir: repoDirName},
			)
			nc := reporting.NewNodeCapture(nodeName, repoDirName, capturer)
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
				case reporting.FetchingCookbookArtifacts:
					fmt.Println(" - Capturing cookbook artifacts...")
				case reporting.FetchingPolicyData:
					fmt.Println(" - Capturing policy data...")

				case reporting.FetchingComplete:
				}
			}
			if nc.Error != nil {
				return nc.Error
			}

			var remainingCBs []reporting.NodeCookbook
			var cookbookDirName string
			if nc.Policy == nil {
				remainingCBs = nc.Cookbooks
				cookbookDirName = "cookbooks"
			} else {
				remainingCBs = locksToCookbooks(nc.Policy.CookbookLocks)
				cookbookDirName = "cookbook_artifacts"
			}

			// Try to gather sources for all cookbooks; stop when we've found
			// them all or the operator provides a blank input when asked for a new
			// path.
			requestGatherSources(repoDirName)
			baseUserPath := requestCookbookPath(remainingCBs)
			for len(remainingCBs) > 0 && baseUserPath != "" {
				remainingCBs, err = resolveCookbooks(cookbookDirName, remainingCBs, repoDirName, baseUserPath)
				if err != nil {
					return err
				}
				baseUserPath = requestCookbookPath(remainingCBs)
			}

			if len(remainingCBs) > 0 {
				fmt.Printf(CookbooksNotSourcedTxt, fmt.Sprintf("%s/%s", repoDirName, cookbookDirName),
					formatCookbooks(remainingCBs))
			}

			fmt.Print(CookbookCaptureCompleteTxt, repoDirName)
			return nil

		},
	}
)

func locksToCookbooks(locks map[string]chef.CookbookLock) []reporting.NodeCookbook {
	var cbs = make([]reporting.NodeCookbook, 0)

	for name, lock := range locks {
		cbs = append(cbs, reporting.NodeCookbook{Name: name, Version: lock.Identifier})
	}
	return cbs
}

func init() {
	addInfraFlagsToCommand(captureCmd)
}

func requestCookbookPath(cookbooks []reporting.NodeCookbook) string {

	userPath := promptUser(fmt.Sprintf(CookbookCaptureRequestCookbookPathTxt, formatCookbooks(cookbooks)))
	for len(userPath) > 0 {
		absPath, _ := filepath.Abs(userPath)
		if _, err := os.Stat(absPath); err == nil {
			break
		}
		userPath = promptUser(fmt.Sprintf("'%s' is not a valid path. Cookbook Location [none]: ", userPath))
	}
	return userPath
}

func requestGatherSources(repoPath string) {
	fmt.Printf(CookbookCaptureGatherSourcesTxt, repoPath)
	promptUser(PressEnterToContinueTxt)
}

func promptUser(msg string) string {
	fmt.Print(msg)
	var answer string
	fmt.Scanf("%s\n", &answer)
	fmt.Println("") // Blank linke after taking input separates any response we may write
	return answer
}

func formatCookbooks(cookbooks []reporting.NodeCookbook) string {
	var cbNames []string
	for _, cb := range cookbooks {
		// TODO - do we want the hash shown for cookbook artifacts? Does it matter?
		cbNames = append(cbNames, fmt.Sprintf("  - %s (v%s)", cb.Name, cb.Version))
	}

	return strings.Join(cbNames, "\n")
}

// For a given `sourcePath`, replace copy of cookbook in repoDir/cookbooks/CB
// with a symlink to cookbooks found in the provided source.
// TODO: In the unlikely event that user directories are using FAT
//       symlinking will not work.
// TODO (future) - split terminal IO, and move this into an appropriate package.
func resolveCookbooks(cookbookDirName string, cookbooks []reporting.NodeCookbook, repoDir string, sourcePath string) ([]reporting.NodeCookbook, error) {
	if sourcePath == "" || repoDir == "" {
		return cookbooks, nil
	}

	var (
		targetPath string
		unresolved []reporting.NodeCookbook
	)
	foundCount := 0
	for _, cb := range cookbooks {
		fullSourcePath, _ := filepath.Abs(fmt.Sprintf("%s/%s", sourcePath, cb.Name))
		if cookbookDirName == "cookbooks" {
			targetPath, _ = filepath.Abs(fmt.Sprintf("%s/cookbooks/%s", repoDir, cb.Name))
		} else { // cookbook_artifacts
			targetPath, _ = filepath.Abs(fmt.Sprintf("%s/cookbook_artifacts/%s-%s", repoDir, cb.Name, cb.Version))
		}

		if _, err := os.Stat(fullSourcePath); err == nil {

			savedTargetPath := fmt.Sprintf("%s.server", targetPath)
			err := os.Rename(targetPath, savedTargetPath)
			if err != nil {
				return nil, errors.Wrapf(err, "could not rename unsourced cookbook %s to %s", targetPath, savedTargetPath)
			}
			err = os.Symlink(fullSourcePath, targetPath)
			if err != nil {
				return nil, errors.Wrapf(err, "Could not link %s to %s", fullSourcePath, targetPath)
			}

			foundCount += 1
		} else {
			unresolved = append(unresolved, cb)
		}

	}

	if len(unresolved) > 0 {
		if foundCount > 0 {
			fmt.Printf("  Found %d/%d cookbooks in %s\n", foundCount, len(cookbooks), sourcePath)
		} else {
			fmt.Printf("  No cookbooks found. Please check the path you provided: %s\n", sourcePath)
		}
	} else {
		fmt.Printf("  Found all remaining cookbooks in %s\n", sourcePath)
	}

	return unresolved, nil
}
