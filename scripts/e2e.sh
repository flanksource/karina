#!/bin/bash

args=""
if [[ "$1" != "" ]]; then
  args=" -run $1"
fi
export CWD_VOL=$(docker volume create)
container=$(docker create -v $CWD_VOL:$PWD alpine /bin/sleep 30)
cleanup() {
  echo "Cleaning up"
  docker rm --force $container
  docker volume rm $CWD_VOL
}
docker cp $PWD $container:/$(dirname $PWD)
trap cleanup EXIT
mkdir -p test-output
go test -v ./test -race -coverprofile=integ.txt -covermode=atomic $args | tee e2e.out
set -euxo pipefail
cat e2e.out | go2xunit --fail -output test-output/$(date +%Y%m%d%M%H%M%S).xml
