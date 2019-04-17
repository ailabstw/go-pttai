#!/bin/bash

fromYear=$1
toYear=$2

find `pwd` -name "*.go" -print|xargs -t -I {} /bin/bash -c "sed -i '' 's/Copyright ${fromYear} The go-pttai Authors/Copyright ${toYear} The go-pttai Authors/g' {}"
