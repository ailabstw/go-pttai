#!/bin/bash

# Windows installer instruction
#
#   1. Run the below command before running this script (on UNIX)
#      >> make gptt-windows-amd64
#   2. Prepare a Windows device and
#      a. install golang, NSIS and git
#      b. git clone go-pttai project under go
#      c. copy the binary output from step 1 and rename it to gptt-windows.exe
#   3. Run this script with input argumnet being the path to gptt-windows.exe (on Windows 64-bits)
#      >> .cc\scripts\build_windows.sh path\to\gptt-windows.exe
#   4. Run the below command to create installer (on Windows 64-bits)
#      >> go run build\ci.go nsis
#   5. The windows installer will be created in the same directory
#


if [[ -z "$1" ]]; then
  echo "please provide path to gptt-windows.exe as input arguemnt"
  exit 1
fi

dir="build/bin/Windows"

rm -rf "${dir}"
mkdir -p "${dir}"

cp "$1" "${dir}/gptt.exe"
cp -R static "${dir}"
cp build/gptt.ico "${dir}/gptt.ico"
cp build/run-gptt-windows.sh "${dir}/gptt-run.bat"
