#!/bin/bash

cd pttai.js

#./scripts/setup.sh

yarn build

cd ..
mkdir -p ./static
rm -rf ./static/*
cp -R pttai.js/build/* ./static
