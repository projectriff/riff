#!/bin/bash

kbld_version="${1:-0.11.0}"
base_url="${2:-https://github.com/k14s/kbld/releases/download}"

if [ "$machine" == "MinGw" ]; then
  curl -L ${base_url}/v${kbld_version}/kbld-windows-amd64.exe > kbld.exe
  mv kbld.exe /usr/bin/
else
  curl -L ${base_url}/v${kbld_version}/kbld-linux-amd64 > kbld
  chmod +x kbld
  sudo mv kbld /usr/local/bin/
fi
