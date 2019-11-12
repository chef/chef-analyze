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

package config

import "github.com/chef/chef-analyze/pkg/dist"

//
// The intend of this file is to have a single place where we can easily
// visualize the list of all error messages that we present to users.
//

var (
	ConfigTomlNotFoundErr = `
  config.toml file not found. (default: $HOME/.chef-workstation/config.toml)

  setup your local configuration file by following this documentation:
    - ` + dist.WSWebsite + `/docs/reference/config/
`
	MalformedConfigTomlFileErr = `
  unable to parse config.toml file.

  verify the format of the configuration file by following this documentation:
    - ` + dist.WSWebsite + `/docs/reference/config/
`
)
