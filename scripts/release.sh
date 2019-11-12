#!/bin/bash
set -x
NAME=$(basename $(git remote get-url origin | sed 's/\.git//'))
GITHUB_USER=$(basename $(dirname $(git remote get-url origin | sed 's/\.git//')))
GITHUB_USER=${GITHUB_USER##*:}
TAG=$(git tag --points-at HEAD )
if [[ "$TAG" == "" ]];  then
  echo "Skipping release of untagged commit"
  exit 0
fi

VERSION="v$TAG built $(date)"

make pack linux darwin compress

echo Releasing
github-release release -u $GITHUB_USER -r ${NAME} --tag $TAG
echo Uploading $NAME
github-release upload -u $GITHUB_USER -r ${NAME} --tag $TAG -n ${NAME} -f .bin/${NAME}
echo Uploading ${NAME}_osx
github-release upload -u $GITHUB_USER -r ${NAME} --tag $TAG -n ${NAME}_osx -f .bin/${NAME}_osx
