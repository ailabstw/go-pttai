#!/bin/bash

prefixDir="build/bin/MacOS"
contentDir="${prefixDir}/Applications/gptt.app/Contents"
dir="${contentDir}/MacOS"
res="${contentDir}/Resources"

rm -rf "${prefixDir}"

mkdir -p "${dir}"
mkdir -p "${res}"


cp build/bin/gptt "${dir}/gptt.bin"
cp build/run-gptt.sh "${dir}/gptt"
cp build/run-gptt.sh "${dir}/run-gptt.sh"

cp build/Info.plist "${contentDir}"
cp build/gptt.icns "${res}/gptt.icns"

productbuild --resources "${res}" --content build/bin/MacOS build/bin/gptt.pkg
