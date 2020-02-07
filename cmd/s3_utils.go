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
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/chef/go-libs/config"
	"github.com/cheggaaa/pb/v3"
	"github.com/pkg/errors"

	"github.com/chef/chef-analyze/pkg/dist"
)

const analyzeTokensDir = "tokens" // Used for $HOME/.chef-workstation/tokens

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
	// shared credentials (~/.aws/credentials), or EC2 Instance Role
	awsSession := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))

	var (
		fileName = path.Base(filePath)
		uploader = s3manager.NewUploader(awsSession)
		progress = pb.StartNew(int(fileInfo.Size()))
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

func GetSessionToken(minDuration int64) error {
	var (
		awsSession = session.Must(session.NewSession())
		svc        = sts.New(awsSession)
		input      = &sts.GetSessionTokenInput{
			DurationSeconds: aws.Int64(minDuration * 60),
		}
	)

	result, err := svc.GetSessionToken(input)
	if err != nil {
		return err
	}

	fmt.Printf("A new session has been created and will be active for %d minutes.\n", minDuration)
	fmt.Printf(
		"Share these environment variables with a user that desires to upload files to %s:\n\n",
		dist.CompanyName)
	fmt.Printf("* Unix systems:\n%s", awsCredentialsToUnixVariables(result))
	fmt.Printf("\n* Windows systems:\n%s\n", awsCredentialsToPowershellVariables(result))
	return saveSessionToken(result, minDuration)
}

func saveSessionToken(token *sts.GetSessionTokenOutput, min int64) error {
	if token == nil {
		return nil
	}

	wsDir, err := config.ChefWorkstationDir()
	if err != nil {
		return err
	}

	var (
		tokensDir    = filepath.Join(wsDir, analyzeTokensDir)
		timestamp    = time.Now().Format("20060102150405")
		tokenName    = fmt.Sprintf("token-%dm-%s", min, timestamp)
		sessionToken = filepath.Join(tokensDir, tokenName)
		shFileName   = filepath.Join(tokensDir, fmt.Sprintf("%s.sh", tokenName))
		ps1FileName  = filepath.Join(tokensDir, fmt.Sprintf("%s.ps1", tokenName))
	)
	err = os.MkdirAll(tokensDir, os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "unable to create %s/ directory", analyzeTokensDir)
	}

	sessionFile, err := os.Create(sessionToken)
	if err != nil {
		return errors.Wrap(err, "unable to save session token")
	}
	sessionFile.WriteString(token.String())
	sessionFile.Close()
	fmt.Printf("Token payload saved to %s\n", sessionToken)

	shFile, err := os.Create(shFileName)
	if err != nil {
		return errors.Wrap(err, "unable to save .sh file")
	}
	shFile.WriteString(awsCredentialsToUnixVariables(token))
	shFile.Close()
	fmt.Printf("Unix shell file saved to %s\n", shFileName)

	ps1File, err := os.Create(ps1FileName)
	if err != nil {
		return errors.Wrap(err, "unable to save .ps1 file")
	}
	ps1File.WriteString(awsCredentialsToPowershellVariables(token))
	ps1File.Close()
	fmt.Printf("Powershell file saved to %s\n", ps1FileName)

	return nil
}

func awsCredentialsToUnixVariables(token *sts.GetSessionTokenOutput) string {
	if token == nil {
		return ""
	}
	return fmt.Sprintf(
		"export AWS_ACCESS_KEY_ID=\"%s\"\nexport AWS_SECRET_ACCESS_KEY=\"%s\"\nexport AWS_SESSION_TOKEN=\"%s\"\n",
		*token.Credentials.AccessKeyId, *token.Credentials.SecretAccessKey, *token.Credentials.SessionToken)
}

func awsCredentialsToPowershellVariables(token *sts.GetSessionTokenOutput) string {
	if token == nil {
		return ""
	}
	return fmt.Sprintf(
		"$Env:AWS_ACCESS_KEY_ID = \"%s\"\n$Env:AWS_SECRET_ACCESS_KEY = \"%s\"\n$Env:AWS_SESSION_TOKEN = \"%s\"\n",
		*token.Credentials.AccessKeyId, *token.Credentials.SecretAccessKey, *token.Credentials.SessionToken)
}
