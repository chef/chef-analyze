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
	"fmt"
	"os"
	"strings"

	chef "github.com/chef/go-chef"
	"github.com/pkg/errors"
)

type NodeCapture struct {
	Progress      chan int
	Error         error
	name          string
	repositoryDir string
	nodes         NodesInterface
	roles         RolesInterface
	env           EnvironmentInterface
	cookbooks     CookbookInterface
	writer        ObjectWriter
}

// Events that we publish on the Progress channel when
// performing a node capture
const (
	FetchingNode = iota
	FetchingEnvironment
	FetchingRoles
	FetchingCookbooks
	FetchingComplete
)

// Initialize a NodeCapture. Creates a Progress channel that callers can monitor for updates
func NewNodeCapture(nodeName string, repositoryDir string, nodes NodesInterface,
	roles RolesInterface, env EnvironmentInterface, cookbooks CookbookInterface, writer ObjectWriter) *NodeCapture {
	return &NodeCapture{name: nodeName, nodes: nodes, roles: roles,
		env: env, cookbooks: cookbooks, repositoryDir: repositoryDir,
		writer: writer,
		// 5 possible events in a Run - let's not block our activity in case the caller
		// doesn't pick them up
		Progress: make(chan int, 5)}
}

func (nc *NodeCapture) Run() {
	defer func() { close(nc.Progress) }()

	nc.Progress <- FetchingNode
	node, err := nc.captureNodeObject()
	if err != nil {
		nc.Error = errors.Wrapf(err, "unable to capture node '%s'", nc.name)
		return
	}

	nc.Progress <- FetchingCookbooks
	err = nc.captureCookbookObjects(node.AutomaticAttributes["cookbooks"].(map[string]interface{}))
	if err != nil {
		nc.Error = errors.Wrapf(err, "unable to capture node cookbooks for '%s'", nc.name)
		return
	}

	nc.Progress <- FetchingEnvironment
	err = nc.captureEnvObject(node.Environment)
	if err != nil {
		nc.Error = errors.Wrapf(err, "unable to capture environment")
		return
	}

	nc.Progress <- FetchingRoles
	err = nc.captureRoleObjects(node.RunList)
	if err != nil {
		nc.Error = errors.Wrapf(err, "unable to capture role(s)")
		return
	}

	nc.Progress <- FetchingComplete
}

// Capture the the node nc.Name from Chef Server to
// repositoryDir/nodes/NAME.json
// Returns the chef.Node object.
func (nc *NodeCapture) captureNodeObject() (*chef.Node, error) {
	node, err := nc.nodes.Get(nc.name)
	if err != nil {
		return nil, errors.Wrapf(err, "unable  retrieve node: %s", nc.name)
	}
	err = nc.writer.WriteNode(&node)
	if err != nil {
		return nil, errors.Wrap(err, "failed to save node. This is a bug, please report it.")
	}

	// In this pass we are not supporting policyfiles. For now, that
	// means we'll raise an error when the node has policyfile enabled,
	// to avoid the risk of reporting back inaccurate results based on non-PF assumptions.
	if node.PolicyName != "" || node.PolicyGroup != "" {
		return nil, errors.New(fmt.Sprintf("Node %s is managed by Policyfile. Unsupported at this time.", nc.name))
	}
	return &node, nil
}

// Given a map of cookbook [ name ] : { "version" : version }, download the cookbooks
// from Chef Server into repositoryDir/cookbooks
func (nc *NodeCapture) captureCookbookObjects(cookbooks map[string]interface{}) error {
	cookbookDir := fmt.Sprintf("%s/cookbooks", nc.repositoryDir)
	for name, version_data := range cookbooks {
		version := safeStringFromMap(version_data.(map[string]interface{}), "version")
		err := nc.cookbooks.DownloadTo(name, version, cookbookDir)
		if err != nil {
			return errors.Wrapf(err, "Failed to download cookbook %s v%s", name, version)
		}

		// The cookbook directory will have been created as 'name-version'.  We need it to be just 'name'
		// so that it can be picked up by chef-zero when requested from within kitchen.
		err = os.Rename(fmt.Sprintf("%s/%s-%s", cookbookDir, name, version), fmt.Sprintf("%s/%s", cookbookDir, name))
		if err != nil {
			return errors.Wrapf(err, "failed to rename cookbook %s-%s to non-versioned name", name, version)
		}
	}
	return nil
}

func (nc *NodeCapture) captureEnvObject(environment string) error {
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

func (nc *NodeCapture) captureRoleObjects(runList []string) error {

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

// filters a run list, return an array of roles
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
