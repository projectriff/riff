#!/bin/bash

kubectl_version="${1:-v1.15.3}"
base_url="${2:-https://storage.googleapis.com/kubernetes-release/release}"

if [ "$machine" == "MinGw" ]; then
  curl -Lo kubectl.exe ${base_url}/${kubectl_version}/bin/windows/amd64/kubectl.exe
  mv kubectl.exe /usr/bin/
else
  curl -Lo kubectl ${base_url}/${kubectl_version}/bin/linux/amd64/kubectl
  chmod +x kubectl
  sudo mv kubectl /usr/local/bin/
fi
