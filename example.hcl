job "example" {
  datacenters = ["dc1"]
  type        = "batch"

  group "crazy-cow" {
    count = 1

    task "mooo" {
      driver = "singularity"

      config {
        image_path = "/root/hello-rootfs.ext4"
      }
    }
  }
}
