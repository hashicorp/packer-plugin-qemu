## 1.1.6 (July 16, 2026)

### IMPROVEMENTS

* Update `github.com/hashicorp/packer-plugin-sdk` from 0.6.9 to 0.6.10. [GH-248]
* Update dependencies in `go.mod`/`go.sum` including `golang.org/x/crypto` to 0.54.0.

### BUG FIXES

* Add suppression for `GO-2026-5932` after verification it only affects the unmaintained `openpgp` subpackage that is not imported by this plugin.

## 1.1.5 (June 17, 2026)

### IMPROVEMENTS

* Add support for `sd` in the `disk_interface` configuration option. [GH-239]
* Update `github.com/hashicorp/packer-plugin-sdk` from 0.6.1 to 0.6.9. [GH-224] [GH-237]
* Update dependencies in `go.mod`/`go.sum` including `github.com/ulikunitz/xz`. [GH-240]

### BUG FIXES

* Fix authorization bypass logic and related validation workflow behavior. [GH-232] [GH-233]

## 1.1.4 (July 31, 2025)

### IMPROVEMENTS:

* Change in the release process of the packer plugins binaries to releases it in the [HashiCorp official releases site](https://releases.hashicorp.com/packer-plugin-docker/).
  This change standardizes our release process and ensures a more secure and reliable pipeline for plugin delivery.


# Changelog of previous releases can be found below.

Please refer to [releases](https://github.com/hashicorp/packer-plugin-qemu/releases) for the latest CHANGELOG information.

---
## 1.0.1 (September 28, 2021)

### IMPROVEMENTS
* Add support for cd_content configuration aregument. [GH-34]
* Update packer-plugin-sdk to version 0.2.5. [GH-44]

### BUG FIXES
* Fix SCSI index conflicts when both cdrom and disk using virtio-scsi
    interface. [GH-40]

## 1.0.0 (June 14, 2021)

* Update packer-plugin-sdk to version 0.2.3. [GH-29]

## 0.0.1 (April 19, 2021)

* QEMU Plugin break out from Packer core. Changes prior to break out can be found in [Packer's CHANGELOG](https://github.com/hashicorp/packer/blob/master/CHANGELOG.md).
