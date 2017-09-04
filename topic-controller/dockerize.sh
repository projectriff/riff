#!/usr/bin/env bash

# This builds the topic-controller from Go sources on *your* machine, targeting Linux OS
# and linking everything statically, to minimize Docker image size
# See e.g. https://blog.codeship.com/building-minimal-docker-containers-for-go-applications/ for details
CGO_ENABLED=0 GOOS=linux go build -v -a -installsuffix cgo -o bin/topic github.com/sk8s/controller/topic

docker build . -f Dockerfile-topic -t sk8s/topic-controller:v0001