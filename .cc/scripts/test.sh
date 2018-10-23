#!/bin/bash

prefix="github.com/ailabstw/go-pttai"

if [ "$#" != "1" ]
then
    echo "usage: test.sh [pkg]"
    exit 255
fi

pkg=$1

if [ -f "${pkg}/log.tmp.txt" ]
then
    rm "${pkg}/log.tmp.txt"
fi

make

gotest -v -timeout 1s "${prefix}/${pkg}"
