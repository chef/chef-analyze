package reporting_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/chef/chef-analyze/pkg/reporting"
	subject "github.com/chef/chef-analyze/pkg/reporting"
	"github.com/go-chef/chef"
	"github.com/stretchr/testify/assert"
)

func TestObjectWriterWriteJSONObject(t *testing.T) {
	baseDir, err := ioutil.TempDir(os.TempDir(), "chefanalyze-unit*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(baseDir)

	object := map[string]interface{}{
		"attribute": map[string]interface{}{
			"name":  "hello",
			"value": "hello",
		},
	}
	ow := subject.ObjectWriter{RootDir: baseDir}
	err = ow.WriteJSON("groupname", "objectname", object)
	if assert.Nil(t, err) {
		// In addition to verifying the content is placed at the correct location,
		// rading the file back in lets us verify Now make sure we get back the same content that we sent out.
		content, err := ioutil.ReadFile(fmt.Sprintf("%s/groupname/objectname.json", baseDir))
		assert.Nil(t, err)
		var objmap map[string]interface{}

		err = json.Unmarshal(content, &objmap)
		assert.Nil(t, err)
		assert.Equal(t, object, objmap)
	}
}

func TestObjectWriter_WriteContent(t *testing.T) {
	baseDir, err := ioutil.TempDir(os.TempDir(), "chefanalyze-unit*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(baseDir)
	expectedContent := []byte("hello world")
	expectedPath := fmt.Sprintf("%s/kitchen.yml", baseDir)

	ow := subject.ObjectWriter{RootDir: baseDir}
	err = ow.WriteContent("kitchen.yml", expectedContent)

	if assert.Nil(t, err) {
		readContent, err := ioutil.ReadFile(expectedPath)
		assert.Nil(t, err)
		assert.Equal(t, expectedContent, readContent)
	}
}

func TestObjectWriter_WriteContentInvalidPath(t *testing.T) {
	expectedContent := []byte("hello world")
	ow := subject.ObjectWriter{RootDir: "invalid"}
	err := ow.WriteContent("kitchen.yml", expectedContent)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "failed to create")
	}
}

func TestObjectWriter_WritePolicyRevision(t *testing.T) {
	type fields struct {
		RootDir string
	}
	type args struct {
		revision *chef.RevisionDetailsResponse
	}

	baseDir, err := ioutil.TempDir(os.TempDir(), "chefanalyze-unit*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(baseDir)
	testRevision := &chef.RevisionDetailsResponse{RevisionID: "1", Name: "foo"}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"TestWritePolicyRevision", fields{RootDir: baseDir}, args{revision: testRevision}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ow := &reporting.ObjectWriter{
				RootDir: tt.fields.RootDir,
			}
			if err := ow.WritePolicyRevision(tt.args.revision); (err != nil) != tt.wantErr { // Unexpected Error
				t.Errorf("ObjectWriter.PolicyRevision() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				content, err := ioutil.ReadFile(fmt.Sprintf("%s/policies/%s-%s.json", baseDir, testRevision.Name, testRevision.RevisionID))
				assert.Nil(t, err)
				var revWritten chef.RevisionDetailsResponse
				err = json.Unmarshal(content, &revWritten)
				assert.Nil(t, err)
				assert.Equal(t, testRevision, &revWritten)
			}
		})
	}
}

