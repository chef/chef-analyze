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

// TODO - puting this in reporting for now, because
// of common bits already in that package.  Longer-term
// we'll want to move the common bits to their own package
// (credentials handling, common object structures like chef client)
package reporting

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/go-chef/chef"
	"github.com/pkg/errors"
)

type NodeCaptureInterface interface {
	CaptureCookbooks(string, map[string]interface{}) ([]NodeCookbook, error)
	CaptureCookbookArtifacts(string, *chef.RevisionDetailsResponse) error
	CaptureNodeObject(node string) (*chef.Node, error)
	CaptureEnvObject(string) error
	CaptureRoleObjects([]string) error
	CapturePolicyObject(string, string) (*chef.RevisionDetailsResponse, error)
	CapturePolicyGroupObject(string) (*chef.PolicyGroup, error)
	CaptureAllDataBagItems() error
	SaveKitchenYML(node *chef.Node) error
}

type NodeCapture struct {
	Cookbooks     []NodeCookbook
	Policy        *chef.RevisionDetailsResponse
	name          string
	capturer      NodeCaptureInterface
	node          *chef.Node
	repositoryDir string
	opts          CaptureOpts
	Progress      chan int
	Error         error
}

type NodeCapturer struct {
	cookbookArtifacts CBAInterface
	cookbooks         CookbookInterface
	dataBags          DataBagInterface
	env               EnvironmentInterface
	nodes             NodesInterface
	policies          PolicyInterface
	policyGroups      PolicyGroupInterface
	roles             RolesInterface
	writer            ObjectWriterInterface
}

type NodeCookbook struct {
	Name    string
	Version string
}

// Represents platform data extracted from a node's
// Automatic attributes.
type nodePlatformData struct {
	OS      string
	Name    string
	Family  string
	Version string
}

// Struct for providing arguments to the
// kitchen.yml template
type kitchenYMLArgs struct {
	ChefVersion string
	NodeName    string
	Image       string
}

// Events that we publish on the Progress channel when
// performing a node capture
const (
	FetchingNode = iota
	FetchingPolicyData
	FetchingEnvironment
	FetchingRoles
	FetchingDataBags
	FetchingCookbooks
	FetchingCookbookArtifacts
	WritingKitchenConfig
	CaptureComplete
)

type CaptureOpts struct {
	DownloadDataBags bool
}

func NewNodeCapture(name string, repositoryDir string, opts CaptureOpts, capturer NodeCaptureInterface) *NodeCapture {
	return &NodeCapture{
		name:          name,
		capturer:      capturer,
		repositoryDir: repositoryDir,
		opts:          opts,
		// 6 max possible events in a Run - let's not block our activity in case the caller
		// doesn't pick them up
		Progress: make(chan int, 7),
	}
}

// Initialize a NodeCapture. Creates a Progress channel that callers can monitor for updates
func (nc *NodeCapture) Run() {
	defer func() { close(nc.Progress) }()

	nc.Progress <- FetchingNode
	node, err := nc.capturer.CaptureNodeObject(nc.name)
	if err != nil {
		nc.Error = errors.Wrapf(err, "unable to capture node '%s'", nc.name)
		return
	}

	if nc.opts.DownloadDataBags {
		nc.Progress <- FetchingDataBags
		err = nc.capturer.CaptureAllDataBagItems()
		if err != nil {
			nc.Error = errors.Wrapf(err, "unable to capture data bag items")
			return
		}
	}

	if len(node.PolicyName) > 0 {
		nc.Progress <- FetchingPolicyData
		group, err := nc.capturer.CapturePolicyGroupObject(node.PolicyGroup)
		if err != nil {
			nc.Error = errors.Wrapf(err, "unable to capture policy group '%s'", node.PolicyGroup)
			return
		}

		policyRevision := group.Policies[node.PolicyName]["revision_id"]
		policy, err := nc.capturer.CapturePolicyObject(node.PolicyName, policyRevision)
		if err != nil {
			nc.Error = errors.Wrapf(err, "unable to capture policy name '%s' revision '%s'", node.PolicyName, policyRevision)
			return
		}
		nc.Policy = policy

		nc.Progress <- FetchingCookbookArtifacts
		err = nc.capturer.CaptureCookbookArtifacts(nc.repositoryDir, policy)
		if err != nil {
			nc.Error = errors.Wrapf(err, "unable to capture cookbook artifacts for policy '%s' revision '%s'", node.PolicyName, policyRevision)
			return
		}

	} else {
		// Cookbooks/Env/Roles only apply to traditionally managed
		// nodes, and not policy-managed.
		nc.Progress <- FetchingCookbooks
		// If a node has never converged, it will not have this attribute:
		cookbooks := node.AutomaticAttributes["cookbooks"]
		if cookbooks != nil {
			nc.Cookbooks, err = nc.capturer.CaptureCookbooks(nc.repositoryDir, cookbooks.(map[string]interface{}))
			if err != nil {
				nc.Error = errors.Wrapf(err, "unable to capture node cookbooks for '%s'", nc.name)
				return
			}
		}

		nc.Progress <- FetchingEnvironment
		err = nc.capturer.CaptureEnvObject(node.Environment)
		if err != nil {
			nc.Error = errors.Wrapf(err, "unable to capture environment")
			return
		}

		nc.Progress <- FetchingRoles
		err = nc.capturer.CaptureRoleObjects(node.RunList)
		if err != nil {
			nc.Error = errors.Wrapf(err, "unable to capture role(s)")
			return
		}
	}

	nc.Progress <- WritingKitchenConfig
	err = nc.capturer.SaveKitchenYML(node)
	if err != nil {
		nc.Error = errors.Wrapf(err, "unable to write Kitchen config")
	}
	nc.Progress <- CaptureComplete
}

