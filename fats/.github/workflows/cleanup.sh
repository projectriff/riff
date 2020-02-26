#!/bin/bash

set -o nounset

source ${FATS_DIR}/.util.sh

source ${FATS_DIR}/macros/cleanup-user-resources.sh
kubectl delete namespace ${NAMESPACE}

echo "Removing riff Streaming Runtime"
kapp delete -n apps -a riff-streaming-runtime -y

echo "Removing KEDA"
kapp delete -n apps -a keda -y

echo "Removing riff Knative Runtime"
kapp delete -n apps -a riff-knative-runtime -y

echo "Removing Knative Serving"
kapp delete -n apps -a knative -y

echo "Removing Contour"
kapp delete -n apps -a contour -y

echo "Removing riff Core Runtime"
kapp delete -n apps -a riff-core-runtime -y

echo "Removing riff Build"
kapp delete -n apps -a riff-build -y
kapp delete -n apps -a riff-builders -y

echo "Removing kpack"
kapp delete -n apps -a kpack -y

echo "Removing Cert Manager"
kapp delete -n apps -a cert-manager -y
