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

export CONFIGURED_VALUE=$(openssl rand -base64 12)
export PLATFORM_CONFIG=test/$SUITE.yaml
echo "::endgroup::"

if [[ "$KUBECONFIG" != "$HOME/.kube/kind-config-kind" ]] ; then
  echo "::group::Provisioning"
  $BIN ca generate --name root-ca --cert-path .certs/root-ca.crt --private-key-path .certs/root-ca.key --password foobar  --expiry 1
  $BIN ca generate --name ingress-ca --cert-path .certs/ingress-ca.crt --private-key-path .certs/ingress-ca.key --password foobar  --expiry 1
  $BIN ca generate --name sealed-secrets --cert-path .certs/sealed-secrets-crt.pem --private-key-path .certs/sealed-secrets-key.pem --password foobar  --expiry 1
  if $BIN provision kind-cluster --trace -vv ; then
    echo "::endgroup::"
  else
    echo "::endgroup::"
    exit 1
  fi
fi

$BIN version

 echo "::group::Deploying Base"
$BIN deploy phases --crds --base --stubs --dex --calico --antrea --minio -v

[[ -e ./test/install_certs.sh ]] && ./test/install_certs.sh
echo "::endgroup::"

echo "::group::Waiting for Base"
# wait for the base deployment with stubs to come up healthy
$BIN test phases --base --stubs --minio  --wait 120 --progress=false
echo "::endgroup::"

echo "::group::Deploy All"
# deploy the OPA bundle to the Minio instance for use by OPA
$BIN opa deploy-bundle test/opa/bundles/automobile.tar.gz

$BIN deploy all -v
echo "::endgroup::"

echo "::group::Test Dry Run"

# wait for up to 4 minutes, rerunning tests if they fail
# this allows for all resources to reconcile and images to finish downloading etc..
$BIN test all -v --wait 300 --progress=false

failed=false
echo "::endgroup::"

echo "::group::Final Test Run"
# E2E do not use --wait at the run level, if needed each individual test implements
# its own wait. e2e tests should always pass once the non e2e have passed
if ! $BIN test all --e2e --progress=false -v --junit-path test-results/results.xml; then
  failed=true
fi
echo "::endgroup::"

echo "::group::Uploading Results"
wget -nv -nc -O build-tools \
  https://github.com/flanksource/build-tools/releases/latest/download/build-tools && \
  chmod +x build-tools

./build-tools junit gh-workflow-commands test-results/results.xml

TESULTS_TOKEN=$(cat test/tesults.yaml | jq -r .\"$KUBERNETES_VERSION-$SUITE\")
if [[ $TESULTS_TOKEN != "" ]]; then
  ./build-tools junit upload-tesults test-results/results.xml --token $TESULTS_TOKEN
fi

mkdir -p artifacts
$BIN snapshot --output-dir snapshot -v --include-specs=true --include-logs=true --include-events=true
zip -r artifacts/snapshot.zip snapshot/*
echo "::endgroup::"
if [[ "$failed" = true ]]; then
  exit 1
fi
