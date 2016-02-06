#!/bin/bash

set -e -x -u

cd ducati-release
export GOPATH=$PWD

cd src/acceptance

ginkgo -r -failFast -randomizeAllSpecs
