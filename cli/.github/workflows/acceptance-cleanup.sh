#!/bin/bash

set -o nounset

fats_dir=`dirname "${BASH_SOURCE[0]}"`/fats

# attempt to uninstall and cleanup test resources
if [ -d "$fats_dir" ]; then
  source $fats_dir/macros/cleanup-user-resources.sh
  kubectl delete namespace $NAMESPACE

  echo "Cleanup riff Core Runtime"
  kapp delete -n apps -a riff-core-runtime -y

  echo "Cleanup riff Knative Runtime"
  kapp delete -n apps -a riff-knative-runtime -y

  echo "Cleanup Knative"
  kapp delete -n apps -a knative -y

  echo "Cleanup Contour"
  kapp delete -n apps -a contour -y

  echo "Cleanup riff Build"
  kapp delete -n apps -a riff-build -y
  kapp delete -n apps -a riff-builders -y

  echo "Cleanup kpack"
  kapp delete -n apps -a kpack -y

  echo "Cleanup Cert Manager"
  kapp delete -n apps -a cert-manager -y

  source $fats_dir/cleanup.sh
fi
