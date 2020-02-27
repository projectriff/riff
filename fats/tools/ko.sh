#!/bin/bash

ko_version="${1:-v0.3.0}"

# avoid installing in the current directory since that may be a module
(cd .. && GO111MODULE=on go get github.com/google/ko/cmd/ko@${ko_version})
