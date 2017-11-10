#!/bin/bash

set -eu

print_help() {
  echo
  echo "  Usage:"
  echo "  SIDECAR_IMAGE_TAG=<tag> HTTP_GATEWAY_IMAGE_TAG=<tag> SK8S_IMAGE_TAG=<tag> $0 <sk8s-version>"
  echo
}

if [ "0" == "$#" ]; then
  print_help
  exit 0

elif [ "$1" == "-h" ]; then
  print_help
  exit 0

else
  export SK8S_VERSION="$1"
fi



dir=$(dirname $0)

export SIDECAR_IMAGE_TAG=${SIDECAR_IMAGE_TAG:-dev}
export HTTP_GATEWAY_IMAGE_TAG=${HTTP_GATEWAY_IMAGE_TAG:-dev}
export SK8S_IMAGE_TAG=${SK8S_IMAGE_TAG:-dev}

for template in $(ls "$dir/templates_erb"); do
  erb "$dir/templates_erb/$template" > "$dir/sk8s/templates/$template"
done

helm package sk8s --version "$SK8S_VERSION"
