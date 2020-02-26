#!/bin/bash

gcloud container clusters create $CLUSTER_NAME \
  --cluster-version=1.15 \
  --machine-type=n1-standard-2 \
  --enable-autoscaling --min-nodes=1 --max-nodes=3 \
  --enable-autorepair \
  --scopes=cloud-platform \
  --num-nodes=3

kubectl create clusterrolebinding cluster-admin-binding \
  --clusterrole=cluster-admin \
  --user=$(gcloud config get-value core/account)

echo "Wait for GKE to be fully ready"
fats_retry kubectl wait pods --for=condition=Ready --all --namespace kube-system --timeout=60s
fats_retry kubectl wait apiservices --for=condition=Available --all --timeout=60s
