#!/bin/bash
BIN=./.bin/platform-cli
mkdir -p .bin
export PLATFORM_CONFIG=test/common.yaml
export GO_VERSION=${GO_VERSION:-1.13}
export KUBECONFIG=~/.kube/config
NAME=$(basename $(git remote get-url origin | sed 's/\.git//'))
GITHUB_USER=$(basename $(dirname $(git remote get-url origin | sed 's/\.git//')))
GITHUB_USER=${GITHUB_USER##*:}
MASTER_HEAD=$(curl https://api.github.com/repos/$GITHUB_USER/$NAME/commits/master | jq -r '.sha')

if git log $MASTER_HEAD..$CIRCLE_SHA1 | grep "skip e2e"; then
  circleci-agent step halt
  exit 0
fi

if ! which gojsontoyaml 2>&1 > /dev/null; then
  go get -u github.com/brancz/gojsontoyaml
fi

go version

if go version | grep  go$GO_VERSION; then
  make pack build
else
  docker run --rm -it -v $PWD:$PWD -v /go:/go -w $PWD --entrypoint make -e GOPROXY=https://proxy.golang.org golang:$GO_VERSION pack build
fi

if [[ "$KUBECONFIG" != "$HOME/.kube/kind-config-kind" ]] ; then
  $BIN ca generate --name ingress-ca --cert-path .certs/ingress-ca-crt.pem --private-key-path .certs/ingress-ca-key.pem --password foobar  --expiry 1
  $BIN ca generate --name sealed-secrets --cert-path .certs/sealed-secrets-crt.pem --private-key-path .certs/sealed-secrets-key.pem --password foobar  --expiry 1
  $BIN provision kind-cluster
fi

$BIN version

$BIN deploy phases --base --stubs --dex --calico -v

[[ -e ./test/install_certs.sh ]] && ./test/install_certs.sh

# wait for the base deployment with stubs to come up healthy
$BIN test phases --base --stubs --wait 120 --progress=false

$BIN deploy phases --vault --postgres-operator -v

$BIN vault init -v

$BIN deploy all -v

# deploy the opa bundles first, as they can take some time to load, this effectively
# parallelizes this work to make the entire test complete faster
$BIN opa bundle automobile -v
# wait for up to 4 minutes, rerunning tests if they fail
# this allows for all resources to reconcile and images to finish downloading etc..
$BIN test all -v --wait 240 --progress=false

failed=false

# e2e do not use --wait at the run level, if needed each individual test implements
# its own wait. e2e tests should always pass once the non e2e have passed
if ! $BIN test all --e2e --progress=false -v --junit-path test-results/results.xml; then
  failed=true
fi

mkdir -p artifacts
$BIN snapshot --output-dir snapshot -v --include-specs=true --include-logs=true --include-events=true
zip -r artifacts/snapshot.zip snapshot/*

if [[ "$failed" = true ]]; then
  exit 1
fi

