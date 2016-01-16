# glossary

## concepts

- `container`: a collection of namespaces, designed to fully isolate a process or set of processes
- `handle`: short name for a container, typically the last part of a filesystem path
- `namespace`: a Linux kernel feature that isolates processes's view of a particular type of feature
  e.g. a process in a "network namespace" sees different network resources than other processes
  ([ref](http://man7.org/linux/man-pages/man7/namespaces.7.html))
- `vxlan`: an implementation of an overlay network which encapsulates ethernet frames inside UDP packets
  the Linux kernel has vxlan support

## tools
- `ip`
  - `ip netns list`
  - `ip netns exec`
- `bridge`
  - `bridge fdb`
- `iptables`
- `ifconfig`


