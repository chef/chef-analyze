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

import "os"

type Feature int

const (
	ChefFeatAll Feature = iota
	ChefFeatAnalyze
)

var featureFlags = map[Feature]string{
	ChefFeatAll:     "CHEF_FEAT_ALL",
	ChefFeatAnalyze: "CHEF_FEAT_ANALYZE",
}

func New(envName string) Feature {
	feat := Feature(len(featureFlags) + 1 + int(ChefFeatAll))
	featureFlags[feat] = envName
	return feat
}

func Get(envName string) (Feature, bool) {
	for feat, name := range featureFlags {
		if envName == name {
			return feat, true
		}
	}
	return -1, false
}

func (feat Feature) Name() string {
	name, _ := featureFlags[feat]
	return name
}

func (feat Feature) Enabled() bool {
	if feat != ChefFeatAll && ChefFeatAll.Enabled() {
		return true
	}

	if os.Getenv(feat.Name()) == "" {
		return false
	}
	// @afiune do we want users to turn features with any value?
	// or do we want to be explicit:
	//
	// CHEF_FEAT_ALL=true
	// CHEF_FEAT_ALL=1
	return true
}
