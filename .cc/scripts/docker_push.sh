#!/bin/bash

branch=`git branch|grep '*'|awk '{print $2}'`

docker tag go-pttai:${branch} ailabstw/go-pttai:latest
docker push ailabstw/go-pttai:latest
