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

func TestCaptureRun(t *testing.T) {
	expectedNode := chef.Node{
		Name:        "node1",
		Environment: "_default",
		RunList:     []string{"cookbook1::recipe1", "role:mockrole"},
		PolicyName:  "",
		PolicyGroup: "",
	}
	writer := ObjectWriterMock{}
	nc := subject.NewNodeCapture("",
		"",
		NodeMock{Error: nil, Node: expectedNode},
		RoleMock{},
		EnvMock{},
		CookbookMock{},
		&writer,
	)
	actualNode, err := nc.CaptureNodeObject()
	assert.Equal(t, &expectedNode, actualNode)
	assert.Equal(t, &expectedNode, writer.ReceivedObject)
	assert.Equal(t, nil, err)
}

func TestCaptureNodeObject(t *testing.T) {
	expectedNode := chef.Node{
		Name:        "node1",
		Environment: "_default",
		RunList:     []string{"cookbook1::recipe1", "role:mockrole"},
		PolicyName:  "",
		PolicyGroup: "",
	}
	writer := ObjectWriterMock{}
	nc := subject.NewNodeCapture("",
		"",
		NodeMock{Error: nil, Node: expectedNode},
		RoleMock{},
		EnvMock{},
		CookbookMock{},
		&writer,
	)
	actualNode, err := nc.CaptureNodeObject()
	assert.Equal(t, &expectedNode, actualNode)
	assert.Equal(t, nil, err)
}

func TestCaptureNodeObjectWithPolicyError(t *testing.T) {
	expectedNode := chef.Node{
		Name:        "node1",
		Environment: "_default",
		RunList:     []string{"cookbook1::recipe1", "role:mockrole"},
		PolicyName:  "",
		PolicyGroup: "mypolicygroup",
	}
	writer := ObjectWriterMock{}
	nc := subject.NewNodeCapture("",
		"",
		NodeMock{Error: nil, Node: expectedNode},
		RoleMock{},
		EnvMock{},
		CookbookMock{},
		&writer,
	)
	actualNode, err := nc.CaptureNodeObject()
	assert.Nil(t, actualNode)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "managed by Policyfile")
	}
}

func TestCaptureNodeObjectWithErrorOnFetch(t *testing.T) {
	underlyingErr := errors.New("fetch error")
	writer := ObjectWriterMock{}
	nc := subject.NewNodeCapture("node1", "",
		NodeMock{Error: underlyingErr},
		RoleMock{},
		EnvMock{},
		CookbookMock{},
		&writer,
	)
	_, err := nc.CaptureNodeObject()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "fetch error") // because it's wrapped
}

func TestCaptureNodeObjectWithErrorOnSave(t *testing.T) {
	node := chef.Node{}
	writer := ObjectWriterMock{Error: errors.New("failed to write")}
	nc := subject.NewNodeCapture("", "",
		NodeMock{Node: node},
		RoleMock{},
		EnvMock{},
		CookbookMock{},
		&writer,
	)
	_, err := nc.CaptureNodeObject()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "failed to write")
	}
}

func TestCaptureRoleObjects(t *testing.T) {
	expectedRole := chef.Role{
		Name: "role1",
	}
	writer := ObjectWriterMock{}
	nc := subject.NewNodeCapture("", "",
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

func TestCaptureRoleObjectsWithErrorOnFetch(t *testing.T) {
	writer := ObjectWriterMock{}
	nc := subject.NewNodeCapture("",
		"",
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

func TestCaptureRoleObjectsWithErrorOnSave(t *testing.T) {
	writer := ObjectWriterMock{Error: errors.New("failed to write")}
	nc := subject.NewNodeCapture("",
		"",
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

func TestCaptureEnvObject(t *testing.T) {
	expectedEnv := chef.Environment{
		Name: "env1",
	}
	writer := ObjectWriterMock{}
	nc := subject.NewNodeCapture("", "",
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

func TestCaptureEnvObjectWithErrorOnFetch(t *testing.T) {
	writer := ObjectWriterMock{}
	nc := subject.NewNodeCapture("",
		"",
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

func TestCaptureEnvObjectWithErrorOnSave(t *testing.T) {
	writer := ObjectWriterMock{Error: errors.New("failed to write")}
	nc := subject.NewNodeCapture("",
		"",
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

func TestCaptureCookbooks(t *testing.T) {
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
	nc := subject.NewNodeCapture("temp", baseDir,
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

	err = nc.CaptureCookbooks(cbmap)
	assert.Nil(t, err)
}

func TestCaptureCookbooksWithDownloadError(t *testing.T) {
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
	nc := subject.NewNodeCapture("temp", baseDir,
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

	err = nc.CaptureCookbooks(cbmap)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "download error in test1")
		assert.Contains(t, err.Error(), "Failed to download cookbook")
	}
}
func TestCaptureCookbooksWithRenameError(t *testing.T) {
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
	nc := subject.NewNodeCapture("temp", baseDir,
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

	err = nc.CaptureCookbooks(cbmap)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "failed to rename cookbook")
	}
}
