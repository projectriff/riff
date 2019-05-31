#!/bin/bash

set -o nounset

# duplicated since it may not be available via fats
travis_fold() {
  local action=$1
  local name=$2
  echo -en "travis_fold:${action}:${name}\r\033[0K"
}

# script failed, dump debug info
if [ "$TRAVIS_TEST_RESULT" = "1" ]; then
  travis_fold start debug
  sudo free -m -t
  sudo dmesg
  travis_fold end debug
fi

# attempt to cleanup fats
if [ -d "$FATSDIR" ]; then
  if [ "$TRAVIS_TEST_RESULT" = "1" ]; then
    travis_fold start system-status
    echo "System status"
    kubectl get deployments,services,pods --all-namespaces || true
    kubectl get pods --all-namespaces --field-selector=status.phase!=Running \
      | tail -n +2 | awk '{print "-n", $1, $2}' | xargs -L 1 kubectl describe pod || true
    kubectl describe node || true
    travis_fold end system-status
  fi

  travis_fold start system-uninstall
  echo "Uninstall riff system"
  duffle uninstall myriff --credentials myk8s || true
  kubectl delete namespace $NAMESPACE || true
  travis_fold end system-uninstall

  source $FATSDIR/cleanup.sh
fi
