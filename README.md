# ducati-release

This release should be deployed so that the `ducati` job co-locates with the `garden` job from guardian-release.  See below.

## Dependencies

- [guardian-release](https://github.com/cloudfoundry-incubator/guardian-release)

## Getting started

### Clone the dependencies
```bash
pushd ~/workspace
  git clone https://github.com/cloudfoundry-incubator/ducati-release
  git clone https://github.com/cloudfoundry-incubator/guardian-release
popd
```

### Deploy the releases
```bash
bosh target lite
pushd ~/workspace/guardian-release
  git pull
  git submodule update --init --recursive
  bosh create release
  bosh upload release
popd

pushd ~/workspace/ducati-release
  git pull
  git submodule update --init --recursive
  bosh create release --force && bosh -n upload release
  bosh deployment manifests/ducati-manifest.yml
popd

bosh -n deploy
```