func TestObjectWriter_WritePolicyGroup(t *testing.T) {
	type fields struct {
		RootDir string
	}
	type args struct {
		pname  string
		pgroup *chef.PolicyGroup
	}

	baseDir, err := ioutil.TempDir(os.TempDir(), "chefanalyze-unit*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(baseDir)
	testGroup := &chef.PolicyGroup{Uri: "https://nowhere"}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"TestWritePolicyRevision", fields{RootDir: baseDir}, args{pname: "test1", pgroup: testGroup}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ow := &reporting.ObjectWriter{
				RootDir: tt.fields.RootDir,
			}
			if err := ow.WritePolicyGroup(tt.args.pname, tt.args.pgroup); (err != nil) != tt.wantErr { // Unexpected Error
				t.Errorf("ObjectWriter.WritePolicyGroup() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				content, err := ioutil.ReadFile(fmt.Sprintf("%s/policy_groups/%s.json", baseDir, tt.args.pname))
				assert.Nil(t, err)
				var revWritten chef.PolicyGroup
				err = json.Unmarshal(content, &revWritten)
				assert.Nil(t, err)
				assert.Equal(t, testGroup, &revWritten)
			}
		})
	}
}
func (ow *ObjectWriterMock) WriteDataBag(dataBag *chef.DataBag) error {
	if (ow.Error) == nil {
		ow.ReceivedObject = dataBag
	}
	return ow.Error
}
func (ow *ObjectWriterMock) WriteDataBagItem(bagName, itemName string, dataBagItem interface{}) error {
	if (ow.Error) == nil {
		ow.ReceivedObject = dataBagItem
	}
	return ow.Error
}

func TestObjectWriter_WriteRole(t *testing.T) {
	type fields struct {
		RootDir string
	}
	type args struct {
		role *chef.Role
	}

	baseDir, err := ioutil.TempDir(os.TempDir(), "chefanalyze-unit*")
	if err != nil {
		panic(err)
	}

	defer os.RemoveAll(baseDir)

	testRole := chef.Role{Name: "test_role", Description: "Test Description"}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"TestWriteRole", fields{RootDir: baseDir}, args{role: &testRole}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ow := &reporting.ObjectWriter{
				RootDir: tt.fields.RootDir,
			}
			if err := ow.WriteRole(tt.args.role); (err != nil) != tt.wantErr { // Unexpected Error
				t.Errorf("ObjectWriter.WriteRole() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				content, err := ioutil.ReadFile(fmt.Sprintf("%s/roles/%s.json", baseDir, testRole.Name))
				assert.Nil(t, err)
				var roleWritten chef.Role
				err = json.Unmarshal(content, &roleWritten)
				assert.Nil(t, err)
				assert.Equal(t, testRole, roleWritten)
			}
		})
	}
}

func TestObjectWriter_WriteDataBag(t *testing.T) {
	baseDir, err := ioutil.TempDir(os.TempDir(), "chefanalyze-unit*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(baseDir)

	testBag := chef.DataBag{Name: "test_bag"}
	ow := &reporting.ObjectWriter{RootDir: baseDir}
	err = ow.WriteDataBag(&testBag)
	if assert.Nil(t, err) {
		content, err := ioutil.ReadFile(fmt.Sprintf("%s/data_bags/%s.json", baseDir, testBag.Name))
		if assert.Nil(t, err) {
			var bagWritten chef.DataBag
			err = json.Unmarshal(content, &bagWritten)
			assert.Nil(t, err)
			assert.Equal(t, testBag, bagWritten)
		}
	}
}
func TestObjectWriter_WriteDataBagItem(t *testing.T) {
	baseDir, err := ioutil.TempDir(os.TempDir(), "chefanalyze-unit*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(baseDir)

	rawItem := map[string]interface{}{
		"somekey": map[string]interface{}{
			"value1": "value2",
			"value3": true,
		},
	}

	ow := &reporting.ObjectWriter{RootDir: baseDir}
	err = ow.WriteDataBagItem("testbag", "testitem", &rawItem)
	if assert.Nil(t, err) {
		content, err := ioutil.ReadFile(fmt.Sprintf("%s/data_bags/testbag/testitem.json", baseDir))
		if assert.Nil(t, err) {
			var bagItemWritten chef.DataBagItem
			err = json.Unmarshal(content, &bagItemWritten)
			assert.Nil(t, err)
			assert.Equal(t, rawItem, bagItemWritten)
		}
	}
}

func TestObjectWriter_WriteEnvironment(t *testing.T) {
	type fields struct {
		RootDir string
	}
	type args struct {
		env *chef.Environment
	}

	baseDir, err := ioutil.TempDir(os.TempDir(), "chefanalyze-unit*")
	if err != nil {
		panic(err)
	}

	defer os.RemoveAll(baseDir)

	testEvn := chef.Environment{Name: "test_env", Description: "Test Description", ChefType: "chef_type"}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"TestWriteEnv", fields{RootDir: baseDir}, args{env: &testEvn}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ow := &reporting.ObjectWriter{
				RootDir: tt.fields.RootDir,
			}
			if err := ow.WriteEnvironment(tt.args.env); (err != nil) != tt.wantErr {
				t.Errorf("ObjectWriter.WriteEnvironment() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				content, err := ioutil.ReadFile(fmt.Sprintf("%s/environments/%s.json", baseDir, testEvn.Name))
				assert.Nil(t, err)
				var evnWritten chef.Environment
				err = json.Unmarshal(content, &evnWritten)
				assert.Nil(t, err)
				assert.Equal(t, testEvn, evnWritten)
			}
		})
	}
}

