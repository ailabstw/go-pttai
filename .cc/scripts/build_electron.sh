#!/bin/bash

electronDir="build/electron"
contentDir="${electronDir}/app"

if [ -e build/bin/gptt ]; then
    cp build/bin/gptt "${contentDir}/gptt.bin"
fi

if [ -e build/bin/gptt-windows-4.0-amd64.exe ]; then
    cp build/bin/gptt-windows-4.0-amd64.exe "${contentDir}/gptt.exe"
fi

cp -R static "${contentDir}"

cd "${contentDir}"

npm install
npm run dist

echo "installer will be in $electronDir/dist"

