// Copyright 2020 Chef Software, Inc.
// Author: Salim Afiune <afiune@chef.io>
// Author: Marc Paradise <marc@chef.io>
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

package cmd

import (
	"fmt"

	"github.com/chef/chef-analyze/pkg/dist"
	"github.com/spf13/cobra"
)

var infraFlags struct {
	credsFile     string
	clientName    string
	clientKey     string
	chefServerURL string
	profile       string
	noSSLverify   bool
}

// Adds common chef-infra flags to the given command
func addInfraFlagsToCommand(cmd *cobra.Command) {

	// global report commands infraFlags
	cmd.PersistentFlags().StringVarP(
		&infraFlags.credsFile,
		"credentials", "c", "",
		fmt.Sprintf("credentials file (default $HOME/%s/credentials)",
			dist.UserConfDir),
	)
	cmd.PersistentFlags().StringVarP(
		&infraFlags.clientName, "client-name", "n", "",
		fmt.Sprintf("%s API client name", dist.ServerProduct),
	)
	cmd.PersistentFlags().StringVarP(
		&infraFlags.clientKey, "client-key", "k", "",
		fmt.Sprintf("%s API client key", dist.ServerProduct),
	)
	cmd.PersistentFlags().StringVarP(
		&infraFlags.chefServerURL, "chef-server-url", "s", "",
		fmt.Sprintf("%s URL", dist.ServerProduct),
	)
	cmd.PersistentFlags().StringVarP(
		&infraFlags.profile,
		"profile", "p", "default",
		"profile to use from credentials file",
	)
	cmd.PersistentFlags().BoolVarP(
		&infraFlags.noSSLverify,
		"ssl-no-verify", "o", false,
		fmt.Sprintf("Do not verify SSL when connecting to %s (default: verify)",
			dist.ServerProduct),
	)
}
