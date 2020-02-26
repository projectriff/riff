#!/bin/bash

component=$1

source_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )/${component}"

if [ -f $source_dir/templates.yaml.tpl ] ; then
  $( dirname "${BASH_SOURCE[0]}" )/apply-template.sh $source_dir/templates.yaml.tpl >  $source_dir/templates.yaml
fi
