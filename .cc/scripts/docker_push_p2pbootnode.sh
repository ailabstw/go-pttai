#!/bin/sh

branch=`git branch|grep '*'|awk '{print $2}'`

docker tag go-pttai-p2pbootnode:${branch} ailabstw/go-pttai-p2pbootnode:latest
docker push ailabstw/go-pttai-p2pbootnode:latest
