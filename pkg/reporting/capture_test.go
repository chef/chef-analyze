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
	"github.com/go-chef/chef"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type CapturerMock struct {
	NodeReturn                  *chef.Node
	CookbookReturn              []subject.NodeCookbook
	CBAErrorReturn              error
	CookbookArtifactErrorReturn error
	PolicyGroupReturn           *chef.PolicyGroup
	PolicyGroupError            error
	PolicyObjectReturn          *chef.RevisionDetailsResponse
	PolicyObjectErrorReturn     error
	NodeErrorReturn             error
	EnvErrorReturn              error
	RoleErrorReturn             error
	CookbookErrorReturn         error
	KitchenErrorReturn          error
	DataBagErrorReturn          error
}

func (cm *CapturerMock) CaptureCookbooks(string, map[string]interface{}) ([]subject.NodeCookbook, error) {
	return cm.CookbookReturn, cm.CookbookErrorReturn
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

func (cm *CapturerMock) SaveKitchenYML(node *chef.Node) error {
	return cm.KitchenErrorReturn
}

func (cm *CapturerMock) CaptureCookbookArtifacts(string, *chef.RevisionDetailsResponse) error {
	return cm.CBAErrorReturn
}

func (cm *CapturerMock) CapturePolicyObject(string, string) (*chef.RevisionDetailsResponse, error) {
	return cm.PolicyObjectReturn, cm.PolicyObjectErrorReturn
}

func (cm *CapturerMock) CapturePolicyGroupObject(string) (*chef.PolicyGroup, error) {
	return cm.PolicyGroupReturn, cm.PolicyGroupError
}

func (cm *CapturerMock) CaptureAllDataBagItems() error {
	return cm.DataBagErrorReturn
}

func nodeWithChefInstall() *chef.Node {
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
			"chef_packages": map[string]interface{}{
				"chef": map[string]interface{}{
					"version": "99.0",
				},
			},
			"os":               "linux",
			"platform":         "ubuntu",
			"platform_version": "18.04",
			"platform_family":  "debian",
		},
	}
}

func policyManagedNode() *chef.Node {
	return &chef.Node{
		Name:        "node1",
		Environment: "_default",
		RunList:     []string{"cookbook1::recipe1", "role:mockrole"},
		PolicyName:  "pgroup",
		PolicyGroup: "policy",
		AutomaticAttributes: map[string]interface{}{
			"cookbooks": map[string]interface{}{
				"foo": map[string]interface{}{
					"version": "0.1.0",
				},
			},
		},
	}
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
	nc := subject.NewNodeCapture("node1", baseDir, subject.CaptureOpts{}, &CapturerMock{NodeReturn: defaultNode()})
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
	assert.Equal(t, subject.WritingKitchenConfig, val)
	val = <-nc.Progress
	assert.Equal(t, subject.CaptureComplete, val)

	assert.Nil(t, nc.Error)

}

