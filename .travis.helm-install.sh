#!/bin/bash

set -o errexit
set -o pipefail

if [[ `which helm` == "" ]]; then
  curl https://storage.googleapis.com/kubernetes-helm/helm-v2.8.1-linux-amd64.tar.gz | tar xz
  chmod +x linux-amd64/helm
  sudo mv linux-amd64/helm /usr/local/bin/
  rm -rf linux-amd64
fi