func NewNodeCapturer(
	nodes NodesInterface,
	roles RolesInterface,
	env EnvironmentInterface,
	cookbooks CookbookInterface,
	dataBags DataBagInterface,
	policyGroups PolicyGroupInterface,
	policies PolicyInterface,
	cookbookArtifacts CBAInterface,
	writer ObjectWriterInterface) *NodeCapturer {
	return &NodeCapturer{
		nodes:             nodes,
		roles:             roles,
		env:               env,
		cookbooks:         cookbooks,
		dataBags:          dataBags,
		policies:          policies,
		policyGroups:      policyGroups,
		cookbookArtifacts: cookbookArtifacts,
		writer:            writer}

}

// Capture the the node nc.Name from Chef Server to
// repositoryDir/nodes/NAME.json
// Returns the chef.Node object.

func (nc *NodeCapturer) CaptureNodeObject(name string) (*chef.Node, error) {
	node, err := nc.nodes.Get(name)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to retrieve node '%s'", name)
	}
	err = nc.writer.WriteNode(&node)
	if err != nil {
		return nil, errors.Wrap(err, "failed to save node. This is a bug, please report it.")
	}

	return &node, nil
}

// Given a map of cookbook [ name ] : { "version" : version }, download the cookbooks
// from Chef Server into repositoryDir/cookbooks
func (nc *NodeCapturer) CaptureCookbookArtifacts(repositoryDir string, policyRevisionDetails *chef.RevisionDetailsResponse) error {
	artifactDir := fmt.Sprintf("%s/cookbook_artifacts", repositoryDir)

	for ck, cv := range policyRevisionDetails.CookbookLocks {
		err := nc.cookbookArtifacts.DownloadTo(ck, cv.Identifier, artifactDir)
		if err != nil {
			return errors.Wrapf(err, "Failed to download cookbook artifacts %ss v%s", ck, cv.Identifier)
		}

		// The cookbook directory wil have been created as Name-TruncatedCookbookIdent.  chef-zero requires
		// it in the form Name-CookbookIdent, so we'll move it here.
		oldName := fmt.Sprintf("%s/%s-%s", artifactDir, ck, cv.Identifier[0:20])
		newName := fmt.Sprintf("%s/%s-%s", artifactDir, ck, cv.Identifier)
		err = os.Rename(oldName, newName)
		if err != nil {
			return errors.Wrapf(err, "failed to rename cookbook artifact '%s' to '%s'", oldName, newName)
		}
	}
	return nil
}

// Given a map of cookbook [ name ] : { "version" : version }, download the cookbooks
// from Chef Server into repositoryDir/cookbooks
func (nc *NodeCapturer) CaptureCookbooks(repositoryDir string, cookbooks map[string]interface{}) ([]NodeCookbook, error) {
	var outCookbooks []NodeCookbook
	cookbookDir := fmt.Sprintf("%s/cookbooks", repositoryDir)

	for name, version_data := range cookbooks {
		version := safeStringFromMap(version_data.(map[string]interface{}), "version")
		err := nc.cookbooks.DownloadTo(name, version, cookbookDir)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to download cookbook %s v%s", name, version)
		}
		outCookbooks = append(outCookbooks, NodeCookbook{name, version})
		// The cookbook directory will have been created as 'name-version'.  We need it to be just 'name'
		// so that it can be picked up by chef-zero when requested from within kitchen.
		err = os.Rename(fmt.Sprintf("%s/%s-%s", cookbookDir, name, version), fmt.Sprintf("%s/%s", cookbookDir, name))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to rename cookbook %s-%s to non-versioned name", name, version)
		}
	}
	return outCookbooks, nil
}

func (nc *NodeCapturer) CaptureEnvObject(environment string) error {
	env, err := nc.env.Get(environment)
	if err != nil {
		return errors.Wrapf(err, "unable to retrieve node environment: %s", environment)
	}

	err = nc.writer.WriteEnvironment(env)
	if err != nil {
		return errors.Wrap(err, "failed to save environment. This is a bug, please report it.")
	}
	return nil
}

func (nc *NodeCapturer) CapturePolicyObject(policyName string, revision string) (*chef.RevisionDetailsResponse, error) {

	policy, err := nc.policies.GetRevisionDetails(policyName, revision)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to retrieve policy object: %s", policyName)
	}
	err = nc.writer.WritePolicyRevision(&policy)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to save policy object to disk: %s", policyName)
	}

	return &policy, nil
}

