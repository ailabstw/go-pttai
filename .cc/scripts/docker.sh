#!/bin/bash

branch=`git branch|grep '*'|awk '{print $2}'`

port=9774
apiport=14779

while [[ $# -gt 0 ]]
do
    key="$1"
    case $key in
        -p|--port)
        port="$2"
        shift # past argument
        shift # past value
        ;;
        -a|--apiport)
        apiport="$2"
        shift # past argument
        shift # past value
        ;;
    esac
done

echo ""
echo "http://localhost:${port}"
echo ""
sleep 1

docker run --rm -itd -p ${port}:9774 -p ${apiport}:14779 --name go-pttai_${branch} go-pttai:${branch}
