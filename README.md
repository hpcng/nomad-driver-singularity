# nomad-driver-singularity

[![GoDoc](https://godoc.org/github.com/sylabs/nomad-driver-singularity?status.svg)](https://godoc.org/github.com/sylabs/nomad-driver-singularity)
[![Build Status](https://circleci.com/gh/hpcng/nomad-driver-singularity.svg?style=shield)](https://circleci.com/gh/hpcng/workflows/nomad-driver-singularity)
[![Code Coverage](https://codecov.io/gh/sylabs/nomad-driver-singularity/branch/master/graph/badge.svg)](https://codecov.io/gh/sylabs/nomad-driver-singularity)
[![Go Report Card](https://goreportcard.com/badge/github.com/sylabs/nomad-driver-singularity)](https://goreportcard.com/report/github.com/sylabs/nomad-driver-singularity)

[Hashicorp Nomad](https://www.nomadproject.io/) driver plugin using
[Singularity containers](https://github.com/sylabs/singularity) to execute tasks.

## Requirements

- [Nomad](https://www.nomadproject.io/downloads.html) v0.9+
- [Go](https://golang.org/doc/install) v1.11+ (to build the provider plugin)
- [Singularity](https://github.com/singularityware/singularity) v3.1.0+

## Building The Driver

Clone repository on your prefered path

```sh
git clone git@github.com:sylabs/nomad-driver-singularity
```

Enter the provider directory and build the provider

```sh
cd nomad-driver-singularity
make dep
make build
```

## Developing the Provider

If you wish to contribute on the project, you'll first need [Go](http://www.golang.org)
installed on your machine, and have have `singularity` installed.

To compile the provider, run `make build`.
This will build the provider and put the task driver binary under
the NOMAD plugin dir,
which by default is located under `<nomad-data-dir>/plugins/`.

Check Nomad `-data-dir` and `-plugin-dir` flags for more information.

```sh
make build
```

In order to test the provider, you can simply run `make test`.

```sh
make test
```
