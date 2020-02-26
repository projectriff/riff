#!/bin/bash

pivnet_version="${1:-1.0.0}"
base_url="${2:-https://github.com/pivotal-cf/pivnet-cli/releases/download}"

# Install pivnet cli
curl -Lo pivnet ${base_url}/v${pivnet_version}/pivnet-linux-amd64-${pivnet_version} && \
  chmod +x pivnet && sudo mv pivnet /usr/local/bin/

pivnet login --api-token=${PIVNET_REFRESH_TOKEN}