func TestCapture_RunWithDataBags(t *testing.T) {
	baseDir, err := ioutil.TempDir(os.TempDir(), "chefanalyze-unit*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(baseDir)
	nc := subject.NewNodeCapture("node1", baseDir,
		subject.CaptureOpts{DownloadDataBags: true}, &CapturerMock{NodeReturn: defaultNode()})
	go nc.Run()
	val := <-nc.Progress
	assert.Equal(t, subject.FetchingNode, val)
	val = <-nc.Progress
	assert.Equal(t, subject.FetchingDataBags, val)
	val = <-nc.Progress
	assert.Equal(t, subject.FetchingCookbooks, val)
	val = <-nc.Progress
	assert.Equal(t, subject.FetchingEnvironment, val)
	val = <-nc.Progress
	assert.Equal(t, subject.FetchingRoles, val)
	val = <-nc.Progress
	assert.Equal(t, subject.WritingKitchenConfig, val)
	val = <-nc.Progress
	assert.Equal(t, subject.CaptureComplete, val)

	assert.Nil(t, nc.Error)

}

func TestCapture_RunWithPolicyManagedNode(t *testing.T) {
	policyGroup := chef.PolicyGroup{
		Policies: map[string]chef.Revision{
			"my-policy": chef.Revision{
				"revision_id": "123xyz",
			},
		},
	}

	policyDetail := chef.RevisionDetailsResponse{
		Name:       "my-policy",
		RevisionID: "123xyz",
		CookbookLocks: map[string]chef.CookbookLock{
			"alpha": chef.CookbookLock{
				Identifier: "123456789012345678901234567890xyz",
			},
			"gama": chef.CookbookLock{
				Identifier: "abc456789012345678901234567890xyz",
			},
		},
	}
	baseDir, err := ioutil.TempDir(os.TempDir(), "chefanalyze-unit*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(baseDir)
	// PolicyGroupReturn           *chef.PolicyGroup
	// PolicyObjectReturn          *chef.RevisionDetailsResponse
	nc := subject.NewNodeCapture("node1", baseDir,
		subject.CaptureOpts{},
		&CapturerMock{
			NodeReturn:         policyManagedNode(),
			PolicyGroupReturn:  &policyGroup,
			PolicyObjectReturn: &policyDetail,
		})
	go nc.Run()
	val := <-nc.Progress
	assert.Equal(t, subject.FetchingNode, val)
	val = <-nc.Progress
	assert.Equal(t, subject.FetchingPolicyData, val)
	val = <-nc.Progress
	assert.Equal(t, subject.FetchingCookbookArtifacts, val)
	val = <-nc.Progress
	assert.Equal(t, subject.WritingKitchenConfig, val)
	val = <-nc.Progress
	assert.Equal(t, subject.CaptureComplete, val)

	assert.Nil(t, nc.Error)
}

func TestCapture_RunWithPolicyManagedNode_FailPolicyGroupCapture(t *testing.T) {
	baseDir, err := ioutil.TempDir(os.TempDir(), "chefanalyze-unit*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(baseDir)
	nc := subject.NewNodeCapture("node1", baseDir,
		subject.CaptureOpts{},
		&CapturerMock{
			NodeReturn:       policyManagedNode(),
			PolicyGroupError: errors.New("403 on group fetch"),
		})
	nc.Run()

	if assert.NotNil(t, nc.Error) {
		assert.Contains(t, nc.Error.Error(), "unable to capture policy group 'policy': 403 on group fetch")
	}

}
func TestCapture_RunWithPolicyManagedNode_FailPolicyObjectCapture(t *testing.T) {
	baseDir, err := ioutil.TempDir(os.TempDir(), "chefanalyze-unit*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(baseDir)
	nc := subject.NewNodeCapture("node1", baseDir,
		subject.CaptureOpts{},
		&CapturerMock{
			NodeReturn:              policyManagedNode(),
			PolicyGroupReturn:       &chef.PolicyGroup{},
			PolicyObjectErrorReturn: errors.New("sorry fresh out of policy objects"),
		})
	nc.Run()

	if assert.NotNil(t, nc.Error) {
		assert.Contains(t, nc.Error.Error(), "unable to capture policy name 'pgroup' revision ''")
		assert.Contains(t, nc.Error.Error(), "sorry fresh out of policy objects")
	}
}

func TestCapture_RunWithPolicyManagedNode_FailCookbookArtifacts(t *testing.T) {
	baseDir, err := ioutil.TempDir(os.TempDir(), "chefanalyze-unit*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(baseDir)
	nc := subject.NewNodeCapture("node1", baseDir,
		subject.CaptureOpts{},
		&CapturerMock{
			PolicyGroupReturn: &chef.PolicyGroup{},
			NodeReturn:        policyManagedNode(), CBAErrorReturn: errors.New("sorry all out of artifacts"),
		})
	nc.Run()

	if assert.NotNil(t, nc.Error) {
		assert.Contains(t, nc.Error.Error(), "unable to capture cookbook artifacts for policy 'pgroup' revision ''")
		assert.Contains(t, nc.Error.Error(), "sorry all out of artifacts")
	}
}

func TestCapture_RunWithNodeFailure(t *testing.T) {
	nc := subject.NewNodeCapture("node1", "", subject.CaptureOpts{},
		&CapturerMock{NodeErrorReturn: errors.New("no node")})
	nc.Run()
	if assert.NotNil(t, nc.Error) {
		assert.Contains(t, nc.Error.Error(), "no node")
		assert.Contains(t, nc.Error.Error(), "unable to capture node 'node1'")
	}
}

func TestCapture_RunWithNoCookbooksAvailable(t *testing.T) {
	nc := subject.NewNodeCapture("node1", "", subject.CaptureOpts{},
		&CapturerMock{NodeReturn: &chef.Node{}})
	nc.Run()
	assert.Nil(t, nc.Error)
}

func TestCapture_RunWithCookbookFailure(t *testing.T) {
	nc := subject.NewNodeCapture("node1", "", subject.CaptureOpts{},
		&CapturerMock{NodeReturn: defaultNode(),
			CookbookErrorReturn: errors.New("no cookbook")})
	nc.Run()
	if assert.NotNil(t, nc.Error) {
		assert.Contains(t, nc.Error.Error(), "no cookbook")
		assert.Contains(t, nc.Error.Error(), "unable to capture node cookbooks")
	}
}

func TestCapture_RunWithEnvironmentFailure(t *testing.T) {
	nc := subject.NewNodeCapture("node1", "", subject.CaptureOpts{},
		&CapturerMock{NodeReturn: defaultNode(),
			EnvErrorReturn: errors.New("no env")})
	nc.Run()
	if assert.NotNil(t, nc.Error) {
		assert.Contains(t, nc.Error.Error(), "no env")
		assert.Contains(t, nc.Error.Error(), "unable to capture environment")
	}
}

func TestCapture_RunWithRoleFailure(t *testing.T) {

	nc := subject.NewNodeCapture("node1", "", subject.CaptureOpts{},
		&CapturerMock{NodeReturn: defaultNode(),
			RoleErrorReturn: errors.New("no role")})
	nc.Run()
	if assert.NotNil(t, nc.Error) {
		assert.Contains(t, nc.Error.Error(), "no role")
		assert.Contains(t, nc.Error.Error(), "unable to capture role")
	}
}

func TestCapture_RunWithCookbookDataBagFailure(t *testing.T) {
	nc := subject.NewNodeCapture("node1", "", subject.CaptureOpts{DownloadDataBags: true},
		&CapturerMock{NodeReturn: defaultNode(),
			DataBagErrorReturn: errors.New("spilled all over the carpet")})
	nc.Run()
	if assert.NotNil(t, nc.Error) {
		assert.Contains(t, nc.Error.Error(), "spilled all over the carpet")
		assert.Contains(t, nc.Error.Error(), "unable to capture data bag")
	}
}
func TestCapture_RunWithKitchenYMLFailure(t *testing.T) {
	nc := subject.NewNodeCapture("node1", "", subject.CaptureOpts{},
		&CapturerMock{NodeReturn: defaultNode(),
			KitchenErrorReturn: errors.New("failure here")})
	nc.Run()
	if assert.NotNil(t, nc.Error) {
		assert.Contains(t, nc.Error.Error(), "failure here")
		assert.Contains(t, nc.Error.Error(), "unable to write Kitchen config")
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
		DataBagMock{},
		PolicyGroupMock{},
		PolicyMock{},
		CBAMock{},
		&writer,
	)
	actualNode, err := nc.CaptureNodeObject("node1")
	assert.Equal(t, &expectedNode, actualNode)
	assert.Equal(t, nil, err)
}

func TestCapturer_CaptureNodeObjectWithErrorOnFetch(t *testing.T) {
	underlyingErr := errors.New("fetch error")
	writer := ObjectWriterMock{}
	nc := subject.NewNodeCapturer(
		NodeMock{Error: underlyingErr},
		RoleMock{},
		EnvMock{},
		CookbookMock{},
		DataBagMock{},
		PolicyGroupMock{},
		PolicyMock{},
		CBAMock{},
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
		DataBagMock{},
		PolicyGroupMock{},
		PolicyMock{},
		CBAMock{},
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
		DataBagMock{},
		PolicyGroupMock{},
		PolicyMock{},
		CBAMock{},
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
		DataBagMock{},
		PolicyGroupMock{},
		PolicyMock{},
		CBAMock{},
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
		DataBagMock{},
		PolicyGroupMock{},
		PolicyMock{},
		CBAMock{},
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
		DataBagMock{},
		PolicyGroupMock{},
		PolicyMock{},
		CBAMock{},
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
		DataBagMock{},
		PolicyGroupMock{},
		PolicyMock{},
		CBAMock{},
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
		DataBagMock{},
		PolicyGroupMock{},
		PolicyMock{},
		CBAMock{},
		&writer,
	)
	err := nc.CaptureEnvObject("env1")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to write")

}

func TestCapturer_CaptureCookbookArtifacts(t *testing.T) {
	baseDir, err := ioutil.TempDir(os.TempDir(), "chefanalyze-unit*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(baseDir)

	CBAList := chef.CBAGetResponse{
		"alpha": chef.CBA{
			CBAVersions: []chef.CBAVersion{
				chef.CBAVersion{Identifier: "123456789012345678901234567890xyz"},
			},
		},
		"gama": chef.CBA{
			CBAVersions: []chef.CBAVersion{
				chef.CBAVersion{Identifier: "abc456789012345678901234567890xyz"},
			},
		},
	}
	//
	// policyGroupList := chef.PolicyGroupGetResponse{
	//   "my-policygroup": chef.PolicyGroup{
	//     Policies: map[string]chef.Revision{
	//       "my-policy": chef.Revision{
	//         "revision_id": "123xyz",
	//       },
	//     },
	//   },
	// }

	policyDetail := chef.RevisionDetailsResponse{
		Name:       "my-policy",
		RevisionID: "123xyz",
		CookbookLocks: map[string]chef.CookbookLock{
			"alpha": chef.CookbookLock{
				Identifier: "123456789012345678901234567890xyz",
			},
			"gama": chef.CookbookLock{
				Identifier: "abc456789012345678901234567890xyz",
			},
		},
	}

	writer := ObjectWriterMock{}
	nc := subject.NewNodeCapturer(
		NodeMock{},
		RoleMock{},
		EnvMock{},
		CookbookMock{},
		DataBagMock{},
		PolicyGroupMock{},
		PolicyMock{},
		newMockCookbookArtifact(CBAList, nil, nil),
		&writer,
	)

	err = nc.CaptureCookbookArtifacts(baseDir, &policyDetail)
	assert.Nil(t, err)
}

func TestCapturer_CaptureCookbookArtifacts_WithDownloadError(t *testing.T) {
	baseDir, err := ioutil.TempDir(os.TempDir(), "chefanalyze-unit*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(baseDir)

	CBAList := chef.CBAGetResponse{
		"alpha": chef.CBA{
			CBAVersions: []chef.CBAVersion{
				chef.CBAVersion{Identifier: "123456789012345678901234567890xyz"},
			},
		},
		"gama": chef.CBA{
			CBAVersions: []chef.CBAVersion{
				chef.CBAVersion{Identifier: "abc456789012345678901234567890xyz"},
			},
		},
	}
	policyDetail := chef.RevisionDetailsResponse{
		Name:       "my-policy",
		RevisionID: "123xyz",
		CookbookLocks: map[string]chef.CookbookLock{
			"alpha": chef.CookbookLock{
				Identifier: "123456789012345678901234567890xyz",
			},
		},
	}

	writer := ObjectWriterMock{}
	nc := subject.NewNodeCapturer(
		NodeMock{},
		RoleMock{},
		EnvMock{},
		CookbookMock{},
		DataBagMock{},
		PolicyGroupMock{},
		PolicyMock{},
		newMockCookbookArtifact(CBAList, nil, errors.New("Download Error")),
		&writer,
	)

	err = nc.CaptureCookbookArtifacts(baseDir, &policyDetail)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "alpha")
		assert.Contains(t, err.Error(), "Download Error")
	}
}

func TestCapturer_CapturePolicyGroupObject(t *testing.T) {
	policyGroup := chef.PolicyGroup{
		Policies: map[string]chef.Revision{
			"my-policy": chef.Revision{
				"revision_id": "123xyz",
			},
		},
	}
	policyGroupList := chef.PolicyGroupGetResponse{
		"my-policygroup": policyGroup,
	}

	writer := ObjectWriterMock{}
	nc := subject.NewNodeCapturer(
		NodeMock{},
		RoleMock{},
		EnvMock{},
		CookbookMock{},
		DataBagMock{},
		newMockPolicyGroup(policyGroupList, nil),
		PolicyMock{},
		CBAMock{},
		&writer,
	)

	response, err := nc.CapturePolicyGroupObject("my-policygroup")
	if assert.Nil(t, err) {
		assert.Equal(t, policyGroup, *response)
		// Make sure that the writer was asked to save.
		assert.Equal(t, writer.ReceivedObject, &policyGroup)
	}
}

func TestCapturer_CapturePolicyGroupObjectWithErrNotFound(t *testing.T) {
	policyGroup := chef.PolicyGroup{
		Policies: map[string]chef.Revision{
			"my-policy": chef.Revision{
				"revision_id": "123xyz",
			},
		},
	}
	policyGroupList := chef.PolicyGroupGetResponse{
		"my-policygroup": policyGroup,
	}

	writer := ObjectWriterMock{}
	nc := subject.NewNodeCapturer(
		NodeMock{},
		RoleMock{},
		EnvMock{},
		CookbookMock{},
		DataBagMock{},
		newMockPolicyGroup(policyGroupList, nil),
		PolicyMock{},
		CBAMock{},
		&writer,
	)

	response, err := nc.CapturePolicyGroupObject("herbie")
	assert.Nil(t, response)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Could not find policy group: herbie")
	}
}

func TestCapturer_CapturePolicyGroupObject_WithErrorOnFetch(t *testing.T) {
	policyGroupList := chef.PolicyGroupGetResponse{}

	nc := subject.NewNodeCapturer(
		NodeMock{},
		RoleMock{},
		EnvMock{},
		CookbookMock{},
		DataBagMock{},
		newMockPolicyGroup(policyGroupList, errors.New("failed to fetch")),
		PolicyMock{},
		CBAMock{},
		&ObjectWriterMock{},
	)

	_, err := nc.CapturePolicyGroupObject("my-policygroup")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "unable to retrieve policy groups: failed to fetch")
	}
}

func TestCapturer_CapturePolicyGroupObject_WithErrorOnWrite(t *testing.T) {

	policyGroup := chef.PolicyGroup{
		Policies: map[string]chef.Revision{
			"my-policy": chef.Revision{
				"revision_id": "123xyz",
			},
		},
	}
	policyGroupList := chef.PolicyGroupGetResponse{
		"my-policygroup": policyGroup,
	}
	nc := subject.NewNodeCapturer(
		NodeMock{},
		RoleMock{},
		EnvMock{},
		CookbookMock{},
		DataBagMock{},
		newMockPolicyGroup(policyGroupList, nil),
		PolicyMock{},
		CBAMock{},
		&ObjectWriterMock{Error: errors.New("could not write")},
	)

	_, err := nc.CapturePolicyGroupObject("my-policygroup")

	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "failed to save policy group my-policygroup to disk")
		assert.Contains(t, err.Error(), "could not write")
	}
}

