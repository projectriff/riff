#!/bin/bash

uninstall_app() {
  local name=$1

  kapp delete -n apps -a $name -y
}

source $FATS_DIR/macros/cleanup-user-resources.sh

if [ $RUNTIME = "core" ]; then
  echo "Uninstall riff Core runtime"
  uninstall_app riff-core-runtime
fi

if [ $RUNTIME = "knative" ]; then
  echo "Uninstall riff Knartive runtime"
  uninstall_app riff-knative-runtime
  uninstall_app knative
fi

if [ $RUNTIME = "streaming" ]; then
  echo "Uninstall riff Streaming runtime"
  uninstall_app riff-streaming-runtime
  uninstall_app keda

  if [ $GATEWAY = "kafka" ]; then
    echo "Uninstall Kafka"
    uninstall_app internal-only-kafka
  fi
  if [ $GATEWAY = "pulsar" ]; then
    echo "Uninstall Pulsar"
    uninstall_app internal-only-pulsar
  fi
fi

echo "Uninstall riff Build"
uninstall_app riff-build
uninstall_app riff-builders
uninstall_app kpack

echo "Uninstall Contour"
uninstall_app contour

echo "Uninstall Cert Manager"
uninstall_app cert-manager
