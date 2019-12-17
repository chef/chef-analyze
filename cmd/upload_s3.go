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
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/cheggaaa/pb/v3"
)

func UploadToS3(bucket, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	// default to the aws region where we, at Chef, will create the S3 buckets
	region := endpoints.UsEast1RegionID
	if envRegion := os.Getenv("AWS_REGION"); envRegion != "" {
		region = envRegion
	}

	// Load the default credential chain. Such as the environment,
	// shared credentials (~/.aws/credentials), or EC2 Instance Role.
	awsSession := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))

	var (
		fileName = path.Base(filePath)
		uploader = s3manager.NewUploader(awsSession)
		progress = pb.New(int(fileInfo.Size()))
		reader   = progress.NewProxyReader(file)
	)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileName),
		Body:   reader,
	})
	if err != nil {
		return err
	}

	progress.Finish()
	fmt.Printf("File uploaded to %s/%s\n", bucket, fileName)
	return nil
}
