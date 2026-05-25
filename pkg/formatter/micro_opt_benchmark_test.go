package formatter_test

import (
	"fmt"
	"testing"

	subject "github.com/chef/chef-analyze/pkg/formatter"
	"github.com/chef/chef-analyze/pkg/reporting"
)

func benchmarkNodeRecords(n int, policyManaged bool) []*reporting.NodeReportItem {
	records := make([]*reporting.NodeReportItem, 0, n)
	for i := 0; i < n; i++ {
		r := &reporting.NodeReportItem{
			Name:        fmt.Sprintf("node-%04d", i),
			ChefVersion: "18.4.0",
			OS:          "ubuntu",
			OSVersion:   "22.04",
			CookbookVersions: []reporting.CookbookVersion{
				{Name: "apache2", Version: "9.9.1"},
				{Name: "base", Version: "2.4.0"},
				{Name: "users", Version: "1.0.5"},
			},
		}
		if policyManaged {
			r.PolicyGroup = "prod"
			r.Policy = "linux-base"
			r.PolicyRev = "1234567890abcdef"
		}
		records = append(records, r)
	}
	return records
}

func BenchmarkMakeNodesReportTXT_PolicyManaged(b *testing.B) {
	records := benchmarkNodeRecords(300, true)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := subject.MakeNodesReportTXT(records, "name:node*")
		if result == nil || len(result.Report) == 0 {
			b.Fatal("unexpected empty result")
		}
	}
}

func BenchmarkNodesReportSummary(b *testing.B) {
	records := benchmarkNodeRecords(300, false)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := subject.NodesReportSummary(records, "")
		if len(result.Report) == 0 {
			b.Fatal("unexpected empty result")
		}
	}
}
