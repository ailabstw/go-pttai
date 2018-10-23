#!/bin/bash

if [ "$#" != "1" ]
then
    echo "usage: module.sh [module]"
    exit 255
fi

module=$1

python .cc/gen.py module "${module}"
