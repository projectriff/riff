#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

readonly component=$1
readonly version=$2

readonly root_dir=$(cd `dirname $0`/../.. && pwd)
readonly build_dir="${root_dir}/release/build/${component}"
readonly source_dir="${root_dir}/release/source/${component}"
readonly target_dir="${root_dir}/release/target"

mkdir -p ${build_dir}

if [ -f ${source_dir}/values.yaml.tpl ] ; then
  $(cd `dirname $0` && pwd)/apply-template.sh ${source_dir}/values.yaml.tpl > ${source_dir}/values.yaml
fi

if [ -f ${source_dir}/values.yaml ] ; then
  if [ -f ${build_dir}/values.yaml ] ; then
    # merge custom values
    yq merge -i -x ${build_dir}/values.yaml ${source_dir}/values.yaml
  else
    cp ${source_dir}/values.yaml ${build_dir}/values.yaml
  fi
fi

if [ -f ${source_dir}/Chart.yaml ] ; then
  cp ${source_dir}/Chart.yaml ${build_dir}/Chart.yaml
fi

if [ -f ${source_dir}/requirements.yaml ] ; then
  cp ${source_dir}/requirements.yaml ${build_dir}/requirements.yaml
fi

if [ -d ${source_dir}/charts ] ; then
  mkdir -p ${build_dir}/charts
  cp -LR ${source_dir}/charts/* ${build_dir}/charts/
fi

if [ $component == "kafka" ] ; then
  helm package ${build_dir}/../kafka --destination repository --version ${version}
fi

# download config and apply overlays
file=${target_dir}/${component}.yaml
rm -f $file

if [ -f ${source_dir}/templates.yaml ] ; then
  while IFS= read -r line
  do
    arr=($line)
    name=${arr[0]%?}
    url=${arr[1]}
    args=$(echo $line | cut -d "#" -s -f 2)

    echo "" >> ${file}
    echo "---" >> ${file}
    if [[ $url = 'riff+ko://'* ]]; then
      ko resolve -f ${root_dir}$(echo ${url} | cut -c11-) >> ${file}
    else
      curl -L -s ${url} >> ${file}
    fi
  done < "${source_dir}/templates.yaml"
fi

if [ $component == "kafka" ] ; then
  helm template ./repository/kafka-*.tgz --namespace kafka > ${file}

  cat ${file} | sed -e 's/release-name-//g' | sed -e 's/release-name/riff/g' > ${file}.tmp
  mv ${file}.tmp ${file}
fi

if [ -f ${source_dir}/release.patch ] ; then
  patch ${file} ${source_dir}/release.patch
fi

if [ -d ${source_dir}/overlays-release ] ; then
  ytt -f ${source_dir}/overlays-release/ -f ${file} --file-mark $(basename ${file}):type=yaml-plain --ignore-unknown-comments > ${file}.tmp
  mv ${file}.tmp ${file}
fi

# resolve tags to digests
k8s-tag-resolver ${file} -o ${file}.tmp
mv ${file}.tmp ${file}
