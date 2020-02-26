#!/bin/bash

# Login for local pushes
echo "$DOCKER_PASSWORD" | docker login --username $DOCKER_USERNAME --password-stdin

IMAGE_REPOSITORY_PREFIX="${DOCKER_USERNAME}"

fats_image_repo() {
  local function_name=$1

  echo -n "${IMAGE_REPOSITORY_PREFIX}/${function_name}-${CLUSTER_NAME}:latest"
}

fats_delete_image() {
  local image
  IFS=':' read -r -a image <<< "$1"
  local repo=${image[0]}
  local tag=${image[1]}

  echo "Delete image ${repo}:${tag}"
  local token=`curl -s -H "Content-Type: application/json" -X POST -d '{"username": "'${DOCKER_USERNAME}'", "password": "'${DOCKER_PASSWORD}'"}' https://hub.docker.com/v2/users/login/ | jq -r .token`
  curl "https://hub.docker.com/v2/repositories/${repo}/tags/${tag}/" -X DELETE -H "Authorization: JWT ${token}"
}

fats_create_push_credentials() {
  local namespace=$1

  echo "Create auth secret"
  echo -n "${DOCKER_PASSWORD}" | riff credentials apply dockerhub --docker-hub "${DOCKER_USERNAME}" --namespace "${namespace}"
}