func TestCapturer_CapturePolicyObject(t *testing.T) {
	policyDetail := chef.RevisionDetailsResponse{
		Name:       "my-policy",
		RevisionID: "123xyz",
		CookbookLocks: map[string]chef.CookbookLock{
			"alpha": chef.CookbookLock{
				Identifier: "123456789012345678901234567890xyz",
			},
			"gama": chef.CookbookLock{
				Identifier: "abc456789012345678901234567890xyz",
			},
		},
	}

	nc := subject.NewNodeCapturer(
		NodeMock{},
		RoleMock{},
		EnvMock{},
		CookbookMock{},
		DataBagMock{},
		PolicyGroupMock{},
		newMockPolicy(policyDetail, nil),
		CBAMock{},
		&ObjectWriterMock{},
	)

	response, err := nc.CapturePolicyObject("my-policy", "123xyz")
	assert.Nil(t, err)
	if assert.NotNil(t, response) {
		assert.Equal(t, response, &policyDetail)
	}

}

func TestCapturer_CapturePolicyObject_WithErrorOnFetch(t *testing.T) {
	policyDetail := chef.RevisionDetailsResponse{}
	nc := subject.NewNodeCapturer(
		NodeMock{},
		RoleMock{},
		EnvMock{},
		CookbookMock{},
		DataBagMock{},
		PolicyGroupMock{},
		newMockPolicy(policyDetail, errors.New("Failed to fetch")),
		CBAMock{},
		&ObjectWriterMock{},
	)
	_, err := nc.CapturePolicyObject("my-policy", "123xyz")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "unable to retrieve policy object: my-policy: Failed to fetch")
	}
}

