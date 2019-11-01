#!/bin/bash

set -eou pipefail

export COVERAGE_DATE
COVERAGE_DATE="$(date +"%Y-%m-%dT%H:%M:%S")"

COVERAGE_FOLDER=$(echo "$EXPEDITOR_SCM_SETTINGS_REPO" | sed -e 's/\//-/')
COVERAGE_FOLDER="$COVERAGE_FOLDER-$EXPEDITOR_SCM_SETTINGS_BRANCH"

# Generate code coverage to upload to S3 bucket and update to master
hab studio run "source .studiorc && code_coverage"

# Upload coverage to S3 bucket so that we can use them to update the % of coverage to master
# @afiune we can research how to use the buildkite artifacts instead
aws --profile chef-cd s3 sync coverage "s3://chef-workstation-coverage/$COVERAGE_FOLDER/current"
aws --profile chef-cd s3 sync coverage "s3://chef-workstation-coverage/$COVERAGE_FOLDER/$COVERAGE_DATE"
