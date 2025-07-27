
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

**Note: Update to Packer Plugin Installation**

With the new Packer release starting from version 1.14.0, the `packer init` command will automatically install official plugins from the [HashiCorp release site.](https://releases.hashicorp.com/)

Going forward, to use newer versions of official Packer plugins, you'll need to upgrade to Packer version 1.14.0 or later. If you're using an older version, you can still install plugins, but as a workaround, you'll need to [manually install them using the CLI.](https://developer.hashicorp.com/packer/docs/plugins/install#manually-install-plugins-using-the-cli)

There is no change to the syntax or commands for installing plugins.

### Components

#### Builders

- [qemu](/packer/integrations/hashicorp/qemu/latest/components/builder/qemu) - The QEMU builder is able to create [KVM](http://www.linux-kvm.org) virtual machine images.

