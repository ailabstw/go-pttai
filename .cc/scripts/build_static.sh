#!/bin/bash

cd pttai.js

npm install
npm run build

cd ..
mkdir -p ./static
rm -rf ./static/*
cp -R pttai.js/build/* ./static
