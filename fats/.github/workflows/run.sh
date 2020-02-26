#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

source ${FATS_DIR}/.configure.sh

# install tools
${FATS_DIR}/install.sh riff
${FATS_DIR}/install.sh kubectl

kubectl create ns $NAMESPACE
fats_create_push_credentials ${NAMESPACE}
source ${FATS_DIR}/macros/create-riff-dev-pod.sh

if [ "${machine}" != "MinGw" ]; then
  modes="cluster local"
else
  modes="cluster"
fi

for mode in ${modes}; do
  # functions
  # workaround for https://github.com/projectriff/node-function-invoker/issues/113
  if [ ${CLUSTER} = "pks-gcp" ]; then
    languages="command java java-boot"
  else
    languages="command node npm java java-boot"
  fi
  for test in ${languages}; do
    name=fats-${mode}-fn-uppercase-${test}
    image=$(fats_image_repo ${name})

    echo "##[group]Run function ${name}"

    if [ "${mode}" == "cluster" ]; then
      riff function create ${name} \
        --image ${image} \
        --namespace ${NAMESPACE} \
        --git-repo https://github.com/${FATS_REPO} \
        --git-revision ${FATS_REFSPEC} \
        --sub-path functions/uppercase/${test} \
        --tail
    elif [ "${mode}" == "local" ]; then
      riff function create ${name} \
        --image ${image} \
        --namespace ${NAMESPACE} \
        --local-path ${FATS_DIR}/functions/uppercase/${test} \
        --tail
    else
      echo "Unknown mode: ${mode}"
      exit 1
    fi

    # core runtime
    riff core deployer create ${name}-core \
      --function-ref ${name} \
      --ingress-policy External \
      --namespace ${NAMESPACE} \
      --tail
    source ${FATS_DIR}/macros/invoke_contour.sh \
      "$(kubectl get deployers.core.projectriff.io --namespace $NAMESPACE ${name}-core -o jsonpath='{$.status.url}')" \
      "-H Content-Type:text/plain -H Accept:text/plain -d fats" \
      FATS
    source ${FATS_DIR}/macros/invoke_incluster.sh \
      "$(kubectl get deployers.core.projectriff.io --namespace $NAMESPACE ${name}-core -o jsonpath='{$.status.address.url}')" \
      "-H Content-Type:text/plain -H Accept:text/plain -d fats" \
      FATS
    riff core deployer delete ${name}-core --namespace ${NAMESPACE}

    # knative runtime
    riff knative deployer create ${name}-knative \
      --function-ref ${name} \
      --ingress-policy External \
      --namespace ${NAMESPACE} \
      --tail
    source ${FATS_DIR}/macros/invoke_contour.sh \
      "$(kubectl get deployers.knative.projectriff.io --namespace $NAMESPACE ${name}-knative -o jsonpath='{$.status.url}')" \
      "-H Content-Type:text/plain -H Accept:text/plain -d fats" \
      FATS
    source ${FATS_DIR}/macros/invoke_incluster.sh \
      "$(kubectl get deployers.knative.projectriff.io --namespace $NAMESPACE ${name}-knative -o jsonpath='{$.status.address.url}')" \
      "-H Content-Type:text/plain -H Accept:text/plain -d fats" \
      FATS
    riff knative deployer delete ${name}-knative --namespace ${NAMESPACE}

    # cleanup
    riff function delete ${name} --namespace ${NAMESPACE}
    fats_delete_image ${image}

    echo "##[endgroup]"
  done

  # applications
  for test in node java-boot; do
    name=fats-${mode}-app-uppercase-${test}
    image=$(fats_image_repo ${name})

    echo "##[group]Run application ${name}"

    if [ "${mode}" == "cluster" ]; then
      riff application create ${name} \
        --image ${image} \
        --namespace ${NAMESPACE} \
        --git-repo https://github.com/${FATS_REPO} \
        --git-revision ${FATS_REFSPEC} \
        --sub-path applications/uppercase/${test} \
        --tail
    elif [ "${mode}" == "local" ]; then
      riff application create ${name} \
        --image ${image} \
        --namespace ${NAMESPACE} \
        --local-path ${FATS_DIR}/applications/uppercase/${test} \
        --tail
    else
      echo "Unknown mode: ${mode}"
      exit 1
    fi

    # core runtime
    riff core deployer create ${name}-core \
      --application-ref ${name} \
      --ingress-policy External \
      --namespace ${NAMESPACE} \
      --tail
    source ${FATS_DIR}/macros/invoke_contour.sh \
      "$(kubectl get deployers.core.projectriff.io --namespace $NAMESPACE ${name}-core -o jsonpath='{$.status.url}')" \
      "--get --data-urlencode input=fats" \
      FATS
    source ${FATS_DIR}/macros/invoke_incluster.sh \
      "$(kubectl get deployers.core.projectriff.io --namespace $NAMESPACE ${name}-core -o jsonpath='{$.status.address.url}')" \
      "--get --data-urlencode input=fats" \
      FATS
    riff core deployer delete ${name}-core --namespace ${NAMESPACE}

    # knative runtime
    riff knative deployer create ${name}-knative \
      --application-ref ${name} \
      --ingress-policy External \
      --namespace ${NAMESPACE} \
      --tail
    source ${FATS_DIR}/macros/invoke_contour.sh \
      "$(kubectl get deployers.knative.projectriff.io --namespace $NAMESPACE ${name}-knative -o jsonpath='{$.status.url}')" \
      "--get --data-urlencode input=fats" \
      FATS
    source ${FATS_DIR}/macros/invoke_incluster.sh \
      "$(kubectl get deployers.knative.projectriff.io --namespace $NAMESPACE ${name}-knative -o jsonpath='{$.status.address.url}')" \
      "--get --data-urlencode input=fats" \
      FATS
    riff knative deployer delete ${name}-knative --namespace ${NAMESPACE}

    # cleanup
    riff application delete ${name} --namespace ${NAMESPACE}
    fats_delete_image ${image}

    echo "##[endgroup]"
  done
done
