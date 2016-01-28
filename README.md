# ducati-release

This release co-locates both `etcd` and `guardian`, see below for instructions

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

## Caveats

- The architecture of this whole system is shifting
- Deploy only if you know how
