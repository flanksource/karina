#!/bin/bash
BIN=./.bin/platform-cli
mkdir -p .bin
export PLATFORM_CONFIG=test/common.yml
if [[ ! -e ./kind ]]; then
  curl -Lo ./kind https://github.com/kubernetes-sigs/kind/releases/download/v0.5.1/kind-$(uname)-amd64
  chmod +x ./kind
fi

if ! which gojsontoyaml 2>&1 > /dev/null; then
  go get -u github.com/brancz/gojsontoyaml
fi

if ! which expenv 2>&1 > /dev/null; then
  wget https://github.com/CrunchyData/postgres-operator/releases/download/v4.1.0/expenv
  chmod +x expenv
  sudo mv expenv /usr/local/bin
fi

if go version | grep  go1.12; then
  make pack build
else
  docker run --rm -it -v $PWD:$PWD -v /go:/go -w $PWD --entrypoint make -e GOPROXY=https://proxy.golang.org golang:1.12 pack build
fi

# kubernetes_version=$(cat test/common.yml | gojsontoyaml -yamltojson | jq -r '.kubernetes.version')
if [[ "$KUBECONFIG" != "$HOME/.kube/kind-config-kind" ]] ; then
  ./kind create cluster --image kindest/node:${VERSION} --config test/kind.config.yaml
  export KUBECONFIG="$(./kind get kubeconfig-path --name="kind")"
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

if [[ "$failed" = true ]]; then
  exit 1
fi

