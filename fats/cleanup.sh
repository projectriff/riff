#!/bin/bash

set -o nounset

source `dirname "${BASH_SOURCE[0]}"`/.configure.sh

echo "##[group]Cleanup cluster $CLUSTER"
source `dirname "${BASH_SOURCE[0]}"`/clusters/${CLUSTER}/cleanup.sh
echo "##[endgroup]"

echo "##[group]Cleanup registry $REGISTRY"
source `dirname "${BASH_SOURCE[0]}"`/registries/${REGISTRY}/cleanup.sh
echo "##[endgroup]"
