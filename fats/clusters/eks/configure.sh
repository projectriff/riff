#!/bin/bash

# Install eksctl cli
`dirname "${BASH_SOURCE[0]}"`/../../install.sh eksctl

export K8S_SERVICE_TYPE=LoadBalancer
