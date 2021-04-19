packer {
  required_plugins {
    qemu = {
      version = ">= 0.0.1"
      source  = "github.com/hashicorp/qemu"
    }
  }
}

build {
  sources = ["source.qemu.example"]
}
