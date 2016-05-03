#!/bin/bash

set -e -u

THIS_DIR=$(cd $(dirname $0) && pwd)
cd $THIS_DIR

export CONFIG=/tmp/bosh-lite-integration-config.json
export APP_DIR=./example-apps/proxy

echo '
{
  "api": "api.bosh-lite.com",
  "admin_user": "admin",
  "admin_password": "admin",
  "apps_domain": "bosh-lite.com",
  "skip_ssl_validation": true,
  "use_http": true
}
' > $CONFIG

ginkgo -v .
