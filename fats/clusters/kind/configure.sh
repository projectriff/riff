#!/bin/bash

# Install kind cli
`dirname "${BASH_SOURCE[0]}"`/../../install.sh kind

export K8S_SERVICE_TYPE=NodePort
