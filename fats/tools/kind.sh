#!/bin/bash

kind_version="${1:-v0.7.0}"
base_url="${2:-https://github.com/kubernetes-sigs/kind/releases/download}"

if [ "$machine" == "MinGw" ]; then
  curl -Lo kind.exe ${base_url}/${kind_version}/kind-windows-amd64
  mv kind.exe /usr/bin/
else
  curl -Lo kind ${base_url}/${kind_version}/kind-linux-amd64
  chmod +x kind
  sudo mv kind /usr/local/bin/
fi
