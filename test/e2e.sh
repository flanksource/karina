#!/bin/bash
BIN=./.bin/platform-cli
PLATFORM_CONFIG=test/common.yml
if [[ ! -e ./kind ]]; then
  curl -Lo ./kind https://github.com/kubernetes-sigs/kind/releases/download/v0.5.1/kind-$(uname)-amd64
  chmod +x ./kind
fi

if ! which gojsontoyaml 2>&1 ; then
  go get -u github.com/brancz/gojsontoyaml
fi

docker run --rm -it -v $PWD:$PWD -v /go:/go -w $PWD --entrypoint make golang:1.12 setup pack

kubernetes_version=$(cat test/common.yml | gojsontoyaml -yamltojson | jq -r '.kubernetes.version')
./kind create cluster --image kindest/node:${kubernetes_version} --config test/kind.config.yaml
KUBECONFIG="$(./kind get kubeconfig-path --name="kind")"

$BIN version
$BIN deploy base  -vvv
$BIN deploy stubs  -vvv
$BIN deploy all  -vvv
$BIN test all -vvv
