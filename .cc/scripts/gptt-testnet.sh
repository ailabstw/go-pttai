#!/bin/bash

make

./build/bin/gptt -verbosity 4 --datadir ~/.pttai-test --metrics --metrics.influxdb  --testp2p
