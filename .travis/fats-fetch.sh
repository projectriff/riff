#!/bin/bash

dir=${1}
refspec=${2:-48d839feae356d43f76008f2ee495b0300cd503a} # projectriff/fats eventing-ready as of 2018-12-12

if [ ! -f $dir ]; then
  mkdir -p $dir
  curl -L https://github.com/projectriff/fats/archive/${refspec}.tar.gz | \
    tar xz -C $dir --strip-components 1
fi
