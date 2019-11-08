#!/bin/bash
NAME=$(basename $(git remote get-url origin | sed 's/\.git//'))
GITHUB_USER=$(basename $(dirname $(git remote get-url origin | sed 's/\.git//')))
GITHUB_USER=${GITHUB_USER##*:}
TAG=$(git tag --points-at HEAD )
if [[ "$TAG" == "" ]];  then
  echo "Skipping release of untagged commit"
  exit 0
fi

go get -u github.com/gobuffalo/packr/v2/packr2

GOOS=linux packr2 build -o $NAME -ldflags "-X \"main.version=v$TAG built $(date "+%Y-%m-%d %H:%M:%S")\""  main.go

GOOS=darwin packr2 build -o ${NAME}_osx -ldflags "-X \"main.version=v$TAG built $(date "+%Y-%m-%d %H:%M:%S")\""  main.go

GO111MODULE=off go get github.com/aktau/github-release
go get github.com/aktau/github-release
github-release release -u $GITHUB_USER -r ${NAME} --tag $TAG
github-release upload -u $GITHUB_USER -r ${NAME} --tag $TAG -n ${NAME} -f ${NAME}_osx
github-release upload -u $GITHUB_USER -r ${NAME} --tag $TAG -n ${NAME}_osx -f ${NAME}_osx
