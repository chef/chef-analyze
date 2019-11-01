package reporting_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	subject "github.com/chef/chef-analyze/pkg/reporting"
)

func TestDemoFunction(t *testing.T) {
	err := subject.DemoFunction()
	assert.Nil(t, err)
}
