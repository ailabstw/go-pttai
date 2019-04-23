#!/bin/bash

make

./build/bin/gptt -verbosity 4 --datadir ~/.pttai-test2 --metrics --metrics.influxdb  --testwebrtc --rpcport 14780 --httpaddr 127.0.0.1:9775 --port 29488 --p2pport 9488 --ipcdisable --exthttpaddr localhost:9775
