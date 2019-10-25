#!/bin/bash
go version

if [[ ! -e ./kind ]]; then
  curl -Lo ./kind https://github.com/kubernetes-sigs/kind/releases/download/v0.5.1/kind-$(uname)-amd64
  chmod +x ./kind
fi

if ! which gojsontoyaml 2>&1 ; then
  go get -u github.com/brancz/gojsontoyaml
fi

kubernetes_version=$(cat test/common.yml | gojsontoyaml -yamltojson | jq -r '.kubernetes.version')
./kind create cluster --image kindest/node:${kubernetes_version} --config test/kind.config.yaml
KUBECONFIG="$(./kind get kubeconfig-path --name="kind")"
go env
make setup
make pack
PLATFORM_CONFIG=test/common.yml
platform-cli version
platform-cli deploy base  -vvv
platform-cli deploy stubs  -vvv
platform-cli deploy all  -vvv
platform-cli test all -vvv
