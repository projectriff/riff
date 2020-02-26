#!/bin/bash

version=${1:-0.4.0-alpha.1}
base_url=${2:-https://github.com/kubernetes-sigs/aws-iam-authenticator/releases/download}

curl -s -L "${base_url}/${version}/aws-iam-authenticator_${version}_linux_amd64" -o aws-iam-authenticator
chmod +x aws-iam-authenticator
sudo mv aws-iam-authenticator /usr/local/bin
