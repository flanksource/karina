#!/bin/bash

#TODO: verify licensing origin then remove
# Copyright 2020 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit  # exits immediately on any unexpected error (does not bypass traps)
set -o nounset  # will error if variables are used without first being defined
set -o pipefail # any non-zero exit code in a piped command causes the pipeline to fail with that code

ARTIFACTS=$PWD/artifacts
CERTS=$PWD/test/vsphere/openvpn/certs


dump_logs() {
  mkdir -p "$ARTIFACTS"/snapshot
  platform-cli snapshot --output-dir "$ARTIFACTS"/snapshot || echo "Failed to take cluster snapshot."
}

on_exit() {

  # remove the cluster
  platform-cli cleanup $PLATFORM_OPTIONS_FLAGS

  # clean certs dir - remove all except enc
  rm -fv "$CERTS"/{*.crt,*.key}
}

generate_cluster_id() {
  local prefix

  prefix=$(tr </dev/urandom -cd 'a-f0-9' | head -c 5)
  echo "e2e-${prefix}"
}

# These are used inside e2e.yml
PLATFORM_CLUSTER_ID=$(generate_cluster_id)
export PLATFORM_CLUSTER_ID
export PLATFORM_CA=$PWD/.certs/ca.crt.pem
export PLATFORM_PRIVATE_KEY=$PWD/.certs/ca.key.pem
export PLATFORM_CA_CEK=foobar
export PLATFORM_OPTIONS_FLAGS="-e name=${PLATFORM_CLUSTER_ID} -e domain=${PLATFORM_CLUSTER_ID}.lab.flanksource.com -vv"

export PLATFORM_CONFIG=$PWD/test/vsphere/e2e.yml

## These variables are required to configure access to vSphere, they also work with govc
#export GOVC_FQDN="vcenter.lab.flanksource.com"
#export GOVC_DATACENTER="lab"
#export GOVC_CLUSTER="cluster"
#export GOVC_NETWORK="VM Network"
#export GOVC_PASS="${GOVC_PASS}"
#export GOVC_USER="${GOVC_USER}"
#export GOVC_DATASTORE="datastore1"
#export GOVC_INSECURE=1
export GOVC_URL="$GOVC_USER:$GOVC_PASS@$GOVC_FQDN"

trap on_exit EXIT

BIN=./.bin/platform-cli
mkdir -p .bin
export PLATFORM_CONFIG=test/common.yaml
export GO_VERSION=${GO_VERSION:-1.13}
export KUBECONFIG=~/.kube/config

go version

if go version | grep  go$GO_VERSION; then
  make pack build
else
  docker run --rm -it -v $PWD:$PWD -v /go:/go -w $PWD --entrypoint make -e GOPROXY=https://proxy.golang.org golang:$GO_VERSION pack build
fi

$BIN version

# Generate ingress ca
$BIN ca generate --name ingress-ca \
  --cert-path "$PLATFORM_CA" --private-key-path "$PLATFORM_PRIVATE_KEY" \
  --password "$PLATFORM_CA_CEK" --expiry 1


# Create the cluster using the config from PLATFORM_CONFIG
$BIN  provision vsphere-cluster $PLATFORM_OPTIONS_FLAGS

# Install CNI
$BIN  deploy calico $PLATFORM_OPTIONS_FLAGS

components=("base" "stubs" "all")
for component in "${components[@]}"; do
  # Deploy the platform configuration
  $BIN deploy "$component" $PLATFORM_OPTIONS_FLAGS
done

set +o errexit # test failures are reported by Junit

# Run conformance tests
failed=false
if ! $BIN test all $PLATFORM_OPTIONS_FLAGS --wait 240  --junit-path test-results/results.xml; then
   failed=true
fi

# dump the logs into the ARTIFACTS directory
mkdir -p artifacts
platform-cli snapshot $PLATFORM_OPTIONS_FLAGS --output-dir snapshot -v --include-specs=true --include-logs=true --include-events=true 
zip -r artifacts/snapshot.zip snapshot/*

if [[ "$failed" = true ]]; then
  exit 1
fi
