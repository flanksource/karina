#!/bin/bash

set -e

export IMAGES_FILE="/tmp/karina-images.txt"
export CLEANUP=${CLEANUP:=false}
export KIND_NODE_IMAGE="kindest/node"
export KIND_NODE_VERSION=${KIND_NODE_VERSION:-"v1.16.9"}
export KIND_DOCKER_IMAGE="$KIND_NODE_IMAGE:$KIND_NODE_VERSION"

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
    $KARINA provision kind-cluster -c test/minimal.yaml
fi

$KARINA images list -c test/minimal.yaml -o text >> $IMAGES_FILE
$KARINA images list -c test/monitoring.yaml -o text >> $IMAGES_FILE
$KARINA images list -c test/harbor.yaml -o text >> $IMAGES_FILE
$KARINA images list -c test/harbor2.yaml -o text >> $IMAGES_FILE
$KARINA images list -c test/postgres.yaml -o text >> $IMAGES_FILE
$KARINA images list -c test/elastic.yaml -o text >> $IMAGES_FILE
$KARINA images list -c test/security.yaml -o text >> $IMAGES_FILE
$KARINA images list -c test/platform.yaml -o text >> $IMAGES_FILE

cat $IMAGES_FILE | sort | uniq > $IMAGES_FILE

echo "pulling kind node image: $KIND_DOCKER_IMAGE"
docker pull $KIND_DOCKER_IMAGE

echo "creating directory for docker images: $DOCKER_IMAGE_DIRECTORY"
mkdir -p $DOCKER_IMAGE_DIRECTORY

echo "FROM $KIND_DOCKER_IMAGE" > $KIND_DOCKERFILE
echo "RUN mkdir -p /kind/images" >> $KIND_DOCKERFILE

cat $IMAGES_FILE | while read image || [[ -n $image ]];
do
    echo "pulling docker image: $image"
    docker pull $image

    IMAGE_NAME=$(echo $image | $SED 's#/#__#g' | $SED 's#:#_#g')
    echo "exporting image: $image to $DOCKER_IMAGE_DIRECTORY/$IMAGE_NAME.tgz"
    docker save $image > $DOCKER_IMAGE_DIRECTORY/$IMAGE_NAME.tgz
    echo "ADD images/$IMAGE_NAME.tgz /kind/images" >> $KIND_DOCKERFILE
done

pushd $DOCKER_DIRECTORY
    docker build -t $DOCKER_BUILD_IMAGE .
popd

if [[ "$CLEANUP" = true ]]; then
    echo "cleaning up"
    cat $IMAGES_FILE | while read image || [[ -n $image ]];
    do
        echo "removing image: $image"
        docker rmi $image
    done

    rm -rf $DOCKER_DIRECTORY
fi

echo "DONE !!"