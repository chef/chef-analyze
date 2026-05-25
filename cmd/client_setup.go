//
// Copyright 2026 Chef Software, Inc.
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
	"github.com/chef/go-libs/credentials"
	"github.com/go-chef/chef"

	"github.com/chef/chef-analyze/pkg/reporting"
)

var (
	credentialsFromViper = credentials.FromViper
	createOutputDirs     = createOutputDirectories
	newChefClient        = reporting.NewChefClient
)

// setupChefClientFromFlags loads credentials, creates output directories,
// and returns a configured chef client using shared infra flags.
func setupChefClientFromFlags() (*chef.Client, error) {
	creds, err := credentialsFromViper(
		infraFlags.profile,
		overrideCredentials(),
	)
	if err != nil {
		return nil, err
	}

	if err := createOutputDirs(); err != nil {
		return nil, err
	}

	cfg := &reporting.Reporting{Credentials: creds}
	if infraFlags.noSSLverify {
		cfg.NoSSLVerify = true
	}

	return newChefClient(cfg)
}