#!/bin/bash

# Install gcloud cli
`dirname "${BASH_SOURCE[0]}"`/../../install.sh gcloud

export K8S_SERVICE_TYPE=LoadBalancer
