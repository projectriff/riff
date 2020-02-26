#!/bin/bash

IMAGE_REPOSITORY_PREFIX="registry.kube-system.svc.cluster.local:5000"

fats_image_repo() {
  local function_name=$1

  echo -n "${IMAGE_REPOSITORY_PREFIX}/${function_name}/${CLUSTER_NAME}:latest"
}

fats_delete_image() {
  local image=$1

  # nothing to do
}

fats_create_push_credentials() {
  local namespace=$1

  # nothing to do
}
