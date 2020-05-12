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
	"strconv"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/chef/chef-analyze/pkg/dist"
)

var (
	sessionCmd = &cobra.Command{
		Use:    "session MINUTES",
		Hidden: true,
		Short:  fmt.Sprintf("Creates new access credentials to upload files to %s. Expires in MINUTES.", dist.CompanyName),
		Args:   cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			min, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return errors.Wrapf(err,
					"unable to use '%s' as the session duration in minutes.", args[0])
			}
			return GetSessionToken(min)
		},
	}
)
