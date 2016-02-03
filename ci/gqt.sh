#!/bin/bash

(
  cd $(dirname $0)/../..

  export GOROOT=/usr/local/go
  export PATH=$GOROOT/bin:$PATH

  export GOPATH=$PWD
  export PATH=$GOPATH/bin:$PATH

  go build github.com/cloudfoundry-incubator/guardian-cni-adapter
)

export GARDEN_NETWORK_PLUGIN=/var/vcap/packages/cni-plugins/bin/guardian-cni-adapter
export GARDEN_NETWORK_PLUGIN_ARGS="-cniPluginDir,/var/vcap/packages/cni-plugins/bin,-cniConfigDir,/var/vcap/jobs/ducati/cni-conf"

guardian-release/ci/scripts/guardian
