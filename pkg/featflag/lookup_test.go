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

package featflag_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	subject "github.com/chef/chef-analyze/pkg/featflag"
)

func TestLookupUnknown(t *testing.T) {
	featFlag, exist := subject.Lookup("CHEF_FEAT_BAR")
	if assert.False(t, exist, "unknown feature flag should not exist") {
		assert.Nil(t, featFlag, "unknown feature flag should return nil")
	}
	featFlag, exist = subject.Lookup("bar")
	if assert.False(t, exist, "unknown feature flag should not exist") {
		assert.Nil(t, featFlag, "unknown feature flag should return nil")
	}
}

func TestLookupChefFeatAll(t *testing.T) {
	// look up for the environment variable name
	featFlag, exist := subject.Lookup("CHEF_FEAT_ALL")
	if assert.True(t, exist, "global CHEF_FEAT_ALL:all flag should exist") {
		assert.True(t,
			subject.ChefFeatAll.Equals(featFlag),
			"the feature flags don't match",
		)
		assert.False(t,
			featFlag.Enabled(),
			"global feature flag for all features should be disabled",
		)
		assert.Equal(t,
			"CHEF_FEAT_ALL", featFlag.Env(),
			"environment variable mismatch")
		assert.Equal(t,
			"all", featFlag.Key(),
			"config key mismatch")
		assert.Equal(t,
			"(all:CHEF_FEAT_ALL)", subject.ChefFeatAll.String(),
			"string conversion mismatch")
	}

	// look up for the config key
	featFlag, exist = subject.Lookup("all")
	if assert.True(t, exist, "global CHEF_FEAT_ALL:all flag should exist") {
		assert.True(t,
			subject.ChefFeatAll.Equals(featFlag),
			"the feature flags don't match",
		)
		assert.False(t,
			featFlag.Enabled(),
			"global feature flag for all features should be disabled",
		)
		assert.Equal(t,
			"CHEF_FEAT_ALL", featFlag.Env(),
			"environment variable mismatch")
		assert.Equal(t,
			"all", featFlag.Key(),
			"config key mismatch")
		assert.Equal(t,
			"(all:CHEF_FEAT_ALL)", subject.ChefFeatAll.String(),
			"string conversion mismatch")
	}
}

func TestGetFromEnvUnknown(t *testing.T) {
	featFlag, exist := subject.GetFromEnv("CHEF_FEAT_BAR")
	if assert.False(t, exist, "unknown feature flag should not exist") {
		assert.Nil(t, featFlag, "unknown feature flag should return nil")
	}
}

func TestGetFromEnvGlobalChefFeatAll(t *testing.T) {
	featFlag, exist := subject.GetFromEnv("CHEF_FEAT_ALL")
	if assert.True(t, exist, "global CHEF_FEAT_ALL:all flag should exist") {
		assert.True(t,
			subject.ChefFeatAll.Equals(featFlag),
			"the feature flags don't match",
		)
		assert.False(t,
			featFlag.Enabled(),
			"global feature flag for all features should be disabled",
		)
		assert.Equal(t,
			"CHEF_FEAT_ALL", featFlag.Env(),
			"environment variable mismatch")
		assert.Equal(t,
			"all", featFlag.Key(),
			"config key mismatch")
		assert.Equal(t,
			"(all:CHEF_FEAT_ALL)", subject.ChefFeatAll.String(),
			"string conversion mismatch")
	}
}

func TestGetFromKeyUnknown(t *testing.T) {
	featFlag, exist := subject.GetFromKey("bar")
	if assert.False(t, exist, "unknown feature flag should not exist") {
		assert.Nil(t, featFlag, "unknown feature flag should return nil")
	}
}

func TestGetFromKeyGlobalChefFeatAll(t *testing.T) {
	featFlag, exist := subject.GetFromKey("all")
	if assert.True(t, exist, "global CHEF_FEAT_ALL:all flag should exist") {
		assert.True(t,
			subject.ChefFeatAll.Equals(featFlag),
			"the feature flags don't match",
		)
		assert.False(t,
			featFlag.Enabled(),
			"global feature flag for all features should be disabled",
		)
		assert.Equal(t,
			"CHEF_FEAT_ALL", featFlag.Env(),
			"environment variable mismatch")
		assert.Equal(t,
			"all", featFlag.Key(),
			"config key mismatch")
		assert.Equal(t,
			"(all:CHEF_FEAT_ALL)", subject.ChefFeatAll.String(),
			"string conversion mismatch")
	}
}
