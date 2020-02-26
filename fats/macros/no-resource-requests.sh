#!/bin/bash

if [ $(kubectl get nodes -oname | wc -l) = "1" ]; then
  echo "Eliminate pod resource requests"
  fats_retry kubectl apply -f https://storage.googleapis.com/projectriff/no-resource-requests-webhook/no-resource-requests-webhook-20191121210956-521ae2a8c3323540.yaml
  wait_pod_selector_ready app=webhook no-resource-requests
fi
