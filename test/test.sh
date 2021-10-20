#!/bin/bash
if test -f ./.bin/karina; then
    BIN=./.bin/karina
    chmod +x $BIN
elif command -v karina; then
    BIN=$(command -v karina)
else
    echo "No karina binary detected"
    exit 127
fi

export KUBECONFIG=~/.kube/config
REPO=$(basename $(git remote get-url origin | sed 's/\.git//'))
GITHUB_OWNER=$(basename $(dirname $(git remote get-url origin | sed 's/\.git//')))
GITHUB_OWNER=${GITHUB_OWNER##*:}
MASTER_HEAD=$(curl -s https://api.github.com/repos/$GITHUB_OWNER/$REPO/commits/master | jq -r '.sha')
OS="$(uname | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m | sed -e 's/x86_64/amd64/' -e 's/\(arm\)\(64\)\?.*/\1\2/' -e 's/aarch64$/arm64/')"
export KUBERNETES_VERSION=${KUBERNETES_VERSION:-v1.19.11}
export SUITE=${SUITE:-minimal}
if [[ "$1" != "" ]]; then
    export SUITE=$1
fi
export CLUSTER_NAME=kind-$SUITE-$KUBERNETES_VERSION

if [[ "$CI" == "true" ]]; then
    kind delete cluster --name  $(kind get clusters) || echo "No cluster present when starting"
fi
export CONFIGURED_VALUE=$(openssl rand -base64 12)
if [[ "$ADDITIONAL_CONFIG" == "" ]]; then
    export PLATFORM_CONFIG=test/$SUITE.yaml
    echo Using config $PLATFORM_CONFIG
else
    export CONFIG_FILES="-c test/$SUITE.yaml $ADDITIONAL_CONFIG"
    echo Using config $CONFIG_FILES
fi
echo "::endgroup::"

function report() {

    if [[ "$CI" != "true" ]]; then
        echo "Skipping test report when not running in CI"
        return
    fi
    echo "::group::Uploading test results"
    set +e
    KREW=./krew-"${OS}_${ARCH}"
    export PATH="${KREW_ROOT:-$HOME/.krew}/bin:$PATH"
    if ! kubectl krew version 2&1>/dev/null; then
        cd "$(mktemp -d)"
        curl -fsSLO "https://github.com/kubernetes-sigs/krew/releases/latest/download/krew.tar.gz"
        tar zxvf krew.tar.gz
        "$KREW" install krew
    fi
    if ! kubectl resource-snapshot -v 2&1>/dev/null; then
        kubectl krew install resource-snapshot
    fi
    kubectl resource-snapshot

    if [[ -e test-results/results.xml ]]; then
        wget -nv -nc -O build-tools \
        https://github.com/flanksource/build-tools/releases/latest/download/build-tools && \
        chmod +x build-tools

        ./build-tools junit gh-workflow-commands test-results/results.xml
    fi
    mkdir -p artifacts
    if $BIN snapshot --output-dir snapshot -v --include-specs=true --include-logs=true --include-events=true $CONFIG_FILES ; then
        zip -r artifacts/snapshot.zip snapshot/*
    fi

    echo "::endgroup::"
}
trap report EXIT

set -e
echo "$(kubectl config current-context) != kind-$CLUSTER_NAME"

if [[ "$(kubectl config current-context)" != "kind-$CLUSTER_NAME" ]] ; then
    echo "::group::Provisioning"
    if [[ ! -e .certs/root-ca.key ]]; then
        $BIN ca generate --name root-ca --cert-path .certs/root-ca.crt --private-key-path .certs/root-ca.key --password foobar  --expiry 1 $CONFIG_FILES
        $BIN ca generate --name ingress-ca --cert-path .certs/ingress-ca.crt --private-key-path .certs/ingress-ca.key --password foobar  --expiry 1 $CONFIG_FILES
        $BIN ca generate --name sealed-secrets --cert-path .certs/sealed-secrets-crt.pem --private-key-path .certs/sealed-secrets-key.pem --password foobar  --expiry 1 $CONFIG_FILES
    fi
    if $BIN provision kind-cluster --trace -vv $CONFIG_FILES ; then
        echo "::endgroup::"
    else
        echo "::endgroup::"
        exit 1
    fi
fi

$BIN version

echo "::group::Deploying Base"
$BIN deploy phases --bootstrap --stubs -v --prune=false $CONFIG_FILES
echo "::endgroup::"

echo "::group::Waiting for Base"
# wait for the base deployment with stubs to come up healthy
$BIN test phases --bootstrap --stubs --wait 120 --progress=false --fail-on-error=false $CONFIG_FILES
echo "::endgroup::"

echo "::group::Deploy All"
$BIN deploy all --exclude crds --prune=false -v $CONFIG_FILES || (echo "::error::Error while deploying" && exit 1)
echo "::endgroup::"

echo "::group::Test Dry Run"
# wait for up to 8 minutes, rerunning tests if they fail
# this allows for all resources to reconcile and images to finish downloading etc..
$BIN test all -v --wait 480 --progress=false --fail-on-error=false $CONFIG_FILES
echo "::endgroup::"

echo "::group::Final Test Run"
# E2E do not use --wait at the run level, if needed each individual test implements
# its own wait. e2e tests should always pass once the non e2e have passed
$BIN test all --e2e --progress=false -v --junit-path test-results/results.xml $CONFIG_FILES
echo "::endgroup::"
