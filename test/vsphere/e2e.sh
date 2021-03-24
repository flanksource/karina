#!/bin/bash

echo "::group::Setup"
source /dev/stdin < <( sops -d --input-type binary --output-type binary ./test/vsphere/e2e.sh )
mkdir -p .bin
mkdir -p .certs

export TERM=xterm-256color
REPO=$(basename $(git remote get-url origin | sed 's/\.git//'))
BIN=.bin/karina
chmod +x $BIN

generate_cluster_id() {
  echo e2e-$(date "+%d%H%M")
}

export PLATFORM_CLUSTER_ID=$(generate_cluster_id)
export PLATFORM_OPTIONS_FLAGS="-e name=${PLATFORM_CLUSTER_ID} -e domain=${PLATFORM_CLUSTER_ID}.lab.flanksource.com -v"
export PLATFORM_CONFIG=${PLATFORM_CONFIG:-test/vsphere/vsphere.yaml}
unset KUBECONFIG

mkdir -p ~/.ssh
chmod 700 ~/.ssh
echo "$SSH_SECRET_KEY_BASE64" | base64 -d > ~/.ssh/id_rsa
chmod 600 ~/.ssh/id_rsa

sshuttle --dns -e "ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no" -r $SSH_USER@$SSH_JUMP_HOST $VPN_NETWORK &
# Wait for connection
sleep 2s
echo "::endgroup::"

echo "::group::Provisioning"
$BIN ca generate --name root-ca --cert-path .certs/root-ca.crt --private-key-path .certs/root-ca.key --password foobar  --expiry 1
$BIN ca generate --name ingress-ca --cert-path .certs/ingress-ca.crt --private-key-path .certs/ingress-ca.key --password foobar  --expiry 1
$BIN ca generate --name sealed-secrets --cert-path .certs/sealed-secrets-crt.pem --private-key-path .certs/sealed-secrets-key.pem --password foobar  --expiry 1

$BIN provision vsphere-cluster $PLATFORM_OPTIONS_FLAGS
echo "::endgroup::"

echo "::group::Deploying Base"
$BIN deploy phases --bootstrap --stubs   $PLATFORM_OPTIONS_FLAGS
echo "::endgroup::"

echo "::group::Waiting for Base"
# wait for the base deployment with stubs to come up healthy
$BIN test phases  --bootstrap --stubs   --wait 120 --progress=false $PLATFORM_OPTIONS_FLAGS
echo "::endgroup::"

echo "::group::Deploy All"
$BIN deploy all $PLATFORM_OPTIONS_FLAGS
echo "::endgroup::"

echo "::group::Test Dry Run"
$BIN test all -v --wait 300 --progress=false $PLATFORM_OPTIONS_FLAGS
failed=false
echo "::endgroup::"

echo "::group::Final Test Run"
## e2e do not use --wait at the run level, if needed each individual test implements
## its own wait. e2e tests should always pass once the non e2e have passed
if ! $BIN test  all --e2e --progress=false --junit-path test-results/results.xml $PLATFORM_OPTIONS_FLAGS; then
  failed=true
  echo "Failure in feature test"
fi
echo "::endgroup::"

echo "::group::Uploading Results"
wget -nv -nc -O build-tools \
  https://github.com/flanksource/build-tools/releases/latest/download/build-tools && \
  chmod +x build-tools

./build-tools junit gh-workflow-commands test-results/results.xml

TESULTS_TOKEN=$(cat test/tesults.yaml | jq -r .\"$KUBERNETES_VERSION-$SUITE-vsphere\")
if [[ $TESULTS_TOKEN != "" ]]; then
  ./build-tools junit upload-tesults test-results/results.xml --token $TESULTS_TOKEN
fi

mkdir -p artifacts
$BIN snapshot --output-dir snapshot -v --include-specs=true --include-logs=true --include-events=true $PLATFORM_OPTIONS_FLAGS
zip -r artifacts/snapshot.zip snapshot/*
echo "::endgroup::"

echo "::group::Clean up"
$BIN terminate-orphans $PLATFORM_OPTIONS_FLAGS || echo "Orphans not terminated."
$BIN cleanup $PLATFORM_OPTIONS_FLAGS
echo "::endgroup::"

if [[ "$failed" = true ]]; then
  echo "Test failed."
  exit 1
fi
echo "Test passed!"
