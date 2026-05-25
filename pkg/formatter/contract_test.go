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

func TestContract_MakeNodesReportTXT_EmptyWithFilterBoundaryGolden(t *testing.T) {
	t.Parallel()

	result := subject.MakeNodesReportTXT([]*reporting.NodeReportItem{}, "name:boundary-empty")
	require.NotNil(t, result)

	goldenPath := filepath.Join("testdata", "nodes_report_txt_boundary_empty_filter.golden")
	expected, err := os.ReadFile(goldenPath)
	require.NoError(t, err)

	require.Equal(t, string(expected), result.Report)
	require.Empty(t, result.Errors)
}

func TestContract_MakeNodesReportTXT_PolicyCookbooksStripVersionBoundaryGolden(t *testing.T) {
	t.Parallel()

	records := []*reporting.NodeReportItem{
		{
			Name:        "policy-boundary-node",
			ChefVersion: "",
			OS:          "",
			OSVersion:   "",
			CookbookVersions: []reporting.CookbookVersion{
				{Name: "mycookbook", Version: "1.0"},
				{Name: "test", Version: "9.9"},
			},
			PolicyGroup: "staging",
			Policy:      "seven-zip",
			PolicyRev:   "99999xxxx99999",
		},
	}

	result := subject.MakeNodesReportTXT(records, "name:policy-boundary-node")
	require.NotNil(t, result)

	goldenPath := filepath.Join("testdata", "nodes_report_txt_boundary_policy_strip_versions.golden")
	expected, err := os.ReadFile(goldenPath)
	require.NoError(t, err)

	require.Equal(t, string(expected), result.Report)
	require.Empty(t, result.Errors)
}
