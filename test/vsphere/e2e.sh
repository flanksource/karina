#!/bin/bash

mkdir -p .bin
mkdir -p .certs

export TERM=xterm-256color
REPO=$(basename $(git remote get-url origin | sed 's/\.git//'))
BIN=.bin/karina
chmod +x $BIN

export PLATFORM_CLUSTER_ID=vsphere-e2e
export PLATFORM_OPTIONS_FLAGS="-e name=${PLATFORM_CLUSTER_ID} -e domain=${PLATFORM_CLUSTER_ID}.lab.flanksource.com -v"
export PLATFORM_CONFIG=${PLATFORM_CONFIG:-test/vsphere/vsphere.yaml}
unset KUBECONFIG

# delete any dangling VMs before starting
if ! which govc; then
     curl -s -L --output govc.gz https://github.com/vmware/govmomi/releases/download/v0.23.0/govc_linux_amd64.gz
     gunzip govc.gz
     chmod +x govc
     ./govc ls /dc1/vm | grep vsphere-e2e | xargs -I {} ./govc vm.destroy {}
else
  govc ls /dc1/vm | grep vsphere-e2e | xargs -I {} govc vm.destroy {}
fi

printf "\n\n\n\n$(tput bold)Generate Certs$(tput setaf 7)\n"
$BIN ca generate --name root-ca --cert-path .certs/root-ca.crt --private-key-path .certs/root-ca.key --password foobar  --expiry 1
$BIN ca generate --name ingress-ca --cert-path .certs/ingress-ca.crt --private-key-path .certs/ingress-ca.key --password foobar  --expiry 1
$BIN ca generate --name sealed-secrets --cert-path .certs/sealed-secrets-crt.pem --private-key-path .certs/sealed-secrets-key.pem --password foobar  --expiry 1

printf "\n\n\n\n$(tput bold)Provision Cluster$(tput setaf 7)\n"
$BIN provision vsphere-cluster $PLATFORM_OPTIONS_FLAGS

printf "\n\n\n\n$(tput bold)Basic Deployments$(tput setaf 7)\n"

$BIN deploy phases --calico --base --stubs --dex $PLATFORM_OPTIONS_FLAGS

printf "\n\n\n\n$(tput bold)Up?$(tput setaf 7)\n"
# wait for the base deployment with stubs to come up healthy
$BIN test phases --base --stubs --wait 120 --progress=false $PLATFORM_OPTIONS_FLAGS

printf "\n\n\n\n$(tput bold)All Deployments$(tput setaf 7)\n"
$BIN deploy all $PLATFORM_OPTIONS_FLAGS

failed=false

## e2e do not use --wait at the run level, if needed each individual test implements
## its own wait. e2e tests should always pass once the non e2e have passed
if ! $BIN test  all --e2e --progress=false --junit-path test-results/results.xml $PLATFORM_OPTIONS_FLAGS; then
  failed=true
fi

printf "\n\n\n\n$(tput bold)Reporting$(tput setaf 7)\n"
build-tools gh actions report-junit test-results/results.xml --token $GIT_API_KEY --build "$BUILD"

mkdir -p artifacts
$BIN snapshot --output-dir snapshot -v --include-specs=true --include-logs=true --include-events=true $PLATFORM_OPTIONS_FLAGS
zip -r artifacts/snapshot.zip snapshot/*

$BIN terminate-orphans $PLATFORM_OPTIONS_FLAGS || echo "Orphans not terminated."
$BIN cleanup $PLATFORM_OPTIONS_FLAGS

# serves as a workaround for `sops exec-env` not passing exit codes
# in current release: https://github.com/mozilla/sops/issues/626
build-tools junit passfail test-results/results.xml || echo "TEST_FAILURE"

if [[ "$failed" == true ]]; then
  echo "TEST_FAILURE"
  exit 1
fi
