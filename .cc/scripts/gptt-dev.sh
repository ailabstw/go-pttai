#!/bin/bash

make

./build/bin/gptt -verbosity 4 --datadir ~/.pttai-dev --metrics --metrics.influxdb --testnet
