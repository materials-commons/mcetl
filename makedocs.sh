#!/bin/sh

for package in $(go list ./...)
do
    DIR=$(echo $package | sed 's%github.com/materials-commons/mcetl%.%')
    mkdir -p docs/$DIR
    godoc $package > docs/$DIR/package.txt
done
