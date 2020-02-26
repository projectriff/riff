#!/bin/bash

# Create pks cluster
TS_G_ENV=$(echo $TOOLSMITH_ENV | base64 --decode | jq -r .name)
UAA_ADMIN_PASSWORD=$(echo $TOOLSMITH_ENV | base64 --decode | jq -r .pks_api.uaa_admin_password)

pks login -a api.pks.${TS_G_ENV}.cf-app.com -u admin -p ${UAA_ADMIN_PASSWORD} -k

gcp_region=`gcloud config get-value compute/region`

gcloud compute addresses create ${TS_G_ENV}-${CLUSTER_NAME}-ip --region ${gcp_region}
lb_ip=`gcloud compute addresses list --filter="name=(${TS_G_ENV}-${CLUSTER_NAME}-ip)" --format=json | jq -r .[0].address`

pks_hostname=${CLUSTER_NAME}.${TS_G_ENV}.cf-app.com

gcloud dns record-sets transaction start --zone=${TS_G_ENV}-zone
gcloud dns record-sets transaction add ${lb_ip} --name=${pks_hostname}. --ttl=300 --type=A --zone=${TS_G_ENV}-zone
gcloud dns record-sets transaction execute --zone=${TS_G_ENV}-zone

pks create-cluster ${TS_G_ENV}-${CLUSTER_NAME} --external-hostname ${pks_hostname} --plan small --wait

master_ip=`pks cluster ${TS_G_ENV}-${CLUSTER_NAME} --json | jq -r .kubernetes_master_ips[0]`
master_vm=`gcloud compute instances list --filter "tags.items = pcf-${TS_G_ENV} AND tags.items = master AND networkInterfaces.networkIP = ${master_ip}" --format "table[no-heading](name, zone)"`
master_name=`echo $master_vm | cut -d " " -f 1`
master_zone=`echo $master_vm | cut -d " " -f 2`

gcloud compute target-pools create ${TS_G_ENV}-${CLUSTER_NAME}-tp
gcloud compute target-pools add-instances ${TS_G_ENV}-${CLUSTER_NAME}-tp \
  --instances ${master_name} \
  --instances-zone ${master_zone}
gcloud compute forwarding-rules create ${TS_G_ENV}-${CLUSTER_NAME}-fr \
  --target-pool ${TS_G_ENV}-${CLUSTER_NAME}-tp \
  --address ${lb_ip} \
  --ports 8443 \
  --region ${gcp_region}

# shouldn't be nessesary, but sometimes the cluster/network isn't ready
sleep 60

pks get-credentials ${TS_G_ENV}-${CLUSTER_NAME}
