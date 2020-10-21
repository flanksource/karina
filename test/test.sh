#!/bin/bash
BIN=./.bin/karina
chmod +x $BIN
mkdir -p .bin
export GO_VERSION=${GO_VERSION:-1.13}
export KUBECONFIG=~/.kube/config
REPO=$(basename $(git remote get-url origin | sed 's/\.git//'))
GITHUB_OWNER=$(basename $(dirname $(git remote get-url origin | sed 's/\.git//')))
GITHUB_OWNER=${GITHUB_OWNER##*:}
MASTER_HEAD=$(curl -s https://api.github.com/repos/$GITHUB_OWNER/$REPO/commits/master | jq -r '.sha')

[ -z "$KUBERNETES_VERSION" ] && echo -e "KUBERNETES_VERSION not set! Try: \nexport KUBERNETES_VERSION='v1.16.9'" && exit 1
[ -z "$SUITE" ] && echo -e "SUITE not set! Try: \n export SUITE='minimal'\n or one of these (minus extension):"&& ls  test/*.yaml && exit 1
kind delete cluster --name kind-$SUITE-$KUBERNETES_VERSION || echo "No cluster present when starting"

export CONFIGURED_VALUE=`openssl rand -base64 12`
export PLATFORM_CONFIG=test/$SUITE.yaml

if [[ "$KUBECONFIG" != "$HOME/.kube/kind-config-kind" ]] ; then
  $BIN ca generate --name root-ca --cert-path .certs/root-ca.crt --private-key-path .certs/root-ca.key --password foobar  --expiry 1
  $BIN ca generate --name ingress-ca --cert-path .certs/ingress-ca.crt --private-key-path .certs/ingress-ca.key --password foobar  --expiry 1
  $BIN ca generate --name sealed-secrets --cert-path .certs/sealed-secrets-crt.pem --private-key-path .certs/sealed-secrets-key.pem --password foobar  --expiry 1
  $BIN provision kind-cluster --trace -vv|| exit 1
fi

$BIN version

$BIN deploy phases --crds --base --stubs --dex --calico --antrea --minio -v

[[ -e ./test/install_certs.sh ]] && ./test/install_certs.sh

# wait for the base deployment with stubs to come up healthy
$BIN test phases --base --stubs --minio  --wait 120 --progress=false

# deploy the OPA bundle to the Minio instance for use by OPA
$BIN opa deploy-bundle test/opa/bundles/automobile.tar.gz

$BIN deploy all -v

# wait for up to 4 minutes, rerunning tests if they fail
# this allows for all resources to reconcile and images to finish downloading etc..
$BIN test all -v --wait 300 --progress=false

failed=false

# E2E do not use --wait at the run level, if needed each individual test implements
# its own wait. e2e tests should always pass once the non e2e have passed
if ! $BIN test all --e2e --progress=false -v --junit-path test-results/results.xml; then
  failed=true
fi

wget -nv https://github.com/flanksource/build-tools/releases/download/v0.9.9/build-tools
chmod +x build-tools
./build-tools gh actions report-junit test-results/results.xml --token $GIT_API_KEY --build "$BUILD"

mkdir -p artifacts
$BIN snapshot --output-dir snapshot -v --include-specs=true --include-logs=true --include-events=true
zip -r artifacts/snapshot.zip snapshot/*

if [[ "$failed" = true ]]; then
  exit 1
fi
