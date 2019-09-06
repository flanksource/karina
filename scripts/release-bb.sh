
#!/bin/bash

REMOTE=bb
NAME=$(basename $(git remote get-url $REMOTE | sed 's/\.git//'))
USER=$(basename $(dirname $(git remote get-url $REMOTE | sed 's/\.git//')))
USER=${USER##*:}
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

goreleaser release --snapshot --rm-dist

for file in $(find ./dist -type f -name "*.tar.gz" -o -name "*.deb" -o -name "*.rpm"); do
  echo "Uploading $file"
  curl -X POST --user "${BB_PASSWORD}" "https://api.bitbucket.org/2.0/repositories/${USER}/${NAME}/downloads" --form files=@"$file"
done

mv dist/darwin_amd64/${NAME} dist/darwin_amd64/${NAME}_osx
curl -X POST --user "${BB_PASSWORD}" "https://api.bitbucket.org/2.0/repositories/${USER}/${NAME}/downloads" --form files=@"dist/linux_amd64/${NAME}"
curl -X POST --user "${BB_PASSWORD}" "https://api.bitbucket.org/2.0/repositories/${USER}/${NAME}/downloads" --form files=@"dist/windows_amd64/${NAME}.exe"
curl -X POST --user "${BB_PASSWORD}" "https://api.bitbucket.org/2.0/repositories/${USER}/${NAME}/downloads" --form files=@"dist/darwin_amd64/${NAME}_osx"