func TestCapturer_CapturePolicyObject_WithErrorOnWrite(t *testing.T) {
	policyDetail := chef.RevisionDetailsResponse{}
	nc := subject.NewNodeCapturer(
		NodeMock{},
		RoleMock{},
		EnvMock{},
		CookbookMock{},
		DataBagMock{},
		PolicyGroupMock{},
		newMockPolicy(policyDetail, nil),
		CBAMock{},
		&ObjectWriterMock{Error: errors.New("no write for you")},
	)

	_, err := nc.CapturePolicyObject("my-policy", "123xyz")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "failed to save policy object to disk: my-policy: no write for you")
	}
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

	nc := subject.NewNodeCapturer(
		NodeMock{},
		RoleMock{},
		EnvMock{},
		newMockCookbook(cookbookList, nil, nil),
		DataBagMock{},
		PolicyGroupMock{},
		PolicyMock{},
		CBAMock{},
		&ObjectWriterMock{},
	)

	cbmap := map[string]interface{}{
		"foo": map[string]interface{}{
			"version": "0.1.0",
		},
	}

	cbs, err := nc.CaptureCookbooks(baseDir, cbmap)
	if assert.Nil(t, err) {
		assert.Contains(t, cbs, subject.NodeCookbook{"foo", "0.1.0"})
	}
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
		DataBagMock{},
		PolicyGroupMock{},
		PolicyMock{},
		CBAMock{},
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

	_, err = nc.CaptureCookbooks(baseDir, cbmap)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "download error in test1")
		assert.Contains(t, err.Error(), "Failed to download cookbook")
	}
}
func TestCapturer_CaptureCookbooksWithRenameError(t *testing.T) {
	cookbookList := chef.CookbookListResult{
		"foo": chef.CookbookVersions{
			Versions: []chef.CookbookVersion{
				chef.CookbookVersion{Version: "0.1.0"},
			},
		},
	}

	writerMock := ObjectWriterMock{}
	cbMock := newMockCookbook(cookbookList, nil, nil)
	cbMock.createDirOnDownload = false
	nc := subject.NewNodeCapturer(
		NodeMock{},
		RoleMock{},
		EnvMock{},
		cbMock,
		DataBagMock{},
		PolicyGroupMock{},
		PolicyMock{},
		CBAMock{},
		&writerMock,
	)

	// To force this error, we will not manually create the versioned cookbook
	// directory name that a real download cookbooks would do. This will cause the
	// rename to fail since the directory to be renamed won't exist.
	cbmap := map[string]interface{}{
		"foo": map[string]interface{}{
			"version": "0.1.0",
		},
	}

	_, err := nc.CaptureCookbooks("/mocked/anyway", cbmap)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "failed to rename cookbook")
	}
}

