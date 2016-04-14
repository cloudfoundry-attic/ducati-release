# ducati-release

This release should be deployed so that the `ducati` job co-locates with the `garden` job from guardian-release.  See below.

## What you can do
- [Running tests](#running-tests)
- [Deploy and test in isolation](#deploy-and-test-in-isolation)
- [Deploying with Diego](#deploying-with-diego)

## Running tests
```bash
docker-machine create --driver virtualbox --virtualbox-cpu-count 4 --virtualbox-memory 2048 dev-box
eval $(docker-machine env dev-box)
~/workspace/ducati-release/scripts/docker-test
```

## Deploy and test in isolation
```bash
bosh target lite
pushd ~/workspace/guardian-release
  git pull
  git submodule sync
  git submodule update --init --recursive
  bosh -n create release
  bosh upload release
popd

pushd ~/workspace/ducati-release
  git pull
  git submodule sync
  git submodule update --init --recursive
  bosh -n create release --force && bosh -n upload release
  bosh deployment manifests/ducati-manifest.yml
popd

mkdir -p ~/Downloads/releases
pushd ~/Downloads/releases
  curl -L -o consul-release.tgz https://bosh.io/d/github.com/cloudfoundry-incubator/consul-release
  bosh upload release consul-release.tgz
popd

bosh -n deploy
bosh run errand acceptance-tests
```



## Deploying with Diego

Install [Ducatify](https://github.com/cloudfoundry-incubator/ducatify/releases)

```bash
cd ~/go
go get -u github.com/cloudfoundry-incubator/ducatify/cmd/ducatify
```

Then do the BOSH dance:

```bash
bosh target lite

mkdir -p ~/Downloads/releases

pushd ~/Downloads/releases
  curl -L -o etcd-release.tgz https://bosh.io/d/github.com/cloudfoundry-incubator/etcd-release
  curl -L -o cf-release.tgz https://bosh.io/d/github.com/cloudfoundry/cf-release

  bosh upload release etcd-release.tgz
  bosh upload release cf-release.tgz
popd

pushd ~/workspace
  git clone https://github.com/sykesm/diego-release
  git clone https://github.com/cloudfoundry/cf-release
  git clone https://github.com/cloudfoundry-incubator/ducati-release
  git clone https://github.com/cloudfoundry-incubator/guardian-release
popd

pushd ~/workspace/guardian-release
  git checkout develop
  git pull
  git submodule sync
  git submodule update --init --recursive
  bosh -n create release --force && bosh -n upload release
popd

pushd ~/workspace/ducati-release
  git checkout master
  git pull
  git submodule sync
  git submodule update --init --recursive
  bosh -n create release --force && bosh -n upload release
popd

pushd ~/workspace/cf-release
  git checkout runtime-passed
  ./scripts/update
  ./scripts/generate-bosh-lite-dev-manifest
popd

pushd ~/workspace/diego-release
  git checkout ducati-dev
  ./scripts/update
  bosh -n create release
  bosh upload release
  ./scripts/generate-bosh-lite-manifests -g  # use guardian instead of garden-linux

  pushd bosh-lite/deployments
    ducatify --diego diego.yml > diego_with_ducati.yml
  popd
popd

bosh -n -d ~/workspace/cf-release/bosh-lite/deployments/cf.yml deploy
bosh -n -d ~/workspace/diego-release/bosh-lite/deployments/diego_with_ducati.yml deploy

bosh -d ~/workspace/cf-release/bosh-lite/deployments/cf.yml run errand acceptance_tests
```
