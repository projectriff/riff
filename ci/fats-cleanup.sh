#!/bin/bash

set -o nounset

fats_dir=`dirname "${BASH_SOURCE[0]}"`/fats

# attempt to cleanup fats
if [ -d "$fats_dir" ]; then
  echo "Uninstall riff system"
  duffle uninstall riff --credentials k8s || true
  kubectl delete namespace $NAMESPACE || true

  source $fats_dir/cleanup.sh
fi
