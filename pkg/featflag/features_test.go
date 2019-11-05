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

package featflag_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	featflag "github.com/chef/chef-analyze/pkg/featflag"
)

func TestGetFeatureFlag(t *testing.T) {
	featFlag, exist := featflag.Get("CHEF_FEAT_ALL")
	if assert.True(t, exist) {
		assert.Equal(t, featflag.ChefFeatAll, featFlag)
		assert.Equal(t, "CHEF_FEAT_ALL", featFlag.Name())
		assert.False(t, featFlag.Enabled())
	}
}

func TestGetUnknownFeatureFlag(t *testing.T) {
	featFlag, exist := featflag.Get("CHEF_FEAT_BAR")
	if assert.False(t, exist) {
		assert.Equal(t, featflag.Feature(-1), featFlag)
		assert.Equal(t, false, featFlag.Enabled())
	}
}

func TestAllFeatureFlagsDisabledByDefault(t *testing.T) {
	assert.False(t, featflag.ChefFeatAll.Enabled())
	assert.False(t, featflag.ChefFeatAnalyze.Enabled())
}

func TestEnableChefAnalyzeFeatureFlag(t *testing.T) {
	defer os.Setenv("CHEF_FEAT_ANALYZE", "")

	if err := os.Setenv("CHEF_FEAT_ANALYZE", "true"); err != nil {
		t.Fatal("unable to set environment variable")
	}
	assert.True(t, featflag.ChefFeatAnalyze.Enabled())
}

func TestEnableChefAnalyzeThroughGlobalFeatureFlag(t *testing.T) {
	defer os.Setenv("CHEF_FEAT_ALL", "")

	if err := os.Setenv("CHEF_FEAT_ALL", "true"); err != nil {
		t.Fatal("unable to set environment variable")
	}
	assert.True(t, featflag.ChefFeatAnalyze.Enabled())
}

func TestAddFooFeatureFlag(t *testing.T) {
	// Defining a new Feature Flag Foo
	var chefFeatFoo = featflag.New("CHEF_FEAT_FOO")

	assert.Equal(t, "CHEF_FEAT_FOO", chefFeatFoo.Name())
	assert.False(t, chefFeatFoo.Enabled())

	if err := os.Setenv("CHEF_FEAT_FOO", "1"); err != nil {
		t.Fatal("unable to set environment variable")
	}
	assert.True(t, chefFeatFoo.Enabled())
}

func TestUnknownFeatureFlag(t *testing.T) {
	// An unknown feature flag (that is, not defined with featflag.New())
	var chefFeatUnknown featflag.Feature = 999

	assert.Equal(t, "", chefFeatUnknown.Name())
	assert.False(t, chefFeatUnknown.Enabled())
}
