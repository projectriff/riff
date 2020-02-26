#!/bin/bash

ytt_version="${1:-0.22.0}"
base_url="${2:-https://github.com/k14s/ytt/releases/download}"

if [ "$machine" == "MinGw" ]; then
  curl -L ${base_url}/v${ytt_version}/ytt-windows-amd64.exe > ytt.exe
  mv ytt.exe /usr/bin/
else
  curl -L ${base_url}/v${ytt_version}/ytt-linux-amd64 > ytt
  chmod +x ytt
  sudo mv ytt /usr/local/bin/
fi
