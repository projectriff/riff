#!/bin/bash

kustomize_version="${1:-v3.3.0}"
base_url="${2:-https://github.com/kubernetes-sigs/kustomize/releases}"

if [ "$machine" == "MinGw" ]; then
  kustomize_dir=`mkdir -p kustomize.XXXX`

  curl -L ${base_url}/download/kustomize/${kustomize_version}/kustomize_${kustomize_version}_windows_amd64.tar.gz | tar xz -C $kustomize_dir
  mv $kustomize_dir/kustomize.exe /usr/bin/

  rm -rf $kustomize_dir
else
  kustomize_dir=`mktemp -d kustomize.XXXX`

  curl -L ${base_url}/download/kustomize/${kustomize_version}/kustomize_${kustomize_version}_linux_amd64.tar.gz | tar xz -C $kustomize_dir
  chmod +x $kustomize_dir/kustomize
  sudo mv $kustomize_dir/kustomize /usr/local/bin/

  rm -rf $kustomize_dir
fi
