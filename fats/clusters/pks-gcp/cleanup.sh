#!/bin/bash

TS_G_ENV=$(echo $TOOLSMITH_ENV | base64 --decode | jq -r .name)

gcp_region=`gcloud config get-value compute/region`
lb_ip=`gcloud compute addresses list --filter="name=(${TS_G_ENV}-${CLUSTER_NAME}-ip)" --format=json | jq -r .[0].address`

pks_hostname=${CLUSTER_NAME}.${TS_G_ENV}.cf-app.com

gcloud dns record-sets transaction start --zone=${TS_G_ENV}-zone
gcloud dns record-sets transaction remove ${lb_ip} --name=${pks_hostname}. --ttl=300 --type=A --zone=${TS_G_ENV}-zone
gcloud dns record-sets transaction execute --zone=${TS_G_ENV}-zone

gcloud compute forwarding-rules delete ${TS_G_ENV}-${CLUSTER_NAME}-fr --region ${gcp_region}
gcloud compute target-pools delete ${TS_G_ENV}-${CLUSTER_NAME}-tp
gcloud compute addresses delete ${TS_G_ENV}-${CLUSTER_NAME}-ip

pks delete-cluster ${TS_G_ENV}-${CLUSTER_NAME} --non-interactive --wait

# delete orphaned resources

cluster_prefix="${TS_G_ENV}-`echo $CLUSTER_NAME | cut -d '-' -f1`"
before=`date -d @$(( $(date +"%s") - 24*3600)) -u +%Y-%m-%dT%H:%M:%SZ` # yesterday

# TODO restore once we can check the creation timestamp
# fats_echo "Cleanup orphaned clusters"
# pks clusters
# pks clusters --json | jq -r '.[].name' | \
#   xargs -L 1 --no-run-if-empty pks delete-cluster --non-interactive --wait
# pks clusters

fats_echo "Cleanup orphaned forwarding rules"
gcloud compute forwarding-rules list --filter="name ~ ^$cluster_prefix- AND createTime < $before"
gcloud compute forwarding-rules list --filter="name ~ ^$cluster_prefix- AND createTime < $before" --format="table[no-heading](name, region)" | \
  sed 's/ / --region /2' | \
  xargs -L 1 --no-run-if-empty gcloud compute forwarding-rules delete
gcloud compute forwarding-rules list --filter="name ~ ^$cluster_prefix- AND createTime < $before"

fats_echo "Cleanup orphaned target instances"
gcloud compute target-instances list --filter="name ~ ^$cluster_prefix- AND createTime < $before"
gcloud compute target-instances list --filter="name ~ ^$cluster_prefix- AND createTime < $before" --format="table[no-heading](name, zone)" | \
  sed 's/ / --zone /2' | \
  xargs -L 1 --no-run-if-empty gcloud compute target-instances delete
gcloud compute target-instances list --filter="name ~ ^$cluster_prefix- AND createTime < $before"

fats_echo "Cleanup orphaned addresses"
gcloud compute addresses list --filter="name ~ ^$cluster_prefix- AND createTime < $before"
gcloud compute addresses list --filter="name ~ ^$cluster_prefix- AND createTime < $before" --format="table[no-heading](name, region)" | \
  sed 's/ / --region /2' | \
  xargs -L 1 --no-run-if-empty gcloud compute addresses delete
gcloud compute addresses list --filter="name ~ ^$cluster_prefix- AND createTime < $before"

fats_echo "Cleanup orphaned disks"
gcloud compute disks list --filter="name ~ ^disk- AND createTime < $before"
gcloud compute disks list --filter="name ~ ^disk- AND createTime < $before" --format="table[no-heading](name, zone)" | \
  sed 's/ / --zone /2' | \
  xargs -L 1 --no-run-if-empty gcloud compute disks delete
gcloud compute disks list --filter="name ~ ^disk- AND createTime < $before"
