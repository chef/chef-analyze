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

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFeatureFlag(t *testing.T) {
	out, err, exitcode := ChefAnalyzeNoFeatureFlag("help")
	assert.Empty(t,
		err.String(),
		"STDERR should be empty")
	assert.Equal(t, 0, exitcode,
		"EXITCODE is not the expected one")
	assert.Contains(t, out.String(),
		"E003",
		"STDOUT does not contain expected error code")
	assert.Contains(t,
		out.String(),
		"'analyze' is experimental and in development.",
		"STDOUT message doesn't match")
	assert.Contains(t,
		out.String(),
		"enable it by setting the following environment variable:",
		"STDOUT message doesn't match")
	assert.Contains(t,
		out.String(),
		"CHEF_FEAT_ANALYZE=true",
		"STDOUT message doesn't match")
	assert.Contains(t,
		out.String(),
		"permanently enable it by modifying $HOME/.chef-workstation/config.toml",
		"STDOUT message doesn't match")
	assert.Contains(t,
		out.String(),
		"[features]",
		"STDOUT message doesn't match")
	assert.Contains(t,
		out.String(),
		"analyze = true",
		"STDOUT message doesn't match")
}
