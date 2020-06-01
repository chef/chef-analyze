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

package reporting

import (
	"fmt"
	"io/ioutil"
	"os"

	chef "github.com/chef/go-chef"
	"github.com/pkg/errors"
)

type ObjectWriter struct {
	RootDir string
}

type ObjectWriterInterface interface {
	WriteRole(*chef.Role) error
	WriteEnvironment(*chef.Environment) error
	WritePolicyRevision(*chef.RevisionDetailsResponse) error
	WritePolicyGroup(string, *chef.PolicyGroup) error
	WriteNode(*chef.Node) error
	WriteContent(string, []byte) error
}

// Takes a chef.Role and saves it as json in RootDir/roles
func (ow *ObjectWriter) WriteRole(role *chef.Role) error {
	return ow.WriteJSON("roles", role.Name, role)
}

// Takes a chef.Role and saves it as json in RootDir/environments
func (ow *ObjectWriter) WriteEnvironment(env *chef.Environment) error {
	return ow.WriteJSON("environments", env.Name, env)
}

// Takes a chef.Role and saves it as json in RootDir/nodes
func (ow *ObjectWriter) WriteNode(node *chef.Node) error {
	return ow.WriteJSON("nodes", node.Name, node)
}

func (ow *ObjectWriter) WritePolicyRevision(policy *chef.RevisionDetailsResponse) error {
	return ow.WriteJSON("policies", fmt.Sprintf("%s-%s", policy.Name, policy.RevisionID), policy)
}

func (ow *ObjectWriter) WritePolicyGroup(name string, pgroup *chef.PolicyGroup) error {
	return ow.WriteJSON("policy_groups", name, pgroup)
}

// Writes "content" to RootDir/"fileName"
func (ow *ObjectWriter) WriteContent(fileName string, content []byte) error {
	path := fmt.Sprintf("%s/%s", ow.RootDir, fileName)
	f, err := os.Create(path)
	if err != nil {
		return errors.Wrapf(err, "failed to create %s", path)
	}
	_, err = f.Write(content)
	if err != nil {
		return errors.Wrapf(err, "failed to write to %s", path)
	}
	return nil
}

func (ow *ObjectWriter) WriteJSON(objGroupingName string, objName string, object interface{}) error {
	var err error
	dirName := fmt.Sprintf("%s/%s", ow.RootDir, objGroupingName)
	path := fmt.Sprintf("%s/%s/%s.json", ow.RootDir, objGroupingName, objName)
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		err = os.MkdirAll(dirName, 0700)
	}
	if err != nil {
		return errors.Wrapf(err, "Failed to create directory %s", dirName)
	}
	body, err := chef.JSONReader(object)
	if err != nil {
		return errors.Wrapf(err, "Failed to convert %s to json. If you see this, it's a bug.  Please report it.", objName)
	}

	content, err := ioutil.ReadAll(body)
	if err != nil {
		return errors.Wrapf(err, "Failed to read %s %s", objGroupingName, objName)
	}

	err = ioutil.WriteFile(path, content, 0600)
	if err != nil {
		return errors.Wrapf(err, "Failed to save file %s", path)
	}
	return nil
}
