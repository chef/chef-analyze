#!/bin/bash

set -euo pipefail

# the generated code coverage profile coming from our habitat studio
COVERAGE_TXT="coverage/coverage.txt"
# extract the repository name from the environment variable BUILDKITE_REPO
# example: "https://github.com/chef/chef-analyze.git"
REPO_NAME=$(echo $BUILDKITE_REPO | cut -d/ -f5 | cut -d. -f1)
# Github API to comment back to the open pull request
COMMENTS_URL="https://api.github.com/repos/${BUILDKITE_ORGANIZATION_SLUG}/${REPO_NAME}/issues/${BUILDKITE_PULL_REQUEST}/comments"

function post_data() {
  cat <<EOF
{"body":"# Code Coverage \\n$(echo `code_coverage_table`)"}
EOF
}

function code_coverage_table() {
  cat <<EOF
\\nCode details | Function name | % Coverage
\\n------------ | ------------- | ----------
\\n
EOF
  awk '{ printf "%s | %s | %s \\n", $1, $2, $3 }' < $COVERAGE_TXT
}

hab studio run "source .studiorc && code_coverage"

curl -H "Authorization: token $GITHUB_TOKEN" "$COMMENTS_URL" -d "$(post_data)" >/dev/null
