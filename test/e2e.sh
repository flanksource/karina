#!/bin/bash
BIN=./.bin/platform-cli
mkdir -p .bin
export PLATFORM_CONFIG=test/common.yml
if [[ ! -e ./kind ]]; then
  curl -Lo ./kind https://github.com/kubernetes-sigs/kind/releases/download/v0.5.1/kind-$(uname)-amd64
  chmod +x ./kind
fi

if ! which gojsontoyaml 2>&1 /dev/null; then
  go get -u github.com/brancz/gojsontoyaml
fi

if ! which expenv 2>&1 > /dev/null; then
  wget https://github.com/CrunchyData/postgres-operator/releases/download/v4.1.0/expenv
  chmod +x expenv
  sudo mv expenv /usr/local/bin
fi

docker run --rm -it -v $PWD:$PWD -v /go:/go -w $PWD --entrypoint make golang:1.12 setup pack

kubernetes_version=$(cat test/common.yml | gojsontoyaml -yamltojson | jq -r '.kubernetes.version')
./kind create cluster --image kindest/node:${kubernetes_version} --config test/kind.config.yaml
export KUBECONFIG="$(./kind get kubeconfig-path --name="kind")"
$BIN version
$BIN deploy base -vvv
$BIN deploy calico -vvv
$BIN deploy stubs -vvv


# wait for base to be up for up to +- 200 seconds
for i in {1..10}; do
 if $BIN test base; then
    break
  fi
  sleep 20
done

$BIN deploy all -vvv
$BIN test all -vvv
