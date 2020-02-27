#!/bin/bash

# avoid installing in the current directory since that may be a module
( cd / && GO111MODULE=on go get github.com/projectriff/k8s-manifest-scanner/cmd/k8s-tag-resolver )
