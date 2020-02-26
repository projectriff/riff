#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

source $FATS_DIR/.configure.sh

# setup namespace
kubectl create namespace $NAMESPACE
fats_create_push_credentials $NAMESPACE
source ${FATS_DIR}/macros/create-riff-dev-pod.sh

if [ $RUNTIME = "streaming" ]; then
  echo "##[group]Create gateway"
  if [ $GATEWAY = "inmemory" ]; then
    riff streaming inmemory-gateway create test --namespace $NAMESPACE --tail
  fi
  if [ $GATEWAY = "kafka" ]; then
    riff streaming kafka-gateway create test --bootstrap-servers kafka.kafka.svc.cluster.local:9092 --namespace $NAMESPACE --tail
  fi
  if [ $GATEWAY = "pulsar" ]; then
    riff streaming pulsar-gateway create test --service-url pulsar://pulsar.pulsar.svc.cluster.local:6650 --namespace $NAMESPACE --tail
  fi
  echo "##[endgroup]"
fi

for test in java java-boot node npm command; do
  if [ $RUNTIME = "streaming" -a $test = "command" ]; then
    continue
  fi

  name=fats-cluster-uppercase-${test}
  image=$(fats_image_repo ${name})
  curl_opts="-H Content-Type:text/plain -H Accept:text/plain -d release"
  expected_data="RELEASE"

  echo "##[group]Run function $name"

  riff function create $name --image $image --namespace $NAMESPACE --tail \
    --git-repo https://github.com/$FATS_REPO --git-revision $FATS_REFSPEC --sub-path functions/uppercase/${test} &

  if [ $RUNTIME = "core" ] || [ $RUNTIME = "knative" ]; then
    riff $RUNTIME deployer create $name \
      --function-ref $name \
      --ingress-policy External \
      --namespace $NAMESPACE \
      --tail

    # cluster local invoke
    source ${FATS_DIR}/macros/invoke_incluster.sh \
      "$(kubectl get deployers.${RUNTIME}.projectriff.io ${name} --namespace ${NAMESPACE} -ojsonpath='{.status.address.url}')" \
      "${curl_opts}" \
      "${expected_data}"

    # external invoke
    source ${FATS_DIR}/macros/invoke_contour.sh \
      "$(kubectl get deployers.${RUNTIME}.projectriff.io ${name} --namespace ${NAMESPACE} -ojsonpath='{.status.url}')" \
      "${curl_opts}" \
      "${expected_data}"

    riff $RUNTIME deployer delete $name --namespace $NAMESPACE
  fi

  if [ $RUNTIME = "streaming" ]; then
    lower_stream=${name}-lower
    upper_stream=${name}-upper

    riff streaming stream create ${lower_stream} --namespace $NAMESPACE --gateway test --content-type 'text/plain' --tail
    riff streaming stream create ${upper_stream} --namespace $NAMESPACE --gateway test --content-type 'text/plain' --tail

    riff streaming processor create $name --function-ref $name --namespace $NAMESPACE --input ${lower_stream} --output ${upper_stream} --tail

    kubectl exec riff-dev -n $NAMESPACE -- subscribe ${upper_stream} --payload-encoding raw | tee result.txt &
    sleep 10
    kubectl exec riff-dev -n $NAMESPACE -- publish ${lower_stream} --payload "system" --content-type "text/plain"

    actual_data=""
    expected_data="SYSTEM"
    cnt=1
    while [ $cnt -lt 60 ]; do
      echo -n "."
      cnt=$((cnt+1))

      actual_data=$(cat result.txt | jq -r .payload)
      if [ "$actual_data" == "$expected_data" ]; then
        break
      fi

      sleep 1
    done
    fats_assert "$expected_data" "$actual_data"

    kubectl exec riff-dev -n $NAMESPACE -- pkill subscribe

    riff streaming stream delete ${lower_stream} --namespace $NAMESPACE
    riff streaming stream delete ${upper_stream} --namespace $NAMESPACE
    riff streaming processor delete $name --namespace $NAMESPACE
  fi

  riff function delete $name --namespace $NAMESPACE
  fats_delete_image $image

  echo "##[endgroup]"
done

if [ $RUNTIME = "streaming" ]; then
  riff streaming ${GATEWAY}-gateway delete test --namespace $NAMESPACE
fi