func TestObjectWriter_WriteNode(t *testing.T) {
	type fields struct {
		RootDir string
	}
	type args struct {
		node *chef.Node
	}

	baseDir, err := ioutil.TempDir(os.TempDir(), "chefanalyze-unit*")
	if err != nil {
		panic(err)
	}

	defer os.RemoveAll(baseDir)

	testNode := chef.Node{Name: "test_env", Environment: "Production", ChefType: "chef_type", PolicyGroup: "test_group", PolicyName: "test_policy"}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"TestWriteNode", fields{RootDir: baseDir}, args{node: &testNode}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ow := &reporting.ObjectWriter{
				RootDir: tt.fields.RootDir,
			}
			if err := ow.WriteNode(tt.args.node); (err != nil) != tt.wantErr {
				t.Errorf("ObjectWriter.WriteNode() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				content, err := ioutil.ReadFile(fmt.Sprintf("%s/nodes/%s.json", baseDir, testNode.Name))
				assert.Nil(t, err)
				var nodeWritten chef.Node
				err = json.Unmarshal(content, &nodeWritten)
				assert.Nil(t, err)
				assert.Equal(t, testNode, nodeWritten)
			}
		})
	}
}

// This mock is defined here for shared use in other tests.
type ObjectWriterMock struct {
	Error          error
	ReceivedObject interface{}
}

func (ow *ObjectWriterMock) WriteContent(path string, content []byte) error {
	if ow.Error == nil {
		ow.ReceivedObject = content
	}
	return ow.Error
}

func (ow *ObjectWriterMock) WriteRole(role *chef.Role) error {
	if ow.Error == nil {
		ow.ReceivedObject = role
	}
	return ow.Error
}

func (ow *ObjectWriterMock) WriteEnvironment(env *chef.Environment) error {
	if ow.Error == nil {
		ow.ReceivedObject = env
	}
	return ow.Error
}

func (ow *ObjectWriterMock) WriteNode(node *chef.Node) error {
	if ow.Error == nil {
		ow.ReceivedObject = node
	}
	return ow.Error
}

func (ow *ObjectWriterMock) WritePolicy(name string, policy *chef.Policy) error {
	if ow.Error == nil {
		ow.ReceivedObject = policy
	}
	return ow.Error
}

func (ow *ObjectWriterMock) WritePolicyGroup(name string, pgroup *chef.PolicyGroup) error {
	if ow.Error == nil {
		ow.ReceivedObject = pgroup
	}
	return ow.Error
}

func (ow *ObjectWriterMock) WritePolicyRevision(policy *chef.RevisionDetailsResponse) error {
	if ow.Error == nil {
		ow.ReceivedObject = policy
	}
	return ow.Error
}
