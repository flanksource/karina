#!/bin/bash
: ${FETCH_REF:=1}
[ -z "$REFERENCE_VERSION" ] && echo -e "REFERENCE_VERSION not set! Try: \nexport REFERENCE_VERSION='v0.20.4'" && exit 1
BIN=./.bin/karina
mkdir -p .ref/
REF_BIN=.ref/karina
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
kind delete cluster --name upgrade-test || echo "No cluster present when starting"

if [ FETCH_REF ]; then
    wget -nv -O $REF_BIN https://github.com/flanksource/karina/releases/download/$REFERENCE_VERSION/karina
    wget -nv -O .ref/minimal.yaml  https://raw.githubusercontent.com/flanksource/karina/$REFERENCE_VERSION/test/minimal.yaml
    wget -nv -O .ref/$SUITE.yaml  https://raw.githubusercontent.com/flanksource/karina/$REFERENCE_VERSION/test/$SUITE.yaml
    chmod +x $REF_BIN
fi

export CONFIGURED_VALUE=`openssl rand -base64 12`
export PLATFORM_CONFIG=.ref/$SUITE.yaml

if [[ "$KUBECONFIG" != "$HOME/.kube/kind-config-kind" ]] ; then
  $REF_BIN ca generate --name root-ca --cert-path .certs/root-ca.crt --private-key-path .certs/root-ca.key --password foobar  --expiry 1
  $REF_BIN ca generate --name ingress-ca --cert-path .certs/ingress-ca.crt --private-key-path .certs/ingress-ca.key --password foobar  --expiry 1
  $REF_BIN ca generate --name sealed-secrets --cert-path .certs/sealed-secrets-crt.pem --private-key-path .certs/sealed-secrets-key.pem --password foobar  --expiry 1
  $REF_BIN provision kind-cluster -e name=upgrade-test --trace -vv|| exit 1
fi

$REF_BIN version

$REF_BIN deploy phases --crds --base --stubs --dex --calico --antrea --minio -v -e name=upgrade-test

[[ -e ./test/install_certs.sh ]] && ./test/install_certs.sh

# wait for the base deployment with stubs to come up healthy
$REF_BIN test phases --base --stubs --minio  --wait 120 --progress=false -e name=upgrade-test

# deploy the OPA bundle to the Minio instance for use by OPA
$REF_BIN opa deploy-bundle test/opa/bundles/automobile.tar.gz -e name=upgrade-test

$REF_BIN deploy all -v -e name=upgrade-test

# wait for up to 4 minutes, rerunning tests if they fail
# this allows for all resources to reconcile and images to finish downloading etc..
$REF_BIN test all -v --wait 300 --progress=false -e name=upgrade-test

failed=false

# E2E do not use --wait at the run level, if needed each individual test implements
# its own wait. e2e tests should always pass once the non e2e have passed
if ! $REF_BIN test all --e2e --progress=false -v --junit-path test-results/results.xml -e name=upgrade-test; then
  echo "Reference setup failed - Aborting upgrade test"
  failed=true
fi

if [[ "$failed" = false ]]; then
    echo "Upgrade Test Start"
    failed=false
    export PLATFORM_CONFIG=test/$SUITE.yaml

    $BIN version

    if ! $BIN deploy phases --pre --crds --base --stubs --dex --calico --antrea --minio -v -e name=upgrade-test; then
      failed=true
      echo "Failure in base deployment"
    fi

    [[ -e ./test/install_certs.sh ]] && ./test/install_certs.sh

    # wait for the base deployment with stubs to come up healthy
    if ! $BIN test phases --base --stubs --minio  --wait 120 --progress=false -e name=upgrade-test; then
      failed=true
      echo "Failure in base test"
    fi

    # deploy the OPA bundle to the Minio instance for use by OPA
    $BIN opa deploy-bundle test/opa/bundles/automobile.tar.gz -e name=upgrade-test

    if ! $BIN deploy all -v -e name=upgrade-test; then
      failed=true
      echo "Failure in feature deployment"
    fi

    # wait for up to 4 minutes, rerunning tests if they fail
    # this allows for all resources to reconcile and images to finish downloading etc..

    if ! $BIN test all -v --wait 300 --progress=false -e name=upgrade-test; then
      failed=true
      echo "Failure in feature test"
    fi

    # E2E do not use --wait at the run level, if needed each individual test implements
    # its own wait. e2e tests should always pass once the non e2e have passed
    if ! $BIN test all --e2e --progress=false -v --junit-path test-results/results.xml -e name=upgrade-test; then
      failed=true
      echo "Failure in e2e test"
    fi
fi

wget -nv https://github.com/flanksource/build-tools/releases/download/v0.11.3/build-tools
chmod +x build-tools
./build-tools gh actions report-junit test-results/results.xml --token $GIT_API_KEY --build "$BUILD"

mkdir -p artifacts
$BIN snapshot --output-dir snapshot -v --include-specs=true --include-logs=true --include-events=true -e name=upgrade-test
zip -r artifacts/snapshot.zip snapshot/*

if [[ "$failed" = true ]]; then
  exit 1
fi
