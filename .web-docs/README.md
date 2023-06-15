
The Qemu Packer Plugin comes with a single builder able to create KVM virtual machine images.


### Installation

To install this plugin add this code into your Packer configuration and run [packer init](/packer/docs/commands/init)

```hcl
packer {
  required_plugins {
    qemu = {
      version = "~> 1"
      source  = "github.com/hashicorp/qemu"
    }
  }
}
```
Alternatively, you can use `packer plugins install` to manage installation of this plugin.

```sh
packer plugins install github.com/hashicorp/qemu
```

### Components

#### Builders

- [qemu](/packer/integrations/hashicorp/qemu/latest/components/builder/qemu) - The QEMU builder is able to create [KVM](http://www.linux-kvm.org) virtual machine images.

