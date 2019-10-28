package reporting_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/chef/chef-analyze/pkg/config"
	subject "github.com/chef/chef-analyze/pkg/reporting"
)

func TestCookbooks(t *testing.T) {
	err := subject.Cookbooks(&config.Config{})
	assert.NotNil(t, err)
}
