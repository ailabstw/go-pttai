#!/bin/bash

branch=`git branch|grep '*'|awk '{print $2}'`

docker container stop go-pttai_${branch}