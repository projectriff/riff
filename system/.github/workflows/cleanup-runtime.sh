#!/usr/bin/env bash

set -o nounset

readonly root_dir=$(cd `dirname $0`/../../.. && pwd)
readonly fats_dir=$root_dir/fats

source ${fats_dir}/macros/cleanup-user-resources.sh

if [ $RUNTIME = "core" ]; then
  echo "Cleanup riff Core Runtime"
  kapp delete -n apps -a riff-core-runtime -y
fi

if [ $RUNTIME = "knative" ]; then
  echo "Cleanup riff Knative Runtime"
  kapp delete -n apps -a riff-knative-runtime -y

  echo "Cleanup Knative Serving"
  kapp delete -n apps -a knative -y
fi

if [ $RUNTIME = "streaming" ]; then
  echo "Cleanup Kafka"
  kapp delete -n apps -a internal-only-kafka -y

  echo "Cleanup riff Streaming Runtime"
  kapp delete -n apps -a riff-streaming-runtime -y

  echo "Cleanup KEDA"
  kapp delete -n apps -a keda -y

  if [ $GATEWAY = "kafka" ]; then
    echo "Cleanup Kafka"
    kapp delete -n apps -a internal-only-kafka -y
  fi
  if [ $GATEWAY = "pulsar" ]; then
    echo "Cleanup Pulsar"
    kapp delete -n apps -a internal-only-pulsar -y
  fi
fi

echo "Cleanup Contour"
kapp delete -n apps -a contour -y  

echo "Cleanup riff Build"
kapp delete -n apps -a riff-build -y
kapp delete -n apps -a riff-builders -y

echo "Cleanup kpack"
kapp delete -n apps -a kpack -y

echo "Cleanup Cert Manager"
kapp delete -n apps -a cert-manager -y
