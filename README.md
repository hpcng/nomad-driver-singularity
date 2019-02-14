# nomad-driver-singularity

[Hashicorp Nomad](https://www.nomadproject.io/) driver plugin using [Singularity containers](https://github.com/sylabs/singularity) to execute tasks.

Requirements
------------

- [Nomad](https://www.nomadproject.io/downloads.html) 0.9+
- [Go](https://golang.org/doc/install) 1.11 (to build the provider plugin)
- [Singularity](https://github.com/singularityware/singularity) 3.0.3+

Building The Driver
---------------------

Clone repository to: `$GOPATH/src/github.com/sylabs/nomad-driver-singularity

```sh
$ mkdir -p $GOPATH/src/github.com/sylabs; cd $GOPATH/src/github.com/
$ git clone git@github.com:sylabs/nomad-driver-singularity
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/sylabs/nomad-driver-singularity
$ make build
```

Developing the Provider
---------------------------

If you wish to work on the driver, you'll first need [Go](http://www.golang.org) installed on your machine, and have have `singularity` installed. You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make build
```

In order to test the provider, you can simply run `make test`.

```sh
$ make test
```
