#!/bin/sh

branch=`git branch|grep '*'|awk '{print $2}'`
pwd=`pwd`

filename="${pwd}/bootnode.local"


while [[ $# -gt 0 ]]
do
    key="$1"
    case $key in
        -f|--filename)
        filename="$2"
        shift # past argument
        shift # past value
        ;;
    esac
done


docker run -itd -v ${filename}:/nodekey -p ${p2pport}:9487 -p ${port}:9774 -p ${apiport}:14779 --name go-pttai-p2pbootnoode_${branch} go-pttai-p2pbootnode:${branch} p2pbootnode "--nodekey" "/nodekey"
