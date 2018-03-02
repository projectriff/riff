#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

version=`cat VERSION`
images=(
    "function-controller"
    "function-sidecar"
    "http-gateway"
    "topic-controller"
)

docker login -u "$DOCKER_USERNAME" -p "$DOCKER_PASSWORD"

make dockerize

for image in "${images[@]}"
do
	echo "Publishing ${image}"

    docker tag "projectriff/${image}:${version}" "projectriff/${image}:latest"
    docker tag "projectriff/${image}:${version}" "projectriff/${image}:${version}-ci-${TRAVIS_COMMIT}"

    docker push "projectriff/${image}"
done
