#!/bin/bash

# delete job resources

gcloud container clusters delete $CLUSTER_NAME

if [ "$machine" != "MinGw" ]; then
  # delete orphaned resources

  cluster_prefix=`echo $CLUSTER_NAME | cut -d '-' -f1`
  before=`date -d @$(( $(date +"%s") - 24*3600)) -u +%Y-%m-%dT%H:%M:%SZ` # yesterday

  fats_echo "Cleanup orphaned clusters"
  gcloud container clusters list --filter="name ~ ^$cluster_prefix- AND createTime < $before"
  gcloud container clusters list --filter="name ~ ^$cluster_prefix- AND createTime < $before" --format="table[no-heading](name, zone)" | \
    sed 's/ / --zone /2' | \
    xargs -L 1 --no-run-if-empty gcloud container clusters delete
  gcloud container clusters list --filter="name ~ ^$cluster_prefix- AND createTime < $before"

  fats_echo "Cleanup orphaned target pools"
  gcloud compute target-pools list --filter="createTime < $before"
  gcloud compute target-pools list --filter="createTime < $before" --format="table[no-heading](name, region)" | \
    sed 's/ / --region /2' | \
    xargs -L 1 --no-run-if-empty gcloud compute target-pools delete
  gcloud compute target-pools list --filter="createTime < $before"

  fats_echo "Cleanup orphaned firewall rules"
  gcloud compute firewall-rules list --filter="name ~ $CLUSTER_NAME OR (name ~ ^gke-$cluster_prefix- AND createTime < $before)"
  gcloud compute firewall-rules list --filter="name ~ $CLUSTER_NAME OR (name ~ ^gke-$cluster_prefix- AND createTime < $before)" --format="table[no-heading](name)" | \
    xargs -L 1 --no-run-if-empty gcloud compute firewall-rules delete
  gcloud compute firewall-rules list --filter="name ~ $CLUSTER_NAME OR (name ~ ^gke-$cluster_prefix- AND createTime < $before)"

  fats_echo "Cleanup orphaned health checks"
  gcloud compute http-health-checks list --filter="name ~ ^k8s- AND createTime < $before"
  gcloud compute http-health-checks list --filter="name ~ ^k8s- AND createTime < $before" --format="table[no-heading](name)" | \
    xargs -L 1 --no-run-if-empty gcloud compute http-health-checks delete
  gcloud compute http-health-checks list --filter="name ~ ^k8s- AND createTime < $before"

  fats_echo "Cleanup orphaned disks"
  gcloud compute disks list --filter="name ~ $CLUSTER_NAME OR (name ~ ^gke-$cluster_prefix- AND createTime < $before)"
  gcloud compute disks list --filter="name ~ $CLUSTER_NAME OR (name ~ ^gke-$cluster_prefix- AND createTime < $before)" --format="table[no-heading](name, zone)" | \
    sed 's/ / --zone /2' | \
    xargs -L 1 --no-run-if-empty gcloud compute disks delete
  gcloud compute disks list --filter="name ~ $CLUSTER_NAME OR (name ~ ^gke-$cluster_prefix- AND createTime < $before)"
fi
