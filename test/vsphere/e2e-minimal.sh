#!/bin/bash

# dependencies
if ! which make 2>&1 > /dev/null; then
  echo "make required"
  exit 1
fi
if ! which git 2>&1 > /dev/null; then
  echo "git required"
  exit 1
fi
if ! which go 2>&1 > /dev/null; then
  echo "go required"
  exit 1
fi
if ! which jq 2>&1 > /dev/null; then
  echo "jq required"
  exit 1
fi
if ! which mkisofs 2>&1 > /dev/null; then
  echo "mkisofs required"
  exit 1
fi
if ! which gojsontoyaml 2>&1 > /dev/null; then
  go get -u github.com/brancz/gojsontoyaml
fi

mkdir -p .bin
mkdir -p .certs

export TERM=xterm-256color
export GOPATH=$HOME/go
export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin
export GO_VERSION=${GO_VERSION:-1.14}

BIN=./.bin/karina

REPO=$(basename $(git remote get-url origin | sed 's/\.git//'))
MASTER_HEAD=$(curl https://api.github.com/repos/$GITHUB_REPOSITORY/commits/master | jq -r '.sha')
PR_NUM=$(echo $GITHUB_REF | awk 'BEGIN { FS = "/" } ; { print $3 }')
COMMIT_SHA="$GITHUB_SHA"

generate_cluster_id() {
  local prefix

  prefix=$(tr </dev/urandom -cd 'a-f0-9' | head -c 5)
  echo "e2e-${prefix}"
}

PLATFORM_CLUSTER_ID=$(generate_cluster_id)
export PLATFORM_CLUSTER_ID
export PLATFORM_OPTIONS_FLAGS="-e name=${PLATFORM_CLUSTER_ID} -e domain=${PLATFORM_CLUSTER_ID}.lab.flanksource.com -vv"
unset KUBECONFIG
export PLATFORM_CONFIG=test/vsphere/e2e-platform-minimal.yaml

if git log $MASTER_HEAD..$COMMIT_SHA | grep "skip e2e"; then
  exit 0
fi

printf "\n\n\n\n$(tput bold)Build steps$(tput setaf 7)\n"
go version
make setup
make pack build

$BIN version

printf "\n\n\n\n$(tput bold)Generate Certs$(tput setaf 7)\n"
$BIN ca generate --name root-ca --cert-path .certs/root-ca.crt --private-key-path .certs/root-ca.key --password foobar  --expiry 1
$BIN ca generate --name ingress-ca --cert-path .certs/ingress-ca.crt --private-key-path .certs/ingress-ca.key --password foobar  --expiry 1
$BIN ca generate --name sealed-secrets --cert-path .certs/sealed-secrets-crt.pem --private-key-path .certs/sealed-secrets-key.pem --password foobar  --expiry 1

printf "\n\n\n\n$(tput bold)Provision Cluster$(tput setaf 7)\n"
$BIN provision vsphere-cluster $PLATFORM_OPTIONS_FLAGS

printf "\n\n\n\n$(tput bold)Basic Deployments$(tput setaf 7)\n"

$BIN deploy phases --calico $PLATFORM_OPTIONS_FLAGS
$BIN deploy phases --base $PLATFORM_OPTIONS_FLAGS
$BIN deploy phases --stubs $PLATFORM_OPTIONS_FLAGS
$BIN deploy phases --dex $PLATFORM_OPTIONS_FLAGS

printf "\n\n\n\n$(tput bold)Up?$(tput setaf 7)\n"
# wait for the base deployment with stubs to come up healthy
$BIN test phases --base --stubs --wait 120 --progress=false $PLATFORM_OPTIONS_FLAGS

printf "\n\n\n\n$(tput bold)All Deployments$(tput setaf 7)\n"
$BIN deploy all $PLATFORM_OPTIONS_FLAGS

printf "\n\n\n\n$(tput bold)Tests$(tput setaf 7)\n"
## deploy the opa bundles first, as they can take some time to load, this effectively
## parallelizes this work to make the entire test complete faster
$BIN opa bundle automobile -v $PLATFORM_OPTIONS_FLAGS
# wait for up to 4 minutes, rerunning tests if they fail
# this allows for all resources to reconcile and images to finish downloading etc..
$BIN test  base --wait 240 --progress=false $PLATFORM_OPTIONS_FLAGS

failed=false

## e2e do not use --wait at the run level, if needed each individual test implements
## its own wait. e2e tests should always pass once the non e2e have passed
if ! $BIN test  all --e2e --progress=false --junit-path test-results/results.xml $PLATFORM_OPTIONS_FLAGS; then
  failed=true
fi

printf "\n\n\n\n$(tput bold)Reporting$(tput setaf 7)\n"
wget -nv https://github.com/flanksource/build-tools/releases/download/v0.7.0/build-tools
chmod +x build-tools
./build-tools gh report-junit $GITHUB_REPOSITORY $PR_NUM ./test-results/results.xml --auth-token $GIT_API_KEY \
      --success-message="vSphere minimal tests - commit $COMMIT_SHA" \
      --failure-message="vSphere minimal tests - :neutral_face: commit $COMMIT_SHA had some failures or skipped tests. **Is it OK?**"

mkdir -p artifacts
$BIN snapshot --output-dir snapshot -v --include-specs=true --include-logs=true --include-events=true $PLATFORM_OPTIONS_FLAGS
zip -r artifacts/snapshot.zip snapshot/*

$BIN terminate-orphans $PLATFORM_OPTIONS_FLAGS || echo "Orphans not terminated."
$BIN cleanup $PLATFORM_OPTIONS_FLAGS

if [[ "$failed" = true ]]; then
  exit 1
fi
