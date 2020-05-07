//
// Copyright 2019 Chef Software, Inc.
// Author: Salim Afiune <afiune@chef.io>
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

	"github.com/spf13/cobra"

	"github.com/chef/chef-analyze/pkg/dist"
)

var (
	uploadCmd = &cobra.Command{
		Use:     "report upload LOCATION FILE",
		Aliases: []string{"upload"},
		Short:   fmt.Sprintf("Upload a file to %s", dist.CompanyName),
		Args:    cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			// cobra will make sure we always have exactly 2 arguments
			return UploadToS3(args[0], args[1])
		},
	}
)
