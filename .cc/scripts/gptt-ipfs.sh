#!/bin/bash

make

./build/bin/gptt -verbosity 4 --datadir ~/.pttai-ipfs --metrics --metrics.influxdb  --ipfsp2p