func (nc *NodeCapturer) CapturePolicyGroupObject(groupName string) (*chef.PolicyGroup, error) {
	groups, err := nc.policyGroups.List()
	// TODO - policy group API support does not allow us to get a specific
	// named group. We need to get them all, look up the one we care about in
	// the response
	if err != nil {
		return nil, errors.Wrapf(err, "unable to retrieve policy groups")
	}

	group, ok := groups[groupName]
	if !ok {
		return nil, fmt.Errorf("Could not find policy group: %s", groupName)
	}

	err = nc.writer.WritePolicyGroup(groupName, &group)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to save policy group %s to disk", groupName)
	}
	return &group, nil
}

func (nc *NodeCapturer) CaptureRoleObjects(runList []string) error {
	roleNames := filterRoles(runList)
	for _, roleName := range roleNames {
		role, err := nc.roles.Get(roleName)
		if err != nil {
			return errors.Wrapf(err, "unable to retrieve role: %s", roleName)
		}
		err = nc.writer.WriteRole(role)
		if err != nil {
			return errors.Wrapf(err, "failed to save role %s. This is a bug, please report it:", roleName)
		}
	}
	return nil
}

func (nc *NodeCapturer) CaptureAllDataBagItems() error {
	bags, err := nc.dataBags.List()
	if err != nil {
		return errors.Wrap(err, "unable to retrieve list of data bags")
	}
	for name, _ := range *bags {
		items, err := nc.dataBags.ListItems(name)
		if err != nil {
			return errors.Wrapf(err, "unable to retrieve data bag items for %s", name)
		}

		for itemName, _ := range *items {
			item, err := nc.dataBags.GetItem(name, itemName)
			if err != nil {
				return errors.Wrapf(err, "unable to retrieve data bag item %s/%s", name, itemName)
			}

			err = nc.writer.WriteDataBagItem(name, itemName, item)
			if err != nil {
				return errors.Wrapf(err, "unable to save data bag item %s/%s", name, itemName)
			}

		}
	}
	return nil
}

func (nc *NodeCapturer) SaveKitchenYML(node *chef.Node) error {

	packages := node.AutomaticAttributes["chef_packages"]
	if packages == nil {
		return errors.New("could not determine chef client version: node missing automatic attribute 'chef_packages'")
	}

	chef := packages.(map[string]interface{})["chef"]
	if chef == nil {
		return errors.New("could not determine chef client version: node missing automatic attribute chef_packages['chef']")
	}

	version := safeStringFromMap(chef.(map[string]interface{}), "version")
	if version == "" {
		return errors.New("could not determine chef client version: node missing automatic attribute chef_packages['chef']['version']")
	}

	args := kitchenYMLArgs{version, node.Name, determineKitchenImage(node)}
	tmpl_text := `
---
driver:
  name: vagrant

provisioner:
  name: chef_zero_capture
  product_name: chef
  product_version: {{.ChefVersion}}
  json_attributes: false
  client_rb:
    node_name: {{.NodeName}}

platforms:
  - name: {{.Image}}

suites:
  - name: {{.NodeName}}
`
	tmpl, err := template.New("kitchen.yml").Parse(tmpl_text)
	if err != nil {
		return errors.Wrap(err, "could not create new template")
	}
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	err = tmpl.Execute(writer, args)
	writer.Flush()
	if err != nil {
		return errors.Wrap(err, "failed to execute template")
	}

	err = nc.writer.WriteContent("kitchen.yml", b.Bytes())
	if err != nil {
		return errors.Wrap(err, "failed to save")
	}
	return nil

}

func extractPlatformFromNode(node *chef.Node) nodePlatformData {
	return nodePlatformData{
		Name:    safeStringFromMap(node.AutomaticAttributes, "platform"),
		Family:  safeStringFromMap(node.AutomaticAttributes, "platform_family"),
		Version: safeStringFromMap(node.AutomaticAttributes, "platform_version"),
		OS:      safeStringFromMap(node.AutomaticAttributes, "os"),
	}
}

// filters a run list, return an array of role names
// as identified by a run list item being in the form "role[...]"
// Will return zero-size array if no roles are in the run list.
func filterRoles(possibleRoles []string) []string {
	roles := make([]string, 0)
	for _, role := range possibleRoles {
		pos := strings.Index(role, "role[")
		if pos == -1 {
			continue
		}
		// can't remember if chef server allows unicode in run lists,
		// so we'll play it safe and convert to runes.
		runes := []rune(role)
		stop := len(role) - 1
		roles = append(roles, string(runes[5:stop]))
	}

	return roles
}

// In a very limited way takes a guess at the name of the kitchen
// image to use for the node being captured. We will probably
// want to look closer at how and where this is invoked -
// ideally we nee dto provide for user entry of an image name if
// we can't resolve it, or resolve it incorrectly. This is also an
// opportunity to be smart about providers - for example if the original
// node is provided by AWS and we have an AMI we can use, we should do that.
func determineKitchenImage(node *chef.Node) string {
	platform := extractPlatformFromNode(node)
	return fmt.Sprintf("%s-%s", platform.Name, platform.Version)
}
