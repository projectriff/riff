#!/bin/bash

if [[ "${FATS_LOADED:-x}" == "true" ]]; then
  return
fi
FATS_LOADED=true

ANSI_RED="\033[31;1m"
ANSI_GREEN="\033[32;1m"
ANSI_BLUE="\033[34;1m"
ANSI_RESET="\033[0m"
ANSI_CLEAR="\033[0K"

`dirname "${BASH_SOURCE[0]}"`/install.sh kubectl

unameOut="$(uname -s)"
case "${unameOut}" in
    Linux*)     machine=Linux;;
    MINGW*)     machine=MinGw;;
    MSYS_NT*)   machine=MinGw;;
    *)          machine="UNKNOWN:${unameOut}"
esac

if [ "$machine" == "MinGw" ]; then
  sudo() {
    $@
  }
fi

wait_for_service_ip() {
  local name=$1
  local namespace=$2

  wait_kube_ready \
    'service' \
    "$namespace" \
    "$name" \
    '{$.status.loadBalancer.ingress[].ip}' \
    '[0-9]'
}

wait_for_service_hostname() {
  local name=$1
  local namespace=$2
  local pattern=$3

  wait_kube_ready \
    'service' \
    "$namespace" \
    "$name" \
    '{$.status.loadBalancer.ingress[].hostname}' \
    "$pattern"
}

wait_pod_selector_ready() {
  local label=$1
  local namespace=$2

  wait_kube_ready \
    'pods' \
    "$namespace" \
    "$label" \
    '{range .items[*]}{@.metadata.name};{range @.status.conditions[*]}{@.type}={@.status};{end}{end}' \
    ';Ready=True;'
}

wait_kservice_ready() {
  local name=$1
  local namespace=$2

  wait_knative_ready 'services.serving.knative.dev' "$name" "$namespace"
}

wait_knative_ready() {
  local type=$1
  local name=$2
  local namespace=$3

  wait_kube_ready \
    "$type" \
    "$namespace" \
    "$name" \
    ';{range @.status.conditions[*]}{@.type}={@.status};{end}' \
    ';Ready=True;'
}

wait_kube_selector_exists() {
  local type=$1
  local selector=$2
  local namespace=$3
  local name=$4

  until kubectl get $type --namespace $namespace -l $selector \
    -o yaml | grep -qE $name; \
    do sleep 1; \
  done
  echo "$type found for $selector in $namespace"
  kubectl get $type --namespace $namespace -l $selector
}

wait_kube_ready() {
  local type=$1
  local namespace=$2
  local jsonpath=$4
  local pattern=$5

  if [[ $3 = *"="* ]]; then
    local selector=$3

    # TODO look for all resources to be ready, not just one
    until kubectl get $type --namespace $namespace -l $selector \
      -o jsonpath="$jsonpath" 2>&1 | grep -qE $pattern; \
      do sleep 1; \
    done

  else
    local name=$3

    until kubectl get $type --namespace $namespace $name \
      -o jsonpath="$jsonpath" 2>&1 | grep -qE $pattern; \
      do sleep 1; \
    done
  fi
}

fats_echo() {
  echo -e "$ANSI_BLUE[`date -u +%Y-%m-%dT%H:%M:%SZ`]$ANSI_RESET $@"
}

# derived from https://github.com/travis-ci/travis-build/blob/4f580b238530108cdd08719c326cd571d4e7b99f/lib/travis/build/bash/travis_retry.bash
# MIT licenced https://github.com/travis-ci/travis-build/blob/4f580b238530108cdd08719c326cd571d4e7b99f/LICENSE
fats_retry() {
  local result=0
  local count=1
  while [[ "${count}" -le 3 ]]; do
    [[ "${result}" -ne 0 ]] && {
      echo -e "\\n${ANSI_RED}The command \"${*}\" failed. Retrying, ${count} of 3.${ANSI_RESET}\\n" >&2
    }
    "${@}" && { result=0 && break; } || result="${?}"
    count="$((count + 1))"
    sleep 1
  done

  [[ "${count}" -gt 3 ]] && {
    echo -e "\\n${ANSI_RED}The command \"${*}\" failed 3 times.${ANSI_RESET}\\n" >&2
  }

  return "${result}"
}

fats_assert() {
  local expected_data=${1}
  local actual_data=${2}

  if [ "$actual_data" != "$expected_data" ]; then
    echo -e "${ANSI_RED}did not produce expected result${ANSI_RESET}:";
    echo -e "   expected: $expected_data"
    echo -e "   actual: $actual_data"
    exit 1
  fi
}
