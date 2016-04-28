# ducati-release

This release should be deployed so that the `ducati` job co-locates with the `garden` job from garden-runc-release.  See below.

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
pushd ~/workspace/garden-runc-release
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

Then get, create, and upload the necessary releases:

```bash
bosh target lite

mkdir -p ~/Downloads/releases

pushd ~/Downloads/releases
  curl -L -o etcd-release.tgz https://bosh.io/d/github.com/cloudfoundry-incubator/etcd-release
  curl -L -o cf-release.tgz https://bosh.io/d/github.com/cloudfoundry/cf-release

  bosh upload release etcd-release.tgz
popd

pushd ~/workspace
  git clone https://github.com/sykesm/diego-release
  git clone https://github.com/cloudfoundry/cf-release
  git clone https://github.com/cloudfoundry-incubator/ducati-release
  git clone https://github.com/sykesm/garden-runc-release
popd

pushd ~/workspace/cf-release
  git checkout runtime-passed
  ./scripts/update
  bosh -n create release && bosh -n upload release
popd

pushd ~/workspace/garden-runc-release
  git checkout another-terrible-hack
  git pull
  git submodule sync
  git submodule update --init --recursive
  bosh -n create release --force && bosh -n upload release
popd

pushd ~/workspace/ducati-release
  git checkout master
  ./scripts/update
  bosh -n create release --force && bosh -n upload release
popd

pushd ~/workspace/diego-release
  git checkout release-candidate
  ./scripts/update
  bosh -n create release
  bosh upload release
popd
```

Finally, generate the manifests and deploy:

```
CF_DEPLOY=~/workspace/cf-release/bosh-lite/deployments
DIEGO_DEPLOY=~/workspace/diego-release/bosh-lite/deployments

pushd ~/workspace
  cf-release/scripts/generate-bosh-lite-dev-manifest ducati-release/manifests/cf-overrides.yml

  pushd diego-release
    scripts/generate-bosh-lite-manifests -g  # use guardian instead of garden-linux
  popd

  sed 's/\ guardian/\ garden-runc/' < $DIEGO_DEPLOY/diego.yml > $DIEGO_DEPLOY/diego.yml

  ducatify \
      --diego $DIEGO_DEPLOY/diego.yml \
      --cfCreds ducati-release/manifests/cf_creds_stub.yml \
      > $DIEGO_DEPLOY/diego_with_ducati.yml
popd

bosh -n -d $CF_DEPLOY/cf.yml deploy
bosh -n -d $DIEGO_DEPLOY/diego_with_ducati.yml deploy

bosh -d $CF_DEPLOY/cf.yml run errand acceptance_tests
bosh -d $DIEGO_DEPLOY/diego_with_ducati.yml run errand ducati-acceptance
```
