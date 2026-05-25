package formatter_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	subject "github.com/chef/chef-analyze/pkg/formatter"
	"github.com/chef/chef-analyze/pkg/reporting"
)

func TestContract_MakeNodesReportTXT_BoundaryGolden(t *testing.T) {
	t.Parallel()

	records := []*reporting.NodeReportItem{
		{
			Name:             "boundary-node",
			ChefVersion:      "",
			OS:               "",
			OSVersion:        "",
			CookbookVersions: nil,
			PolicyGroup:      "",
			Policy:           "",
			PolicyRev:        "",
		},
	}

	result := subject.MakeNodesReportTXT(records, "name:boundary-node")
	require.NotNil(t, result)

	goldenPath := filepath.Join("testdata", "nodes_report_txt_boundary.golden")
	expected, err := os.ReadFile(goldenPath)
	require.NoError(t, err)

	require.Equal(t, string(expected), result.Report)
	require.Empty(t, result.Errors)
}
