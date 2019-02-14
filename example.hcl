// Copyright (c) 2019, Sylabs Inc. All rights reserved.

job "example" {
  datacenters = ["dc1"]
  type        = "batch"

  group "crazy-cow" {
    count = 1

    task "mooo" {
      driver = "singularity"

      config {
        image_path = "/home/user/example/lolcow.sif"
      }
    }
  }
}
