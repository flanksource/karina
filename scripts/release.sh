#!/bin/bash
NAME=$(basename $(git remote get-url origin | sed 's/\.git//'))
GITHUB_USER=$(basename $(dirname $(git remote get-url origin | sed 's/\.git//')))
GITHUB_USER=${GITHUB_USER##*:}
TAG=$(git tag --points-at HEAD )
if [[ "$TAG" == "" ]];  then
  echo "Skipping release of untagged commit"
  exit 0
fi


if ! which goreleaser 2>&1 > /dev/null; then
  # need to pin the version
  wget -nv https://github.com/goreleaser/goreleaser/releases/download/v0.108.0/goreleaser_amd64.deb
  sudo dpkg -i goreleaser_amd64.deb
fi

if ! which rpmbuild 2>&1 > /dev/null; then
  sudo apt-get update && sudo apt-get install -y rpm
fi

git stash
git clean -fd
goreleaser --rm-dist

GO111MODULE=off go get github.com/aktau/github-release
go get github.com/aktau/github-release
github-release upload -u $GITHUB_USER -r ${NAME} --tag $TAG -n ${NAME} -f dist/linux_amd64/${NAME}
github-release upload -u $GITHUB_USER -r ${NAME} --tag $TAG -n ${NAME}_osx -f dist/darwin_amd64/${NAME}
github-release upload -u $GITHUB_USER -r ${NAME} --tag $TAG -n ${NAME}.exe -f dist/windows_amd64/${NAME}.exe
