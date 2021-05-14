# go-libp2p-quic-transport

[![](https://img.shields.io/badge/project-libp2p-yellow.svg?style=flat-square)](https://libp2p.io/)
[![Godoc Reference](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/l0k18/go-libp2p-kcp-transport)

go-libp2p-kcp-transport uses [kcp-go](https://github.com/xtaci/kcp-go) to provide KCP support for libp2p. This is partly to fill in the gap of the completely unimplemented reliable udp at [github.com/libp2p/go-udp-transport](https://github.com/libp2p/go-udp-transport).

## Install

`go-libp2p-kcp-transport` is a standard Go module which can be installed with:

```sh
go get github.com/l0k18/go-libp2p-kcp-transport
```

This repo is [gomod](https://github.com/golang/go/wiki/Modules)-compatible, and users of
Go 1.11 and later with modules enabled will automatically pull the latest tagged release
by referencing this package. Upgrades to future releases can be managed using `go get`,
or by editing your `go.mod` file as [described by the gomod documentation](https://github.com/golang/go/wiki/Modules#how-to-upgrade-and-downgrade-dependencies).

## Contribute

Feel free to join in. All welcome. Open an [issue](https://github.com/l0k18/go-libp2p-kcp-transport/issues)!
