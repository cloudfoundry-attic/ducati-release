#!/bin/bash

set -e -x -u

bosh logs garden 0 --job
bosh logs garden 1 --job

rm -rf garden0
rm -rf garden1

mkdir -p garden0
pushd garden0
  tar xzf ../garden.0.*.tgz
popd

mkdir -p garden1
pushd garden1
  tar xzf ../garden.1.*.tgz
popd

cat garden0/ducati/ducatid.stdout.log garden1/ducati/ducatid.stdout.log | sort > all.log

rm -rf garden*

