# ducati-release

This release co-locates both `etcd` and `guardian`, see below for instructions

## Dependencies

- [etcd-release](https://github.com/cloudfoundry-incubator/etcd-release)
- [guardian-release](https://github.com/cloudfoundry-incubator/guardian-release)

## Getting started
```bash
pushd ~/workspace/etcd-release
  git pull
  bosh upload release releases/etcd/etcd-26.yml
popd

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

bosh deployment manifests/ducati-manifest.yml
bosh -n deploy
```

## Caveats

- ETCD needs to be used with caution
- The architecture of this whole system is shifting
- Deploy only if you know how
