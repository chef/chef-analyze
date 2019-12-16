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

package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	awscreds "github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/cheggaaa/pb/v3"
)

func UploadToS3(credential, bucket, filePath string) error {
	creds := awscreds.NewSharedCredentials(credential, "default")
	if _, err := creds.Get(); err != nil {
		return err
	}

	awsSession := session.New(&aws.Config{
		Credentials: creds,
	})

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	var (
		progress = pb.StartNew(int(fileInfo.Size()))
		reader   = progress.NewProxyReader(file)
		uploader = s3manager.NewUploader(awsSession)
		fileName = path.Base(filePath)
	)
	output, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileName),
		Body:   reader,
	})
	if err != nil {
		return err
	}

	progress.Finish()
	fmt.Printf("File uploaded to %s\n", output.Location)
	return nil
}
