package main

import (
	"os"

	"github.com/chef/chef-analyze/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