func TestCapturer_CaptureAllDataBags(t *testing.T) {
	dataBags := map[string]string{
		"bag1": "https://url/data/bag1",
	}
	dataBagItemList := map[string]string{
		"item1": "https://url/data/bag1/item1",
	}
	dataBagItem := map[string]interface{}{
		"key1": map[string]interface{}{
			"value1": "value2",
			"value3": true,
		},
	}

	writer := ObjectWriterMock{}
	nc := subject.NewNodeCapturer(
		NodeMock{},
		RoleMock{},
		EnvMock{},
		CookbookMock{},
		DataBagMock{
			desiredDataBagList:     dataBags,
			desiredDataBagItem:     dataBagItem,
			desiredDataBagItemList: dataBagItemList,
		},
		PolicyGroupMock{},
		PolicyMock{},
		CBAMock{},
		&writer,
	)
	err := nc.CaptureAllDataBagItems()
	assert.Equal(t, dataBagItem, writer.ReceivedObject.(map[string]interface{}))
	assert.Nil(t, err)
}

func TestCapturer_CaptureAllDataBagsWithDBListError(t *testing.T) {
	nc := subject.NewNodeCapturer(
		NodeMock{},
		RoleMock{},
		EnvMock{},
		CookbookMock{},
		DataBagMock{desiredDataBagListError: errors.New("i misplaced them")},
		PolicyGroupMock{},
		PolicyMock{},
		CBAMock{},
		&ObjectWriterMock{},
	)
	err := nc.CaptureAllDataBagItems()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "unable to retrieve list of data bags")
		assert.Contains(t, err.Error(), "i misplaced them")
	}
}

