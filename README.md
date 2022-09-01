<p align="center">
  <img alt="datakit logo" src="datakit-logo.png" height="150" />
</p>

[![Slack Status](https://img.shields.io/badge/slack-join_chat-orange?logo=slack&style=plastic)](https://app.slack.com/client/T032YB4B6TA/)
[![MIT License](https://img.shields.io/badge/license-MIT-green?style=plastic)](LICENSE)

<h2>
  <a href="https://datakit.tools">Website</a>
  <span> • </span>
  <a href="https://www.guance.com/dataflux/datakit">Doc</a>
</h2>


## _Read this in other languages._
<kbd>[<img title="中文 (Simplified)" alt="中文 (Simplified)" src="https://cdn.staticaly.com/gh/hjnilsson/country-flags/master/svg/cn.svg" width="22">](README.zh_CN.md)</kbd>

DataKit is an open source, integrated data collection agent, which provides full platform (Linux/Windows/macOS) support and has comprehensive data collection capability, covering various scenarios such as host, container, middleware, tracing, logging and security inspection.

## Key Features

- Support collection of metrics, logging and tracing
- Fully support Kubernetes ecology
- [Pipeline](https://docs.guance.com/datakit/pipeline): Simple structured data extraction
- Supports third-party data import:
    - [Telegraf](https://docs.guance.com/datakit/telegraf)
    - [Prometheus](https://docs.guance.com/datakit/prom)
    - [Statsd](https://docs.guance.com/datakit/statsd)
    - [Fluentd](https://docs.guance.com/datakit/logstreaming-fluentd)
    - [Filebeats](https://docs.guance.com/datakit/beats_output)
    - [Function](https://docs.guance.com/dataflux-func/write-data-via-datakit)
    - Tracing 
        - [OpenTelemetry](https://docs.guance.com/datakit/opentelemetry)
        - [DDTrace](https://docs.guance.com/datakit/ddtrace)
        - [Zipkin](https://docs.guance.com/datakit/zipkin)
        - [Jaeger](https://docs.guance.com/datakit/jaeger)
        - [Skywalking](https://docs.guance.com/datakit/skywalking)

## Changelog

All DataKit changelog refers to [here](https://docs.guance.com/datakit/changelog).

## Minimal Requirements

| OS                                                             | Arch                | Install Path                                                                   |
| ---                                                            | ---                 | ---                                                                            |
| Linux Kernel 2.6.23+                                           | amd64/386/arm/arm64 | `/usr/local/datakit`                                                           |
| macOS 10.12+([Why](https://github.com/golang/go/issues/25633)) | amd64               | `/usr/local/datakit`                                                           |
| Windows 7+/Server 2008R2+                                      | amd64/386           | 64-bit：`C:\Program Files\datakit`<br />32-bit：`C:\Program Files(32)\datakit` |


## Install DataKit

We can directly obtain the DataKit installation command from [guance cloud](http://guance.com). Most of the installation commands seems like that:

- Linux & Mac
```shell
DK_DATAWAY="https://openway.guance.com?token=<YOUR-TOKEN>" bash -c "$(curl -L https://static.guance.com/datakit/install.sh)"
```

- Windows

```powershell
$env:DK_DATAWAY="https://openway.guance.com?token=<YOUR-TOKEN>";Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -source https://static.guance.com/datakit/install.ps1 -destination .install.ps1; powershell .install.ps1;
```

- [Kubernetes DaemonSet](https://docs.guance.com/datakit/datakit-daemonset-deploy)

For more documentations about DataKit installation, see [here](https://docs.guance.com/datakit/datakit-install).

## Build From Source

DataKit building relies on some external tools/libs, we must install them all before compile the source code.

> - **We recommend to build source on Ubuntu 20.04+**, other linux distribition may failed to install these dependencies. We do not support build DataKit on Windows.
> - Please build the project with `make`, we haven't testing with Golang/VSCode IDEs

### Setup Golang

Install and setup Golang(1.18.3+):

```shell
export GOPRIVATE=gitlab.jiagouyun.com/*
export GOPROXY=https://goproxy.cn,direct
export GOPATH=~/go            # depends on your local settings
export GOROOT=~/golang-1.18.3 # depends on your local settings
export PATH=$GOROOT/bin:~/go/bin:$PATH
```

### Install other tools

> !!! Do not install these dependencies under datakit source code dir.

- make: `apt-get install make`
- gcc: `apt-get install gcc`
- gcc-multilib: `apt-get install -y gcc-multilib`
- tree: `apt-get install tree`
- packr2: `go install github.com/gobuffalo/packr/v2/packr2@v2.8.3`
- goyacc: `go install golang.org/x/tools/cmd/goyacc@latest`
- lint related:
  - lint: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.46.2`
- eBPF related:
	- clang 10.0+: `apt-get install clang`
	- llvm 10.0+: `apt-get install llvm`
	- kernel headers
		- apt: `apt-get install -y linux-headers-$(uname -r)`
- Documentation exporting:
	- [waque 1.13.1+](https://github.com/yesmeck/waque)

### Build

1. Clone code

```shell
$ mkdir -p $GOPATH/src/gitlab.jiagouyun.com/cloudcare-tools
$ cd $GOPATH/src/gitlab.jiagouyun.com/cloudcare-tools

$ git clone https://github.com/GuanceCloud/datakit.git   # may be blocked by GFW
$ git clone https://jihulab.com/guance-cloud/datakit.git # jihulab mirror

$ cd datakit
```

2. Building

```shell
make
```

If building ok, all binaries are generated under *dist*:

```
dist
├── [4.0K]  datakit-linux-amd64
│   ├── [ 72M]  datakit
│   └── [4.0K]  externals
│       ├── [ 14M]  logfwd
│       └── [10.0M]  oracle
├── [4.0K]  local
│   ├── [ 26M]  installer-linux-amd64
│   └── [ 228]  version
└── [4.0K]  standalone
    └── [4.0K]  datakit-ebpf-linux-amd64
		        └── [ 38M]  datakit-ebpf
```

We can build all platforms(Linux/Mac/Windows) with following command:

```shell
make testing
```

## Basic Usage

We can use `help` command to see more usage of DataKit:

```shell
datakit help

# Or

./dist/datakit-linux-amd64/datakit help
```

## Contributing

Before contributing, check out some guideline of DataKit:

- Read [architecure introduciton](https://docs.guance.com/datakit/datakit-arch)
- Read [development guideline](https://docs.guance.com/datakit/development)

## Full Documentation

For full documents of DataKit, see

- [DataKit Doc](https://docs.guance.com/datakit)
