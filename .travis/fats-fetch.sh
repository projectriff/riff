#!/bin/bash

dir=${1}
refspec=${2:-fd45d79d84e51ee086ed51bd182dd05699639009} # projectriff/fats eventing-ready as of 2018-12-12

if [ ! -f $dir ]; then
  mkdir -p $dir
  curl -L https://github.com/projectriff/fats/archive/${refspec}.tar.gz | \
    tar xz -C $dir --strip-components 1
fi
