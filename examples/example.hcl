// Copyright (c) 2019, Sylabs Inc. All rights reserved.

job "example1" {
  datacenters = ["dc1"]
  type        = "batch"

  group "crazy-cow" {
    count = 1

    task "mooo" {
      driver = "singularity"

      // You can pass env vars to the runtime
      env {
        SINGULARITYENV_FOO = "var"
      }

      config {
        // For this example we are enabling debug and verbose options
        // to retrieve logs via alloc logs
        debug = true
        verbose = true
        // this example run an image from sylabs container library with the
        // canonical example of lolcow
        image = "library://sylabsed/examples/lolcow:latest"
        // command can be run, exec or test
        command = "run"
      }
    }
  }
}
