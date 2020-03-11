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

package reporting

import (
	"crypto/sha256"
	"fmt"

	"github.com/chef/go-libs/config"
	"github.com/chef/go-libs/credentials"
)

// Reporting holds state for the command
type Reporting struct {
	// our Config (default: $HOME/.chef-workstation/config.toml)
	config.Config
	// our Credentials (default: $HOME/.chef/credentials)
	credentials.Credentials
	// skip the SSL certificate verification
	NoSSLVerify bool
}

// OverrideFunc to override any particular setting from the a reporting struct
type OverrideFunc func(*Reporting)

// NewDefault returns a reporting Config instance using the defaults
func NewDefault(overrides ...OverrideFunc) (Reporting, error) {
	rCfg := Reporting{}

	creds, err := credentials.NewDefault()
	if err != nil {
		return rCfg, err
	}
	rCfg.Credentials = creds

	// reporting doesn't really "require" the config.toml to exist
	// so we won't error if that happens
	if cfg, err := config.New(); err == nil {
		rCfg.Config = cfg
		// TODO @afiune log a debug message
	}

	for _, f := range overrides {
		f(&rCfg)
	}

	return rCfg, nil
}

// This returns the value referenced by `key` in `values`. If value is nil,
// it returns an empty string; otherwise it returns the original string.
func safeStringFromMap(values map[string]interface{}, key string) string {
	if values[key] == nil {
		return ""
	}
	return values[key].(string)
}

// Converts a string to its sha256 hash value.
// Keeps blank strings blank.
func hashString(name string) string {
	if name == "" {
		return name
	}
	return fmt.Sprintf("%x", sha256.Sum256([]byte(name)))
}
