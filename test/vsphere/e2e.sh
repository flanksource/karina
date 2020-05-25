#!/bin/bash

#export AWS_ACCESS_KEY_ID=$SOPS_AWS_ACCESS_KEY_ID
#export AWS_SECRET_ACCESS_KEY=$SOPS_AWS_SECRET_ACCESS_KEY



BIN=./.bin/karina
mkdir -p .bin

export GOPATH=$HOME/go
export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin


export PLATFORM_CONFIG=test/vsphere/e2e.yaml
export GO_VERSION=${GO_VERSION:-1.14}
export KUBECONFIG=~/.kube/config
REPO=$(basename $(git remote get-url origin | sed 's/\.git//'))
GITHUB_OWNER=$(basename $(dirname $(git remote get-url origin | sed 's/\.git//')))
GITHUB_OWNER=${GITHUB_OWNER##*:}
MASTER_HEAD=$(curl https://api.github.com/repos/$GITHUB_OWNER/$REPO/commits/master | jq -r '.sha')

PR_NUM=$(echo $GITHUB_REF | awk 'BEGIN { FS = "/" } ; { print $3 }')
COMMIT_SHA="$GITHUB_SHA"

if git log $MASTER_HEAD..$COMMIT_SHA | grep "skip e2e"; then
  #TODO: more halt required here?
  exit 0
fi

if ! which gojsontoyaml 2>&1 > /dev/null; then
  go get -u github.com/brancz/gojsontoyaml
fi

make setup

go version

if go version | grep  go$GO_VERSION; then
  make pack build
else
  docker run --rm -it -v $PWD:$PWD -v /go:/go -w $PWD --entrypoint make -e GOPROXY=https://proxy.golang.org golang:$GO_VERSION pack build
fi


$BIN ca generate --name root-ca --cert-path .certs/root-ca.crt --private-key-path .certs/root-ca.key --password foobar  --expiry 1
$BIN ca generate --name ingress-ca --cert-path .certs/ingress-ca.crt --private-key-path .certs/ingress-ca.key --password foobar  --expiry 1
$BIN ca generate --name sealed-secrets --cert-path .certs/sealed-secrets-crt.pem --private-key-path .certs/sealed-secrets-key.pem --password foobar  --expiry 1
$BIN provision vsphere-cluster || exit 1

$BIN version

$BIN kubeconfig admin -c test/vsphere/e2e.yaml > config.yaml
export KUBECONFIG=$PWD/config.yaml

$BIN deploy phases --calico -v
$BIN deploy phases --base -v
$BIN deploy phases --stubs -v
$BIN deploy phases --dex -v

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

wget https://github.com/flanksource/build-tools/releases/download/v0.7.0/build-tools
chmod +x build-tools
./build-tools gh report-junit $GITHUB_OWNER/platform-cli $PR_NUM ./test-results/results.xml --auth-token $GITHUB_TOKEN \
      --success-message="commit $COMMIT_SHA" \
      --failure-message=":neutral_face: commit $COMMIT_SHA had some failures or skipped tests. **Is it OK?**"

mkdir -p artifacts
$BIN snapshot --output-dir snapshot -v --include-specs=true --include-logs=true --include-events=true
zip -r artifacts/snapshot.zip snapshot/*


if [[ "$failed" = true ]]; then
  exit 1
fi
