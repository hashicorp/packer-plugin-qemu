# Copyright IBM Corp. 2013, 2025
# SPDX-License-Identifier: MPL-2.0

# For full specification on the configuration of this file visit:
# https://github.com/hashicorp/integration-template#metadata-configuration
integration {
  name = "QEMU"
  description = "The Qemu Packer Plugin comes with a single builder able to create KVM virtual machine images."
  identifier = "packer/hashicorp/qemu"
  component {
    type = "builder"
    name = "QEMU"
    slug = "qemu"
  }
}
