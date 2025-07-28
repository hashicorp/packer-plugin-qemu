# Packer Plugin Qemu
The `Qemu` multi-component plugin can be used with HashiCorp [Packer](https://www.packer.io)
to create custom images. For the full list of available features for this plugin see [docs](docs).

## Installation

### Using pre-built releases

#### Using the `packer init` command

Starting from version 1.7, Packer supports a new `packer init` command allowing
automatic installation of Packer plugins. Read the
[Packer documentation](https://www.packer.io/docs/commands/init) for more information.

To install this plugin, copy and paste this code into your Packer configuration .
Then, run [`packer init`](https://www.packer.io/docs/commands/init).

```hcl
packer {
  required_plugins {
    qemu = {
      version = ">= 1.1.3"
      source  = "github.com/hashicorp/qemu"
    }
  }
}
```

**Note: Update to Packer Plugin Installation**

With the new Packer release starting from version 1.14.0, the `packer init` command will automatically install official plugins from the [HashiCorp release site.](https://releases.hashicorp.com/)

Going forward, to use newer versions of official Packer plugins, you'll need to upgrade to Packer version 1.14.0 or later. If you're using an older version, you can still install plugins, but as a workaround, you'll need to [manually install them using the CLI.](https://developer.hashicorp.com/packer/docs/plugins/install#manually-install-plugins-using-the-cli)

There is no change to the syntax or commands for installing plugins.

#### Manual installation

You can find pre-built binary releases of the plugin [here](https://github.com/hashicorp/packer-plugin-qemu/releases).
Once you have downloaded the latest archive corresponding to your target OS,
uncompress it to retrieve the plugin binary file corresponding to your platform.
To install the plugin, please follow the Packer documentation on
[installing a plugin](https://www.packer.io/docs/extending/plugins/#installing-plugins).


### From Sources

If you prefer to build the plugin from sources, clone the GitHub repository
locally and run the command `go build` from the root
directory. Upon successful compilation, a `packer-plugin-qemu` plugin
binary file can be found in the root directory.
To install the compiled plugin, please follow the official Packer documentation
on [installing a plugin](https://www.packer.io/docs/extending/plugins/#installing-plugins).


### Configuration

For more information on how to configure the plugin, please read the
documentation located in the [`docs/`](docs) directory.


## Contributing

* If you think you've found a bug in the code or you have a question regarding
  the usage of this software, please reach out to us by opening an issue in
  this GitHub repository.
* Contributions to this project are welcome: if you want to add a feature or a
  fix a bug, please do so by opening a Pull Request in this GitHub repository.
  In case of feature contribution, we kindly ask you to open an issue to
  discuss it beforehand.
