# ducati-release

Co-locate this release with [guardian-release](https://github.com/cloudfoundry-incubator/guardian-release)

Then configure guardian to use the `ducati` external networker instead of the built-in one.

## Getting started
```bash
pushd ~/workspace/guardian-release
  git pull
  git submodule update --init --recursive
  bosh create release
  bosh upload release
popd

pushd ~/workspace/ducati-release
  git pull
  git submodule update --init --recursive
  bosh create release
  bosh upload release
popd

bosh deployment manifests/bosh-lite.yml
bosh -n deploy
```
