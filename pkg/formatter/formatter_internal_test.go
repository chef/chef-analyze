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

package formatter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringOrEmptyPlaceholder(t *testing.T) {
	assert.Equal(t, "-", stringOrEmptyPlaceholder(""))
	assert.Equal(t, "foo", stringOrEmptyPlaceholder("foo"))
}

func TestStringOrUnknownPlaceholder(t *testing.T) {
	assert.Equal(t, "unknown", stringOrUnknownPlaceholder(""))
	assert.Equal(t, "foo", stringOrUnknownPlaceholder("foo"))
}

func TestStringOrPlaceholder(t *testing.T) {
	assert.Equal(t, "placeholder", stringOrPlaceholder("", "placeholder"))
	assert.Equal(t, "foo", stringOrPlaceholder("foo", "placeholder"))
}
