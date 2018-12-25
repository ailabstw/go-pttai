#!/bin/bash

make

./build/bin/gptt -verbosity 4 --datadir ~/.pttai --metrics --metrics.influxdb  --testp2p
