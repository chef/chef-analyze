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

package reporting_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	subject "github.com/chef/chef-analyze/pkg/reporting"
	chef "github.com/chef/go-chef"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type CapturerMock struct {
	NodeReturn          *chef.Node
	NodeErrorReturn     error
	EnvErrorReturn      error
	RoleErrorReturn     error
	CookbookErrorReturn error
}

func (cm *CapturerMock) CaptureCookbooks(string, map[string]interface{}) error {
	return cm.CookbookErrorReturn
}

func (cm *CapturerMock) CaptureNodeObject(node string) (*chef.Node, error) {
	return cm.NodeReturn, cm.NodeErrorReturn
}
func (cm *CapturerMock) CaptureEnvObject(string) error {
	return cm.EnvErrorReturn
}
func (cm *CapturerMock) CaptureRoleObjects([]string) error {
	return cm.RoleErrorReturn
}
func defaultNode() *chef.Node {
	return &chef.Node{
		Name:        "node1",
		Environment: "_default",
		RunList:     []string{"cookbook1::recipe1", "role:mockrole"},
		PolicyName:  "",
		PolicyGroup: "",
		AutomaticAttributes: map[string]interface{}{
			"cookbooks": map[string]interface{}{
				"foo": map[string]interface{}{
					"version": "0.1.0",
				},
			},
		},
	}
}

func TestCapture_Run(t *testing.T) {
	baseDir, err := ioutil.TempDir(os.TempDir(), "chefanalyze-unit*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(baseDir)
	nc := subject.NewNodeCapture("node1", baseDir, &CapturerMock{NodeReturn: defaultNode()})
	go nc.Run()
	val := <-nc.Progress
	assert.Equal(t, subject.FetchingNode, val)
	val = <-nc.Progress
	assert.Equal(t, subject.FetchingCookbooks, val)
	val = <-nc.Progress
	assert.Equal(t, subject.FetchingEnvironment, val)
	val = <-nc.Progress
	assert.Equal(t, subject.FetchingRoles, val)
	val = <-nc.Progress
	assert.Equal(t, subject.FetchingComplete, val)

	assert.Nil(t, nc.Error)

}

func TestCapture_RunWithNodeFailure(t *testing.T) {
	nc := subject.NewNodeCapture("node1", "", &CapturerMock{NodeErrorReturn: errors.New("no node")})
	nc.Run()
	if assert.NotNil(t, nc.Error) {
		assert.Contains(t, nc.Error.Error(), "no node")
		assert.Contains(t, nc.Error.Error(), "unable to capture node 'node1'")
	}
}

func TestCapture_RunWithNoCookbooksAvailable(t *testing.T) {
	nc := subject.NewNodeCapture("node1", "", &CapturerMock{NodeReturn: &chef.Node{}})
	nc.Run()
	assert.Nil(t, nc.Error)
}

func TestCapture_RunWithCookbookFailure(t *testing.T) {
	nc := subject.NewNodeCapture("node1", "", &CapturerMock{NodeReturn: defaultNode(),
		CookbookErrorReturn: errors.New("no cookbook")})
	nc.Run()
	if assert.NotNil(t, nc.Error) {
		assert.Contains(t, nc.Error.Error(), "no cookbook")
		assert.Contains(t, nc.Error.Error(), "unable to capture node cookbooks")
	}
}

func TestCapture_RunWithEnvironmentFailure(t *testing.T) {
	nc := subject.NewNodeCapture("node1", "", &CapturerMock{NodeReturn: defaultNode(),
		EnvErrorReturn: errors.New("no env")})
	nc.Run()
	if assert.NotNil(t, nc.Error) {
		assert.Contains(t, nc.Error.Error(), "no env")
		assert.Contains(t, nc.Error.Error(), "unable to capture environment")
	}
}

func TestCapture_RunWithRoleFailure(t *testing.T) {
	nc := subject.NewNodeCapture("node1", "", &CapturerMock{NodeReturn: defaultNode(),
		RoleErrorReturn: errors.New("no role")})
	nc.Run()
	if assert.NotNil(t, nc.Error) {
		assert.Contains(t, nc.Error.Error(), "no role")
		assert.Contains(t, nc.Error.Error(), "unable to capture role")
	}
}

func TestCapturer_CaptureNodeObject(t *testing.T) {
	expectedNode := chef.Node{
		Name:        "node1",
		Environment: "_default",
		RunList:     []string{"cookbook1::recipe1", "role:mockrole"},
		PolicyName:  "",
		PolicyGroup: "",
	}
	writer := ObjectWriterMock{}
	nc := subject.NewNodeCapturer(
		NodeMock{Error: nil, Node: expectedNode},
		RoleMock{},
		EnvMock{},
		CookbookMock{},
		&writer,
	)
	actualNode, err := nc.CaptureNodeObject("node1")
	assert.Equal(t, &expectedNode, actualNode)
	assert.Equal(t, nil, err)
}

func TestCapturer_CaptureNodeObjectWithPolicyError(t *testing.T) {
	expectedNode := chef.Node{
		Name:        "node1",
		Environment: "_default",
		RunList:     []string{"cookbook1::recipe1", "role:mockrole"},
		PolicyName:  "",
		PolicyGroup: "mypolicygroup",
	}
	writer := ObjectWriterMock{}
	nc := subject.NewNodeCapturer(
		NodeMock{Error: nil, Node: expectedNode},
		RoleMock{},
		EnvMock{},
		CookbookMock{},
		&writer,
	)
	actualNode, err := nc.CaptureNodeObject("node1")
	assert.Nil(t, actualNode)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "managed by Policyfile")
	}
}

