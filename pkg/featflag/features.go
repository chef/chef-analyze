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

package featflag

import (
	"fmt"
	"os"
	"strings"

	"github.com/chef/chef-analyze/pkg/config"
)

// this go library is an abstraction that manipulates feature flags through
// environment variables and a configuration file, by default it contains
// global flags that can be used in multiple go packages.
//
// example 1: use a global feature flag
// ```go
// if featflag.ChefFeatAnalyze.Enabled() {
//   // the analyze feature is enabled, act upon it
// }
// ```
//
// example 2: define a local feature flag
// ```go
// chefFeatXYZ := featflag.New("CHEF_FEAT_XYZ", "xyz")
// if chefFeatXYZ.Enabled() {
//   // the XYZ feature is enabled, act upon it
// }
// ```
type Feature struct {
	// the key associated to the feature flag defined inside the configuration file (config.toml)
	//
	// example of a config key:
	// ```toml
	// [features]
	// analyze = true
	// xyz = false
	// ```
	configKey string

	// the environment variable name associated to the feature flag
	//
	// example of environment variables:
	// CHEF_FEAT_ANALYZE=true
	// CHEF_FEAT_XYZ=true
	envName string
}

// global vs local feature flags
//
// when do I define a global feature flag?
// > when that flag is being used by multiple packages
var (
	// this special feature will enable all features at once
	ChefFeatAll = Feature{
		configKey: "all",
		envName:   "CHEF_FEAT_ALL",
	}

	// enables the chef-analyze feature
	ChefFeatAnalyze = Feature{
		configKey: "analyze",
		envName:   "CHEF_FEAT_ANALYZE",
	}

	// a list of all feature flags, global and local
	featureFlags = []Feature{ChefFeatAll, ChefFeatAnalyze}

	// config instance to access feature keys
	cfg *config.Config
)

func init() {
	c, _ := config.New()
	cfg = &c
}

// registers a new feature flag
//
// example of a new feature flag called 'foo':
// ```go
// chefFeatFoo := featflag.New("CHEF_FEAT_FOO", "foo")
// chefFeatFoo.Enabled()  // returns true if the feature flag is enabled
// ```
func New(envName, key string) Feature {
	feat := Feature{
		configKey: key,
		envName:   envName,
	}
	featureFlags = append(featureFlags, feat)
	return feat
}

func UseConfig(c *config.Config) {
	cfg = c
}

func ListAll() string {
	list := make([]string, len(featureFlags))
	for i, feat := range featureFlags {
		list[i] = feat.String()
	}

	return strings.Join(list, " ")
}

func GetFromKey(key string) (*Feature, bool) {
	for _, feat := range featureFlags {
		if feat.configKey == key {
			return &feat, true
		}
	}
	return nil, false
}

func GetFromEnv(name string) (*Feature, bool) {
	for _, feat := range featureFlags {
		if feat.envName == name {
			return &feat, true
		}
	}
	return nil, false
}

func (feat *Feature) String() string {
	return fmt.Sprintf("(%s:%s)", feat.Key(), feat.Env())
}

func (feat *Feature) Env() string {
	return feat.envName
}

func (feat *Feature) Key() string {
	return feat.configKey
}

func (feat *Feature) Equals(xfeat *Feature) bool {
	if feat.String() == xfeat.String() {
		return true
	}

	return false
}

// a feature flag is enabled when:
//
// 1) either the configured env variable is set to any value
// 2) or, the configured key is found inside the configuration file (config.toml)
//
// (the verification is done in the above order)
func (feat *Feature) Enabled() bool {
	if !feat.Equals(&ChefFeatAll) && ChefFeatAll.Enabled() {
		return true
	}

	if os.Getenv(feat.Env()) != "" {
		// users can use any value to enable a feature flag
		//
		// example:
		// CHEF_FEAT_ALL=true
		// CHEF_FEAT_ALL=1
		return true
	}

	return feat.valueFromConfig()
}

func (feat *Feature) valueFromConfig() bool {
	if cfg == nil {
		return false
	}

	value, ok := cfg.Features[feat.Key()]
	if ok {
		return value
	}

	return false
}
