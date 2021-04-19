source "qemu" "example" {
  boot_command      = [
    "<enter><wait><f6><wait><esc><wait>",
    "<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
    "<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
    "<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
    "<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
    "<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
    "<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
    "<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
    "<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
    "<bs><bs><bs>",
    "/install/vmlinuz noapic ",
    "file=/floppy/preseed.cfg ",
    "debian-installer=en_US auto locale=en_US kbd-chooser/method=us ",
    "hostname=vagrant ",
    "fb=false debconf/frontend=noninteractive ",
    "keyboard-configuration/modelcode=SKIP ",
    "keyboard-configuration/layout=USA ",
    "keyboard-configuration/variant=USA console-setup/ask_detect=false ",
    "passwd/user-fullname=vagrant ",
    "passwd/user-password=vagrant ",
    "passwd/user-password-again=vagrant ",
    "passwd/username=vagrant ",
    "initrd=/install/initrd.gz -- <enter>"
  ]
  disk_size         = "32768"
  floppy_files      = ["http/preseed.cfg"]
  iso_urls = [
    "http://releases.ubuntu.com/16.04/ubuntu-16.04.7-server-amd64.iso"
  ]
  iso_checksum = "sha256:b23488689e16cad7a269eb2d3a3bf725d3457ee6b0868e00c8762d3816e25848"
  output_directory  = "output-ubuntu1804"
  shutdown_command  = "echo 'vagrant'|sudo -S shutdown -P now"
  ssh_password      = "vagrant"
  ssh_username      = "vagrant"
  ssh_wait_timeout  = "10000s"
  vm_name           = "ubuntu1804"
  use_default_display = true
}

build {
  sources = ["source.qemu.example"]
}
