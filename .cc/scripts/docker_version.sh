#!/bin/sh

name="go-pttai"

if [ $# == 1 ]
then
    name=$1
    shift
fi

docker exec ${name} gptt version
