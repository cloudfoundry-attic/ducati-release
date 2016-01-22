#!/bin/bash

set -e -u -x

cd ducati-release
export GOPATH=$PWD

pushd src/github.com/cloudfoundry-incubator/ducati-cni-plugins
  ginkgo -r
popd

pushd src/github.com/cloudfoundry-incubator/ducati
  ginkgo -r
popd

