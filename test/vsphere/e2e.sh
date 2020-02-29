#!/bin/bash

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

# Binary
GO_VERSION=${GO_VERSION:-1.12}
SOPS_VERSION=${SOPS_VERSION:-3.5.0}

install_platformcli() {
  docker run --rm -it -v "$PWD":"$PWD" -v "$PWD"/go:"$PWD"/go \
    -u "$(id -u)":"$(id -g)" \
    -w "$PWD" --entrypoint make \
    -e GOPROXY=https://proxy.golang.org \
    -e XDG_CACHE_HOME=/tmp/.cache \
    golang:"$GO_VERSION" pack build

  sudo mv "$PWD"/.bin/platform-cli /usr/local/bin/platform-cli
  sudo chmod +x /usr/local/bin/platform-cli
  sudo apt-get update
  sudo apt-get install -y genisoimage
}

install_sops() {
  curl -L https://github.com/mozilla/sops/releases/download/v"$SOPS_VERSION"/sops-v"$SOPS_VERSION".linux -o /tmp/sops
  sudo mv /tmp/sops /usr/local/bin/sops
  sudo chmod +x /usr/local/bin/sops
}

decrypt_vpn_certs() {
  export AWS_ACCESS_KEY_ID=$SOPS_AWS_ACCESS_KEY_ID
  export AWS_SECRET_ACCESS_KEY=$SOPS_AWS_SECRET_ACCESS_KEY

  for filename in "$CERTS"/*.enc; do
    # save the decrypted certs in the same folder w/o the ext .enc
    sops --config "$PWD"/test/vsphere/.sops.yaml -d \
      --input-type binary --output-type binary "$filename" >"${filename%.*}"
  done

  # clean a WARNING in openvpn
  # WARNING: file '/openvpn/certs/user.key' is group or others accessible
  chmod 400 "$CERTS"/user.key
}

dump_logs() {
  mkdir -p "$ARTIFACTS"/snapshot
  platform-cli snapshot --output-dir "$ARTIFACTS"/snapshot || echo "Failed to take cluster snapshot."
}

on_exit() {
  # dump the logs into the ARTIFACTS directory
  dump_logs

  # remove the cluster
  # shellcheck disable=SC2086
  platform-cli cleanup $PLATFORM_OPTIONS_FLAGS

  # kill the VPN
  docker kill vpn

  # clean certs dir - remove all except enc
  rm "$CERTS"/{*.crt,*.key}
}

wait_for_vpn() {
  while [ -n "$(ip addr show tun0 2>&1 >/dev/null)" ]; do
    sleep 0.1
  done
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

# These variables are required to configure access to vSphere, they also work with govc
export GOVC_FQDN="vcenter.lab.flanksource.com"
export GOVC_DATACENTER="lab"
export GOVC_CLUSTER="cluster"
export GOVC_NETWORK="VM Network"
export GOVC_PASS="${GOVC_PASS}"
export GOVC_USER="${GOVC_USER}"
export GOVC_DATASTORE="datastore1"
export GOVC_INSECURE=1
export GOVC_URL="$GOVC_USER:$GOVC_PASS@$GOVC_FQDN"

trap on_exit EXIT

install_platformcli

install_sops

# Decrypt certs encrypted with SOPS and KMS Key from https://github.com/flanksource/vsphere-lab/blob/master/.sops.yaml
decrypt_vpn_certs

# Run the vpn client in container
docker run --rm -d --name vpn -v "${PWD}/test/vsphere/openvpn/:/openvpn/" \
  -w "/openvpn/" --cap-add=NET_ADMIN --net=host --device=/dev/net/tun \
  gcr.io/cluster-api-provider-vsphere/extra/openvpn:latest

echo "nameserver 10.255.0.2" | sudo tee -a /run/resolvconf/resolv.conf >/dev/null

## Wait for VPN
wait_for_vpn

# Tail the vpn logs
docker logs vpn

# Generate Ingress certs
platform-cli ca generate --name ingress-ca \
  --cert-path "$PLATFORM_CA" --private-key-path "$PLATFORM_PRIVATE_KEY" \
  --password "$PLATFORM_CA_CEK" --expiry 1

# Create the cluster using the config from PLATFORM_CONFIG
# shellcheck disable=SC2086
platform-cli provision vsphere-cluster $PLATFORM_OPTIONS_FLAGS

# Install CNI
# shellcheck disable=SC2086
platform-cli deploy calico $PLATFORM_OPTIONS_FLAGS

# Build the base platform configuration
# shellcheck disable=SC2086

# wait all nodes to be up
# FIXME: For some weird reasons the command above does not complete
# like if it was async and the job get stuck here
# kubectl --kubeconfig "${PWD}/${PLATFORM_CLUSTER_ID}-admin.yml" wait --for=condition=Ready --all --timeout=1h nodes

# Deploy the platform configuration
# shellcheck disable=SC2086
platform-cli deploy all $PLATFORM_OPTIONS_FLAGS

# Run conformance tests
# shellcheck disable=SC2086
platform-cli test $PLATFORM_OPTIONS_FLAGS
