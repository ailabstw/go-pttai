#!/bin/bash

if [ "$#" != "1" ]
then
    echo "usage: e2e.sh [e2e] (no test in the end)"
    exit 255
fi

e2e=$1

python .cc/gen.py e2e "${e2e}"
