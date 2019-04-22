#!/bin/bash

if [ "$#" != 1 ]
then
    echo "usage: e2e.sh [test]"
    exit 255
fi

regex=`echo "$@" | perl -pe 's/(^|_)(\w)/\U\2/g'`
echo "regex: ${regex}"

make
make p2pbootnode

gotest -v -timeout 300s -run "${regex}" "./e2e"

echo "post-check: bootnode:"
ps ax|grep bootnode
echo ""
echo "post-check: gptt:"
ps ax|grep gptt