func TestCapturer_CaptureAllDataBagsWithListItemsError(t *testing.T) {
	dataBags := map[string]string{
		"bag1": "https://url/data/bag1",
	}
	nc := subject.NewNodeCapturer(
		NodeMock{},
		RoleMock{},
		EnvMock{},
		CookbookMock{},
		DataBagMock{
			desiredDataBagList:          dataBags,
			desiredDataBagItemListError: errors.New("empty bags"),
		},
		PolicyGroupMock{},
		PolicyMock{},
		CBAMock{},
		&ObjectWriterMock{},
	)
	err := nc.CaptureAllDataBagItems()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "unable to retrieve data bag items for bag1")
		assert.Contains(t, err.Error(), "empty bags")
	}
}

func TestCapturer_CaptureAllDataBagsWithGetItemError(t *testing.T) {
	dataBags := map[string]string{
		"bag1": "https://url/data/bag1",
	}
	dataBagItemList := map[string]string{
		"item1": "https://url/data/bag1/item1",
	}
	nc := subject.NewNodeCapturer(
		NodeMock{},
		RoleMock{},
		EnvMock{},
		CookbookMock{},
		DataBagMock{
			desiredDataBagList:      dataBags,
			desiredDataBagItemList:  dataBagItemList,
			desiredDataBagItemError: errors.New("i swear they were right here"),
		},
		PolicyGroupMock{},
		PolicyMock{},
		CBAMock{},
		&ObjectWriterMock{},
	)
	err := nc.CaptureAllDataBagItems()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "unable to retrieve data bag item bag1/item1")
		assert.Contains(t, err.Error(), "i swear they were right here")
	}
}

