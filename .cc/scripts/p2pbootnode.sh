#!/bin/bash

BOOTNODE_FILE="bootnode.local"
PORT=9487

if [ "$#" -ge "1" ]
then
    BOOTNODE_FILE=$1
fi

if [ "$#" -ge "2" ]
then
    PORT=$2
fi

make p2pbootnode

if [ ! -f ${BOOTNODE_FILE} ]
then
    ./build/bin/p2pbootnode --genkey ${BOOTNODE_FILE}
fi

./build/bin/p2pbootnode --nodekey ${BOOTNODE_FILE} --addr "/ip4/0.0.0.0/tcp/${PORT}/http/p2p-webrtc-direct"

