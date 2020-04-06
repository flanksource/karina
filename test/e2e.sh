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

$BIN deploy calico -v

[[ -e ./test/install_certs.sh ]] && ./test/install_certs.sh


for app in base stubs dex vault; do
  $BIN deploy $app -v
done

for app in base stubs; do
  $BIN test $app --wait 200
done

$BIN vault init -v

$BIN deploy postgres-operator install -v

$BIN test postgres-operator --wait 420 --e2e

$BIN deploy all -v

$BIN deploy opa bundle automobile -v

$BIN deploy opa policies test/opa/policies -v

failed=false
if ! $BIN test all -v --wait 240 --junit-path test-results/results.xml; then
  failed=true
fi

if ! $BIN test opa test/opa/opa-fixtures --wait 180 --junit-path test-results/opa-results.xml; then
  failed=true
fi

mkdir -p artifacts
$BIN snapshot --output-dir snapshot -v --include-specs=true --include-logs=true --include-events=true
zip -r artifacts/snapshot.zip snapshot/*

if [[ "$failed" = true ]]; then
  exit 1
fi

