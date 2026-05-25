package cmd

import (
	"strconv"
	"testing"

	"github.com/chef/chef-analyze/pkg/reporting"
)

func benchmarkCookbooks(n int) []reporting.NodeCookbook {
	items := make([]reporting.NodeCookbook, n)
	for i := 0; i < n; i++ {
		s := strconv.Itoa(i)
		items[i] = reporting.NodeCookbook{
			Name:    "cookbook-" + s,
			Version: "1." + s,
		}
	}
	return items
}

func BenchmarkFormatCookbooks_1000(b *testing.B) {
	cookbooks := benchmarkCookbooks(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = formatCookbooks(cookbooks)
	}
}
