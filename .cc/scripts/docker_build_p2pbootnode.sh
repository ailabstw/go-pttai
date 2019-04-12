#!/bin/sh

branch=`git branch|grep '*'|awk '{print $2}'`

docker build -t go-pttai-p2pbootnode:${branch} -f Dockerfile.bootnode .
