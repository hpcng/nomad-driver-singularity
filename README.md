# nomad-driver-singularity

[![GoDoc](https://godoc.org/github.com/sylabs/nomad-driver-singularity?status.svg)](https://godoc.org/github.com/sylabs/nomad-driver-singularity)
[![Build Status](https://circleci.com/gh/hpcng/nomad-driver-singularity.svg?style=shield)](https://circleci.com/gh/hpcng/workflows/nomad-driver-singularity)
[![Code Coverage](https://codecov.io/gh/hpcng/nomad-driver-singularity/branch/master/graph/badge.svg)](https://codecov.io/gh/hpcng/nomad-driver-singularity)
[![Go Report Card](https://goreportcard.com/badge/github.com/hpcng/nomad-driver-singularity)](https://goreportcard.com/report/github.com/hpcng/nomad-driver-singularity)

[Hashicorp Nomad](https://www.hashicorp.com/en/products/nomad) driver plugin using
[Singularity containers](https://github.com/sylabs/singularity) to execute tasks.

## Requirements

- [Nomad](https://github.com/hashicorp/nomad/releases) v0.9+
- [Go](https://golang.org/doc/install) v1.11+ (to build the provider plugin)
- [Singularity](https://github.com/singularityware/singularity) v3.1.0+

## Building The Driver

Clone repository on your preferred path

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

## User Guide

The Singularity Nomad driver provides an interface for using Singularity for
running application containers.

### Task Configuration

```hcl
task "lolcow" {
  driver = "Singularity"

  config {
     # this example run an image from sylabs container library with the
     # canonical example of lolcow
     image = "library://sylabsed/examples/lolcow:latest"
     # command can be run, exec or test
     command = "run"
  }
}
```

The `Singularity` driver supports the following configuration in the job spec:

- `image` - The Singularity image to run. It can be a local path or a supported URI.

  ```hcl
  config {
    image = "library://sylabsed/examples/lolcow:latest"
  }
  ```

- `verbose` - (Optional) Enables extra verbosity in the Singularity runtime logging.
  Defaults to `false`.

  ```hcl
  config {
    verbose = "false"
  }
  ```

- `debug` - (Optional) Enables extra debug output in the Singularity runtime
  logging. Defaults to `false`.

  ```hcl
  config {
    debug = "false"
  }
  ```

- `command` - Singularity command action; can be `run`, `exec` or `test`.

  ```hcl
  config {
    command = "run"
  }
  ```

- `args` - (Optional) Singularity command action arguments, when trying to pass arguments to `run`, `exec` or `test`.
  Multiple args can be given by a comma separated list.

  ```hcl
  config {
    args = [ "echo", "hello Cloud" ]
  }
  ```

- `binds` - (Optional) A user-bind path specification. This spec has the format `src[:dest[:opts]]`, where src and
  dest are outside and inside paths. If dest is not given, it is set equal to src.
  Mount options ('opts') may be specified as 'ro' (read-only) or 'rw' (read/write, which
  is the default). Multiple bind paths can be given by a comma separated list.

  ```hcl
  config {
    bind = [ "host/path:/container/path" ]
  }
  ```

- `overlay` - (Optional) Singularity command action flag, to enable an overlayFS image for persistent data
  storage or as read-only layer of container. Multiple overlay paths can be given by a comma separated list.

  ```hcl
  config {
    overlay = [ "host/path/to/overlay" ]
  }
  ```

- `security` - (Optional) Allows the root user to leverage security modules such as
  SELinux, AppArmor, and seccomp within your Singularity container.
  You can also change the UID and GID of the user within the container at runtime.

  ```hcl
  config {
    security = [ "uid:1000 " ]
  }
  ```

- `contain` - (Optional) Use minimal `/dev` and empty other directories (e.g. `/tmp` and `$HOME`) instead of sharing filesystems from your host.

  ```hcl
  config {
    contain = "false"
  }
  ```

- `workdir` - (Optional) Working directory to be used for `/tmp`, `/var/tmp` and `$HOME` (if -c/--contain was also used).

  ```hcl
  config {
    workdir = "/path/to/folder"
  }
  ```

- `pwd` - (Optional) Initial working directory for payload process inside the container.

  ```hcl
  config {
    pwd = "/path/to/folder"
  }
  ```

### Networking

Currently the `Singularity` driver only supports host networking. For more detailed instructions on how to set up networking options, please refer to the `Singularity` user guides [singularity-network](https://www.sylabs.io/guides/3.1/user-guide/networking.html).

### Client Requirements

The `Singularity` driver requires the following:

- 64-bit Linux host
- The `linux_amd64` Nomad binary
- The Singularity driver binary placed in the `plugin_dir` directory.
- [`Singularity`](https://github.com/sylabs/singularity) v3.1.1+

### Plugin Options

- `enabled` - The `Singularity` driver may be disabled on hosts by setting this option to `false` (defaults to `true`).

- `singularity_cache` - The location in which all containers are stored
  (commonly defaults to `/var/lib/singularity`). Refer to 
  [`Singularity-cache`](https://www.sylabs.io/guides/3.1/user-guide/appendix.html#c)
  for more details.

An example of using these plugin options:

```hcl
plugin "nomad-driver-Singularity" {
  config {
    enabled = true
    singularity_path = "/var/lib/singularity"
  }
}
```

Please note the plugin name should match whatever name you have specified for the external driver in the `plugin_dir` directory.

### Client Attributes

The `Singularity` driver will set the following client attributes:

- `driver.singularity` - Set to `1` if Singularity is found and enabled on the host node.
- `driver.singularity.version` - Version of `Singularity` e.g.: `3.1.0`.

### Resource Isolation

This driver supports CPU and memory isolation via the `Singularity` cgroups feature. Network
isolation is supported via `--net` and `--network` feature (Singularity v3.1.1+ required).