func TestCapturer_CaptureAllDataBagsWithWriteError(t *testing.T) {
	dataBags := map[string]string{
		"bag1": "https://url/data/bag1",
	}
	dataBagItemList := map[string]string{
		"item1": "https://url/data/bag1/item1",
	}
	dataBagItem := map[string]interface{}{
		"key1": map[string]interface{}{
			"value1": "value2",
			"value3": true,
		},
	}

	nc := subject.NewNodeCapturer(
		NodeMock{},
		RoleMock{},
		EnvMock{},
		CookbookMock{},
		DataBagMock{
			desiredDataBagList:     dataBags,
			desiredDataBagItem:     dataBagItem,
			desiredDataBagItemList: dataBagItemList,
		},
		PolicyGroupMock{},
		PolicyMock{},
		CBAMock{},
		&ObjectWriterMock{Error: errors.New("dropped the bag")},
	)
	err := nc.CaptureAllDataBagItems()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "unable to save data bag item bag1/item1")
		assert.Contains(t, err.Error(), "dropped the bag")
	}
}

func TestCapturer_SaveKitchenYML(t *testing.T) {
	writerMock := ObjectWriterMock{}
	node := nodeWithChefInstall()
	nc := subject.NewNodeCapturer(
		NodeMock{},
		RoleMock{},
		EnvMock{},
		CookbookMock{},
		DataBagMock{},
		PolicyGroupMock{},
		PolicyMock{},
		CBAMock{},
		&writerMock,
	)
	err := nc.SaveKitchenYML(node)

	expectedYML := []byte(`
---
driver:
  name: vagrant

provisioner:
  name: chef_zero_capture
  product_name: chef
  product_version: 99.0
  json_attributes: false
  client_rb:
    node_name: node1

platforms:
  - name: ubuntu-18.04

suites:
  - name: node1
`)

	if assert.Nil(t, err) {
		// cast to string to keep it readable if it fails
		assert.Equal(t, string(expectedYML), string(writerMock.ReceivedObject.([]byte)))
	}
}

