#!/bin/bash

# Install gcloud for GCR access
`dirname "${BASH_SOURCE[0]}"`/../../install.sh gcloud

if [ "$machine" == "MinGw" ]; then
  # `gcloud auth configure-docker` doesn't work on Windows for some reason
  echo "${GCLOUD_CLIENT_SECRET}" | base64 --decode | docker login -u _json_key --password-stdin https://gcr.io
else
  gcloud auth configure-docker
fi

IMAGE_REPOSITORY_PREFIX="gcr.io/`gcloud config get-value project`"

fats_image_repo() {
  local function_name=$1

  echo -n "${IMAGE_REPOSITORY_PREFIX}/${function_name}/${CLUSTER_NAME}:latest"
}

fats_delete_image() {
  local image=$1

  # drop the tag if there is also a digest, preserving the digest
  image=$(echo $image | sed -e 's|:[^@:]*@|@|g')

  gcloud container images delete $image --force-delete-tags
}

fats_create_push_credentials() {
  local namespace=$1

  echo "Create auth secret"
  echo $GCLOUD_CLIENT_SECRET | base64 --decode > key.json
  riff credentials apply gcr --gcr key.json --namespace "${namespace}"
  rm key.json
}
