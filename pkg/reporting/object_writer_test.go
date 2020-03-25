package reporting_test

import chef "github.com/chef/go-chef"

type ObjectWriterMock struct {
	Error                 error
	SavedRoleCount        int
	SavedEnvironmentCount int
	SavedNodeCount        int
	ReceivedObject        interface{}
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