func TestCapturer_SaveKitchenYMLWithMissingAttrs(t *testing.T) {
	writerMock := ObjectWriterMock{}
	nc := subject.NewNodeCapturer(
		NodeMock{},
		RoleMock{},
		EnvMock{},
		CookbookMock{},
		DataBagMock{},
		PolicyGroupMock{},
		PolicyMock{},
		CBAMock{},
		&writerMock,
	)

	// Missing chef_packages['chef_packages']
	node := defaultNode()
	err := nc.SaveKitchenYML(node)
	if assert.NotNil(t, err) {
		// cast to string to keep it readable if it fails
		assert.Contains(t, err.Error(), "missing automatic attribute 'chef_packages'")
	}

	// Missing chef_packages['chef']
	node = nodeWithChefInstall()
	i := node.AutomaticAttributes["chef_packages"].(map[string]interface{})
	delete(i, "chef")
	err = nc.SaveKitchenYML(node)
	if assert.NotNil(t, err) {
		// cast to string to keep it readable if it fails
		assert.Contains(t, err.Error(), "missing automatic attribute chef_packages['chef']")
	}

	// Missing chef_packages['chef'].version
	node = nodeWithChefInstall()
	i = node.AutomaticAttributes["chef_packages"].(map[string]interface{})["chef"].(map[string]interface{})
	delete(i, "version")
	err = nc.SaveKitchenYML(node)
	if assert.NotNil(t, err) {
		// cast to string to keep it readable if it fails
		assert.Contains(t, err.Error(), "missing automatic attribute chef_packages['chef']['version']")
	}
}

func TestCapturer_SaveKitchenYMLWithFailedWrite(t *testing.T) {

}
