#!/bin/bash

if [ "$#" != "1" ]
then
    echo "usage: dev_struct.sh [struct]"
    exit 255
fi

struct=$1

python .cc/gen.py struct "${struct}"
