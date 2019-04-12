#!/bin/sh

IMAGE_NAME="ailabstw/go-pttai:latest"
NAME="go-pttai"
THE_DIR="/home/admin/pttai.docker"
HTTPPORT="9774"
APIPORT="14779"



while [[ $# -gt 0 ]]
do
    key="$1"
    case $key in
        -i|--image)
        IMAGE_NAME="$2"
        shift # past argument
        shift # past value
        ;;
        -n|--name)
        NAME="$2"
        shift # past argument
        shift # past value
        ;;
        -p|--port)
        HTTPPORT="$2"
        shift # past argument
        shift # past value
        ;;
        -a|--aiport)
        APIPORT="$2"
        shift # past argument
        shift # past value
        ;;
    esac
done


docker pull ${image_name}
docker container stop ${name}
docker container rm ${name}
docker run -e HTTPPORT=${HTTPPORT} -e APIPORT=${APIPORT} -itd --restart=always -p 9487:9487 -p 127.0.0.1:9774:9774 -p 127.0.0.1:14779:14779 -v /home/admin/pttai.docker:/root/.pttai --name ${name} ${image_name} gptt "--testp2p" "--httpdir" "/static" "--httpaddr" "0.0.0.0:9774" "--rpcaddr" "0.0.0.0" "--exthttpaddr" "http://localhost:${HTTPPORT}" "--extrpcaddr" "http://localhost:${APIPORT}"
docker exec ${name} gptt version
