#!/bin/bash
BIN=./.bin/platform-cli
mkdir -p .bin
export PLATFORM_CONFIG=test/common.yml
export GO_VERSION=${GO_VERSION:-1.13}
export KUBECONFIG=~/.kube/config

export

git log --name-only -n 1

if git log origin/master..HEAD | grep "skip e2e"; then
  circleci-agent step halt
  exit 0
fi

if [[ ! -e ./kind ]]; then
  curl -Lo ./kind https://github.com/kubernetes-sigs/kind/releases/download/v0.5.1/kind-$(uname)-amd64
  chmod +x ./kind
fi

if ! which gojsontoyaml 2>&1 > /dev/null; then
  go get -u github.com/brancz/gojsontoyaml
fi

if ! which expenv 2>&1 > /dev/null; then
  if [[ "$OSTYPE" =~ ^darwin ]]; then
    wget https://github.com/CrunchyData/postgres-operator/releases/download/v4.2.0/expenv-mac
    chmod +x expenv-mac
    sudo mv expenv-mac /usr/local/bin/expenv
  else
    wget https://github.com/CrunchyData/postgres-operator/releases/download/v4.2.0/expenv
    chmod +x expenv
    sudo mv expenv /usr/local/bin
  fi
fi

go version

if go version | grep  go$GO_VERSION; then
  make pack build
else
  docker run --rm -it -v $PWD:$PWD -v /go:/go -w $PWD --entrypoint make -e GOPROXY=https://proxy.golang.org golang:$GO_VERSION pack build
fi

# kubernetes_version=$(cat test/common.yml | gojsontoyaml -yamltojson | jq -r '.kubernetes.version')
if [[ "$KUBECONFIG" != "$HOME/.kube/kind-config-kind" ]] ; then
  $BIN provision kind-cluster
fi

$BIN version

$BIN deploy calico -v

[[ -e ./test/install_certs.sh ]] && ./test/install_certs.sh

.bin/kubectl -n kube-system set env daemonset/calico-node FELIX_IGNORELOOSERPF=true

$BIN deploy base -v

$BIN deploy stubs -v

$BIN test base --wait 200

$BIN deploy pgo install -v

$BIN test pgo --wait 200

$BIN deploy harbor -v

$BIN deploy all -v

$BIN deploy opa install -v

$BIN deploy velero

$BIN deploy fluentd

$BIN deploy eck

$BIN deploy opa policies test/opa/policies -v

echo "Sleeping for 30s, waiting for OPA policies to load"
sleep 30

failed=false
if ! $BIN test all -v --wait 240 --junit-path test-results/results.xml; then
  failed=true
fi

if ! $BIN test opa test/opa/opa-fixtures --junit-path test-results/opa-results.xml; then
  failed=true
fi

mkdir -p artifacts
$BIN snapshot --output-dir snapshot -v
zip -r artifacts/snapshot.zip snapshot/*

if [[ "$failed" = true ]]; then
  exit 1
fi

