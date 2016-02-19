#!/bin/bash

set -e -u -x

cd ducati-release
export GOPATH=$PWD

declare -a packages=(
  "src/github.com/cloudfoundry-incubator/ducati-daemon"
  "src/github.com/cloudfoundry-incubator/guardian-cni-adapter"
  "src/github.com/cloudfoundry-incubator/ducati-cni-plugins"
  "src/integration"
  "src/github.com/cloudfoundry-incubator/ducati"
  )

for dir in "${packages[@]}"; do
  pushd $dir
    ginkgo -r  -randomizeAllSpecs -randomizeSuites "$@"
  popd
done