func TestCapturer_CaptureNodeObjectWithErrorOnFetch(t *testing.T) {
	underlyingErr := errors.New("fetch error")
	writer := ObjectWriterMock{}
	nc := subject.NewNodeCapturer(
		NodeMock{Error: underlyingErr},
		RoleMock{},
		EnvMock{},
		CookbookMock{},
		&writer,
	)
	_, err := nc.CaptureNodeObject("node1")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "fetch error") // because it's wrapped
}

func TestCapturer_CaptureNodeObjectWithErrorOnSave(t *testing.T) {
	node := chef.Node{}
	writer := ObjectWriterMock{Error: errors.New("failed to write")}
	nc := subject.NewNodeCapturer(
		NodeMock{Node: node},
		RoleMock{},
		EnvMock{},
		CookbookMock{},
		&writer,
	)
	_, err := nc.CaptureNodeObject("node1")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "failed to write")
	}
}

func TestCapturer_CaptureRoleObjects(t *testing.T) {
	expectedRole := chef.Role{
		Name: "role1",
	}
	writer := ObjectWriterMock{}
	nc := subject.NewNodeCapturer(
		NodeMock{},
		RoleMock{Role: &expectedRole},
		EnvMock{},
		CookbookMock{},
		&writer,
	)
	err := nc.CaptureRoleObjects([]string{"role[role1]"})
	assert.Equal(t, &expectedRole, writer.ReceivedObject)
	assert.Nil(t, err)
}

func TestCapturer_CaptureRoleObjectsWithErrorOnFetch(t *testing.T) {
	writer := ObjectWriterMock{}
	nc := subject.NewNodeCapturer(
		NodeMock{},
		RoleMock{Error: errors.New("failed to fetch")},
		EnvMock{},
		CookbookMock{},
		&writer,
	)
	err := nc.CaptureRoleObjects([]string{"role[role1]"})
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "unable to retrieve")
		assert.Contains(t, err.Error(), "failed to fetch")
	}
}

func TestCapturer_CaptureRoleObjectsWithErrorOnSave(t *testing.T) {
	writer := ObjectWriterMock{Error: errors.New("failed to write")}
	nc := subject.NewNodeCapturer(
		NodeMock{},
		RoleMock{},
		EnvMock{},
		CookbookMock{},
		&writer,
	)
	err := nc.CaptureRoleObjects([]string{"role[role1]"})
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "failed to save role")
		assert.Contains(t, err.Error(), "failed to write")
	}
}

func TestCapturer_CaptureEnvObject(t *testing.T) {
	expectedEnv := chef.Environment{
		Name: "env1",
	}
	writer := ObjectWriterMock{}
	nc := subject.NewNodeCapturer(
		NodeMock{},
		RoleMock{},
		EnvMock{Env: &expectedEnv},
		CookbookMock{},
		&writer,
	)
	err := nc.CaptureEnvObject("env1")
	assert.Equal(t, writer.ReceivedObject, &expectedEnv)
	assert.Nil(t, err)
}

