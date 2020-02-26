#!/bin/bash

kubebuilder_version="${1:-2.0.0}"


# from https://book.kubebuilder.io/quick-start.html#installation
os=`go env GOOS`
arch=`go env GOARCH`

# download kubebuilder and extract it to tmp
curl -sL https://go.kubebuilder.io/dl/${kubebuilder_version}/${os}/${arch} | tar -xz -C /tmp/

# move to a long-term location and put it on your path
# (you'll need to set the KUBEBUILDER_ASSETS env var if you put it somewhere else)
sudo mv /tmp/kubebuilder_${kubebuilder_version}_${os}_${arch} /usr/local/kubebuilder
echo "##[add-path]/usr/local/kubebuilder/bin"
