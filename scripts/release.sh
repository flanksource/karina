#!/bin/bash
set -x
cd $GITHUB_WORKSPACE
GITHUB_USER=$(echo $GITHUB_REPOSITORY | cut -d/ -f1)
NAME=$(echo $GITHUB_REPOSITORY | cut -d/ -f2)
TAG=$(echo $GITHUB_REF | sed 's|refs/tags/||')
VERSION="v$TAG built $(date)"

make setup pack linux darwin compress

github-release release -u $GITHUB_USER -r ${NAME} --tag $TAG  || true
github-release upload -R -u $GITHUB_USER -r ${NAME} --tag $TAG -n ${NAME} -f .bin/${NAME}
github-release upload -R -u $GITHUB_USER -r ${NAME} --tag $TAG -n ${NAME}_osx -f .bin/${NAME}_osx
