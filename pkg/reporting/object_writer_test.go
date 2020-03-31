package reporting_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	subject "github.com/chef/chef-analyze/pkg/reporting"
	chef "github.com/chef/go-chef"
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
	if (ow.Error) == nil {
		ow.ReceivedObject = role
	}
	return ow.Error
}

func (ow *ObjectWriterMock) WriteEnvironment(env *chef.Environment) error {
	if (ow.Error) == nil {
		ow.ReceivedObject = env
	}
	return ow.Error
}

func (ow *ObjectWriterMock) WriteNode(node *chef.Node) error {
	if (ow.Error) == nil {
		ow.ReceivedObject = node
	}
	return ow.Error
}