func TestCapturer_CaptureEnvObjectWithErrorOnFetch(t *testing.T) {
	writer := ObjectWriterMock{}
	nc := subject.NewNodeCapturer(
		NodeMock{},
		RoleMock{},
		EnvMock{Error: errors.New("failed to fetch")},
		CookbookMock{},
		&writer,
	)
	err := nc.CaptureEnvObject("env1")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to fetch")
}

func TestCapturer_CaptureEnvObjectWithErrorOnSave(t *testing.T) {
	writer := ObjectWriterMock{Error: errors.New("failed to write")}
	nc := subject.NewNodeCapturer(
		NodeMock{},
		RoleMock{},
		EnvMock{},
		CookbookMock{},
		&writer,
	)
	err := nc.CaptureEnvObject("env1")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to write")

}

func TestCapturer_CaptureCookbooks(t *testing.T) {
	baseDir, err := ioutil.TempDir(os.TempDir(), "chefanalyze-unit*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(baseDir)

	cookbookList := chef.CookbookListResult{
		"foo": chef.CookbookVersions{
			Versions: []chef.CookbookVersion{
				chef.CookbookVersion{Version: "0.1.0"},
			},
		},
	}

	writer := ObjectWriterMock{}
	nc := subject.NewNodeCapturer(
		NodeMock{},
		RoleMock{},
		EnvMock{},
		newMockCookbook(cookbookList, nil, nil),
		&writer,
	)

	cbmap := map[string]interface{}{
		"foo": map[string]interface{}{
			"version": "0.1.0",
		},
	}

	err = nc.CaptureCookbooks(baseDir, cbmap)
	assert.Nil(t, err)
}

func TestCapturer_CaptureCookbooksWithDownloadError(t *testing.T) {
	baseDir, err := ioutil.TempDir(os.TempDir(), "chefanalyze-unit*")
	if err != nil {
		panic(err)
	}
	assert.Nil(t, err)
	defer os.RemoveAll(baseDir)

	cookbookList := chef.CookbookListResult{
		"foo": chef.CookbookVersions{
			Versions: []chef.CookbookVersion{
				chef.CookbookVersion{Version: "0.1.0"},
			},
		},
	}

	writer := ObjectWriterMock{}
	nc := subject.NewNodeCapturer(
		NodeMock{},
		RoleMock{},
		EnvMock{},
		newMockCookbook(cookbookList, nil, errors.New("download error in test1")),
		&writer,
	)

	cookbookPath := fmt.Sprintf("%s/cookbooks/foo-0.1.0", baseDir)
	// pre-setup to ensure that the rename of the downloaded cookbook (foo-0.1.0 to foo)
	// will succeed; required because the cookbook mock doesn't perform an actual download
	err = os.MkdirAll(cookbookPath, 0755)
	if err != nil {
		panic(err)
	}

	cbmap := map[string]interface{}{
		"foo": map[string]interface{}{
			"version": "0.1.0",
		},
	}

	err = nc.CaptureCookbooks(baseDir, cbmap)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "download error in test1")
		assert.Contains(t, err.Error(), "Failed to download cookbook")
	}
}
func TestCapturer_CaptureCookbooksWithRenameError(t *testing.T) {
	baseDir, err := ioutil.TempDir(os.TempDir(), "chefanalyze-unit*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(baseDir)

	cookbookList := chef.CookbookListResult{
		"foo": chef.CookbookVersions{
			Versions: []chef.CookbookVersion{
				chef.CookbookVersion{Version: "0.1.0"},
			},
		},
	}

	writer := ObjectWriterMock{}
	cbMock := newMockCookbook(cookbookList, nil, nil)
	cbMock.createDirOnDownload = false
	nc := subject.NewNodeCapturer(
		NodeMock{},
		RoleMock{},
		EnvMock{},
		cbMock,
		&writer,
	)

	// To force this error, we will not manually create the versioned cookbook
	// directory name that a real download cookbooks would do. This will cause the
	// rename to fail since the directory to be renamed won't exist.

	cbmap := map[string]interface{}{
		"foo": map[string]interface{}{
			"version": "0.1.0",
		},
	}

	err = nc.CaptureCookbooks(baseDir, cbmap)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "failed to rename cookbook")
	}
}
