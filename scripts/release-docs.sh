#!/bin/bash
if [[ "$CIRCLE_PR_NUMBER" != "" ]]; then
  echo Skipping release of a PR build
  circleci-agent step halt
  exit 0
fi

make deploy-docs
