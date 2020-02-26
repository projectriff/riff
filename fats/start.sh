#!/bin/bash

source `dirname "${BASH_SOURCE[0]}"`/.configure.sh

echo "##[group]Starting registry $REGISTRY"
source `dirname "${BASH_SOURCE[0]}"`/registries/${REGISTRY}/start.sh
echo "##[endgroup]"

echo "##[group]Starting cluster $CLUSTER"
source `dirname "${BASH_SOURCE[0]}"`/clusters/${CLUSTER}/start.sh
echo "##[endgroup]"
