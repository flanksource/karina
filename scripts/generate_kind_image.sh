#!/bin/bash

set -e

export IMAGES_FILE="/tmp/karina-images.txt"
export CLEANUP=${CLEANUP:=false}
export DOCKER_PUSH=${DOCKER_PUSH:=false}
export KIND_NODE_IMAGE="kindest/node"
export KIND_NODE_VERSION=${KIND_NODE_VERSION:-"v1.16.9"}
export KIND_DOCKER_IMAGE="$KIND_NODE_IMAGE:$KIND_NODE_VERSION"
export KUBECONFIG=~/.kube/config

export DOCKER_DIRECTORY="docker-images/build"
export DOCKER_IMAGE_DIRECTORY="docker-images/build/images"
export KIND_DOCKERFILE="$DOCKER_DIRECTORY/Dockerfile"

if [ "$(uname)" == "Darwin" ]; then
    export SED="gsed"
elif [ "$(expr substr $(uname -s) 1 5)" == "Linux" ]; then
    export SED="sed"
fi

export TAG_DATE=$(date +%Y%m%d%M%H%M%S)
export DOCKER_TAG=${DOCKER_TAG:-$KIND_NODE_VERSION-$TAG_DATE}
export DOCKER_IMAGE=${DOCKER_IMAGE:-flanksource/kind-node}
export DOCKER_BUILD_IMAGE="$DOCKER_IMAGE:$DOCKER_TAG"
export KARINA=${KARINA:-./.bin/karina}

chmod +x $KARINA

if [[ "$DEPLOY_KIND_CLUSTER" = true ]]; then
    $KARINA ca generate --name root-ca --cert-path .certs/root-ca.crt --private-key-path .certs/root-ca.key --password foobar  --expiry 1
    $KARINA ca generate --name ingress-ca --cert-path .certs/ingress-ca.crt --private-key-path .certs/ingress-ca.key --password foobar  --expiry 1
    $KARINA ca generate --name sealed-secrets --cert-path .certs/sealed-secrets-crt.pem --private-key-path .certs/sealed-secrets-key.pem --password foobar  --expiry 1
    $KARINA provision kind-cluster -c test/minimal.yaml
fi

$KARINA images list -c test/minimal.yaml  >> $IMAGES_FILE
$KARINA images list -c test/monitoring.yaml  >> $IMAGES_FILE
$KARINA images list -c test/harbor2.yaml >> $IMAGES_FILE
$KARINA images list -c test/postgres.yaml  >> $IMAGES_FILE
$KARINA images list -c test/elastic.yaml  >> $IMAGES_FILE
$KARINA images list -c test/security.yaml  >> $IMAGES_FILE
$KARINA images list -c test/platform.yaml  >> $IMAGES_FILE
cat $IMAGES_FILE | sort | uniq | sponge  $IMAGES_FILE
echo "pulling kind node image: $KIND_DOCKER_IMAGE"
docker pull $KIND_DOCKER_IMAGE

echo "creating directory for docker images: $DOCKER_IMAGE_DIRECTORY"
mkdir -p $DOCKER_IMAGE_DIRECTORY

echo "FROM $KIND_DOCKER_IMAGE" > $KIND_DOCKERFILE
echo "RUN mkdir -p /kind/images" >> $KIND_DOCKERFILE

cat $IMAGES_FILE | while read image || [[ -n $image ]];
do
    IMAGE_NAME=$(echo $image | $SED 's#/#__#g' | $SED 's#:#_#g')

    if [[ ! -e  $DOCKER_IMAGE_DIRECTORY/$IMAGE_NAME.tgz ]]; then
        echo "pulling docker image: $image"
        docker pull $image
        echo "exporting image: $image to $DOCKER_IMAGE_DIRECTORY/$IMAGE_NAME.tgz"
        docker save $image > $DOCKER_IMAGE_DIRECTORY/$IMAGE_NAME.tgz
    fi
    echo "ADD images/$IMAGE_NAME.tgz /kind/images" >> $KIND_DOCKERFILE

    if [[ "$CLEANUP" = true ]]; then
        echo "removing image: $image"
        docker rmi $image
    fi
done

pushd $DOCKER_DIRECTORY
    docker build -t $DOCKER_BUILD_IMAGE .
popd

if [[ "$CLEANUP" == true ]]; then
    rm -rf $DOCKER_DIRECTORY
fi

if [[ "$DOCKER_PUSH" == true ]]; then
    echo $DOCKER_PASS | docker login -u $DOCKER_USER --password-stdin
    docker push $DOCKER_BUILD_IMAGE
fi

echo "DONE !!"
