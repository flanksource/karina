#!/bin/bash
: ${FETCH_REF:=1}
mkdir -p .bin .ref
REFERENCE_VERSION=${REFERENCE_VERSION:-v0.24.1}
KUBERNETES_VERSION=${KUBERNETES_VERSION:-v1.18.6}
SUITE=${SUITE:-minimal}
if test -f ./.bin/karina; then
    BIN=./.bin/karina
    chmod +x $BIN
elif command -v karina; then
    BIN=$(command -v karina)
else
    echo "No karina binary detected"
    exit 127
fi

REF_BIN=.ref/karina

export KUBECONFIG=~/.kube/config
REPO=$(basename $(git remote get-url origin | sed 's/\.git//'))
GITHUB_OWNER=$(basename $(dirname $(git remote get-url origin | sed 's/\.git//')))
GITHUB_OWNER=${GITHUB_OWNER##*:}
MASTER_HEAD=$(curl -s https://api.github.com/repos/$GITHUB_OWNER/$REPO/commits/master | jq -r '.sha')

kind delete cluster --name upgrade-test || echo "No cluster present when starting"

ARCH=
if [[ "$OSTYPE" == "darwin"* ]]; then
  ARCH=_osx
fi
if [[ ! -e $REF_BIN ]]; then
    wget -nv -O $REF_BIN https://github.com/flanksource/karina/releases/download/$REFERENCE_VERSION/karina$ARCH
    wget -nv -O .ref/minimal.yaml  https://raw.githubusercontent.com/flanksource/karina/$REFERENCE_VERSION/test/minimal.yaml
    wget -nv -O .ref/$SUITE.yaml  https://raw.githubusercontent.com/flanksource/karina/$REFERENCE_VERSION/test/$SUITE.yaml
    chmod +x $REF_BIN
fi

export CONFIGURED_VALUE=`openssl rand -base64 12`
export PLATFORM_CONFIG=.ref/$SUITE.yaml

echo "::group::Setting up reference cluster with karina $REFERENCE_VERSION"

if [[ "$KUBECONFIG" != "$HOME/.kube/kind-config-kind" ]] ; then
  $REF_BIN ca generate --name root-ca --cert-path .certs/root-ca.crt --private-key-path .certs/root-ca.key --password foobar  --expiry 1
  $REF_BIN ca generate --name ingress-ca --cert-path .certs/ingress-ca.crt --private-key-path .certs/ingress-ca.key --password foobar  --expiry 1
  $REF_BIN ca generate --name sealed-secrets --cert-path .certs/sealed-secrets-crt.pem --private-key-path .certs/sealed-secrets-key.pem --password foobar  --expiry 1
  $REF_BIN provision kind-cluster -e name=upgrade-test --trace -vv -e kubernetes.version=$KUBERNETES_VERSION || exit 1
fi

$REF_BIN version

$REF_BIN deploy phases --crds --base --stubs --dex --calico --antrea --minio -v -e name=upgrade-test

[[ -e ./test/install_certs.sh ]] && ./test/install_certs.sh

# wait for the base deployment with stubs to come up healthy
$REF_BIN test phases --base --stubs --minio  --wait 120 --progress=false -e name=upgrade-test --fail-on-error=false

$REF_BIN deploy all -v -e name=upgrade-test ||  (echo "::error::Error while deploying reference version" && exit 1)

$REF_BIN test all --e2e --progress=false -v --wait 300 -e name=upgrade-test --fail-on-error=false

function report() {
  set +e
  echo "::group::Uploading test results"
  if [[ "$CI" == "true" ]]; then
    wget -nv -nc -O build-tools \
      https://github.com/flanksource/build-tools/releases/latest/download/build-tools && \
      chmod +x build-tools

    ./build-tools junit gh-workflow-commands test-results/results.xml

    mkdir -p artifacts
    $BIN snapshot --output-dir snapshot -v --include-specs=true --include-logs=true --include-events=true -e name=upgrade-test || echo "::error::Error while creating snapshot"
    zip -r artifacts/snapshot.zip snapshot/* || echo "::error::Error while zipping snapshot"
  else
    echo "Skipping test report when not running in CI"
  fi
  echo "::endgroup::"
}
trap report EXIT

set -e

echo "::endgroup::"

echo "::group::Initial upgrade"
export PLATFORM_CONFIG=test/$SUITE.yaml
$BIN version
$BIN deploy phases --bootstrap --stubs --prune=false  -v -e name=upgrade-test || (echo "::error::Error while upgrading" && exit 1)
echo "::endgroup::"

echo "::group::Waiting for cluster to be health"
# wait for up to 5 minutes, rerunning tests if they fail
# this allows for all resources to reconcile and images to finish downloading etc..
$BIN test all -v --wait 300 --progress=false -e name=upgrade-test --fail-on-error=false
echo "::endgroup::"


echo "::group::Final Tests"
$BIN deploy all --prune=false -v -e name=upgrade-test ||  (echo "::error::Error while performing full upgrade" && exit 1)
$BIN test all --e2e --progress=false -v --junit-path test-results/results.xml -e name=upgrade-test
echo "::endgroup::"
