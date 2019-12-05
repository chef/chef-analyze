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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

func StoreErrorsFromBuilders(timestamp string, errBuilders ...strings.Builder) error {
	var (
		errorsDir  = filepath.Join(AnalyzeCacheDir, "errors")
		errorsName = fmt.Sprintf("cookbooks-%s.err", timestamp)
		errorsPath = filepath.Join(errorsDir, errorsName)
	)

	// create reports directory
	err := os.MkdirAll(errorsDir, os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "unable to create reports/ directory")
	}

	// create a new errors file
	errorsFile, err := os.Create(errorsPath)
	if err != nil {
		return errors.Wrap(err, "unable to create report")
	}

	for _, errBldr := range errBuilders {
		if errBldr.Len() > 0 {
			errorsFile.WriteString(errBldr.String())
		}
	}

	fmt.Println("Error(s) found during the report generation, find more details at:")
	fmt.Printf("  => %s\n", errorsPath)
	return nil
}
