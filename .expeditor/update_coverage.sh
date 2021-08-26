#!/bin/bash

set -eou pipefail

COVERAGE_FOLDER=$(echo "$EXPEDITOR_SCM_SETTINGS_REPO" | sed -e 's/\//-/')
COVERAGE_FOLDER="$COVERAGE_FOLDER-$EXPEDITOR_SCM_SETTINGS_BRANCH"

aws --profile chef-cd s3 cp "s3://chef-workstation-coverage/$COVERAGE_FOLDER/current" coverage --recursive

COVERAGE=$(grep -w total: coverage/coverage.txt| awk '{print $NF}' | sed -e 's/%//')
COVERAGE_MASTER=$(cat CODE_COVERAGE)

if [ "$COVERAGE_MASTER" == "$COVERAGE" ]; then
  echo "This change neither increased nor decreased the code coverage from main. (${COVERAGE_MASTER}%)"
  echo "TIP: Add the GH tag 'Expeditor: Skip Code Coverage' to avoid triggering this task."
  exit 0
fi

case $COVERAGE in
  100) COLOR=brightgreen ;;
  9[0-9]*) COLOR=brightgreen ;;
  8[0-9]*) COLOR=green ;;
  7[0-9]*) COLOR=yellowgreen ;;
  6[0-9]*) COLOR=yellow ;;
  5[0-9]*) COLOR=orange ;;
  *) COLOR=red ;;
esac

sed -i -r "s|badge/coverage-[0-9]+.[0-9]+\\%25-[a-z]+\\)|badge/coverage-${COVERAGE}\\%25-${COLOR}\\)|" README.md
git add README.md

echo "$COVERAGE" > CODE_COVERAGE
git add CODE_COVERAGE

git commit --message "Update code coverage to ${COVERAGE}%"
git push origin main
