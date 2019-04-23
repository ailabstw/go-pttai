#!/bin/bash

make

rm log.tmp.txt
./build/bin/gptt --datadir ~/.pttai --metrics --metrics.influxdb  --testwebrtc --log log.tmp.txt  2> log.err.txt
