#!/bin/bash

BOOTNODE_FILE="bootnode.local"
if [ "$#" == "1" ]
then
    BOOTNODE_FILE=$1
fi

make p2pbootnode

if [ ! -f ${BOOTNODE_FILE} ]
then
    ./build/bin/p2pbootnode --genkey ${BOOTNODE_FILE}
fi


./build/bin/p2pbootnode --nodekey ${BOOTNODE_FILE}
