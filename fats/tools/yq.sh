#!/bin/bash

yq_version="${1:-2.4.1}"

curl -L https://github.com/mikefarah/yq/releases/download/${yq_version}/yq_linux_amd64 -o yq
chmod +x yq
sudo mv yq /usr/local/bin/
