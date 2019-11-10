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

which packr2 2>&1 > /dev/null || go get github.com/gobuffalo/packr/v2/packr2
which github-release 2>&1 > /dev/null || go get github.com/aktau/github-release
which upx 2>&1 >  /dev/null  || (sudo apt-get update && sudo apt-get install -y upx-ucl)

echo Building $NAME $VERSION
GOOS=linux packr2 build -o $NAME -ldflags "-X \"main.version=$VERSION\""  main.go
echo Building ${NAME}_osx $VERSION
GOOS=darwin packr2 build -o ${NAME}_osx -ldflags "-X \"main.version=$VERSION\""  main.go
echo Compressing
upx ${NAME} ${NAME}_osx
echo Releasing
github-release release -u $GITHUB_USER -r ${NAME} --tag $TAG
echo Uploading $NAME
github-release upload -u $GITHUB_USER -r ${NAME} --tag $TAG -n ${NAME} -f ${NAME}
echo Uploading ${NAME}_osx
github-release upload -u $GITHUB_USER -r ${NAME} --tag $TAG -n ${NAME}_osx -f ${NAME}_osx
