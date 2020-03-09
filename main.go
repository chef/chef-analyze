//
// Copyright 2019 Chef Software, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package main

import (
	"fmt"
	"os"

	"github.com/chef/go-libs/featflag"

	"github.com/chef/chef-analyze/cmd"
	"github.com/spf13/cobra"
)

func main() {
	if featflag.ChefFeatAnalyze.Enabled() {
		if err := cmd.Execute(); err != nil {
			// Cobra returns this error if the root command (`chef-analyze`) is called without _any_
			// arguments, because our root command does not have a `Run` field. Cobra automatically
			// prints the help text in this case so we do not need to do anything custom. But it is
			// not a failure so we should not have a non-zero exit code.
			if err == cobra.ErrSubCommandRequired {
				os.Exit(0)
			}
			os.Exit(-1)
		}
	} else {
		fmt.Printf("`%s` is experimental and in development.\n\n", featflag.ChefFeatAnalyze.Key())
		fmt.Printf("Temporarily enable `%s` with the environment variable:\n", featflag.ChefFeatAnalyze.Key())
		fmt.Printf("\t%s=true\n\n", featflag.ChefFeatAnalyze.Env())
		fmt.Printf("Or, permanently by modifying $HOME/.chef-workstation/config.toml with:\n")
		fmt.Printf("\t[features]\n\t%s = true\n", featflag.ChefFeatAnalyze.Key())
	}
}
