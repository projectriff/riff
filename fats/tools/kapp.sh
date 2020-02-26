#!/bin/bash

kapp_version="${1:-0.19.0}"
base_url="${2:-https://github.com/k14s/kapp/releases/download}"

if [ "$machine" == "MinGw" ]; then
  curl -L ${base_url}/v${kapp_version}/kapp-windows-amd64.exe > kapp.exe
  mv kapp.exe /usr/bin/
else
  curl -L ${base_url}/v${kapp_version}/kapp-linux-amd64 > kapp
  chmod +x kapp
  sudo mv kapp /usr/local/bin/
fi
