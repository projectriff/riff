#!/bin/bash

# Install aws cli
`dirname "${BASH_SOURCE[0]}"`/../install.sh aws
`dirname "${BASH_SOURCE[0]}"`/../install.sh aws-iam-authenticator

eksctl_version="${1:-0.1.16}"
base_url=${2:-https://github.com/weaveworks/eksctl/releases/download}
eksctl_dir=`mktemp -d eksctl.XXXX`

curl -s -L "${base_url}/${eksctl_version}/eksctl_Linux_amd64.tar.gz" | tar xz -C $eksctl_dir
chmod +x $eksctl_dir/eksctl
sudo mv $eksctl_dir/eksctl /usr/local/bin

rm -rf $eksctl_dir
