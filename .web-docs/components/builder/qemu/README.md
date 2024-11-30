Type: `qemu`
Artifact BuilderId: `transcend.qemu`

The Qemu Packer builder is able to create [KVM](http://www.linux-kvm.org) virtual
machine images.

The builder builds a virtual machine by creating a new virtual machine from
scratch, booting it, installing an OS, rebooting the machine with the boot media
as the virtual hard drive, provisioning software within the OS, then shutting it
down. The result of the Qemu builder is a directory containing the image file
necessary to run the virtual machine on KVM.

## Basic Example

Here is a basic example. This example is functional so long as you fixup paths
to files, URLS for ISOs and checksums.

**HCL2**

```hcl
source "qemu" "example" {
  iso_url           = "http://mirror.raystedman.net/centos/6/isos/x86_64/CentOS-6.9-x86_64-minimal.iso"
  iso_checksum      = "md5:af4a1640c0c6f348c6c41f1ea9e192a2"
  output_directory  = "output_centos_tdhtest"
  shutdown_command  = "echo 'packer' | sudo -S shutdown -P now"
  disk_size         = "5000M"
  format            = "qcow2"
  accelerator       = "kvm"
  http_directory    = "path/to/httpdir"
  ssh_username      = "root"
  ssh_password      = "s0m3password"
  ssh_timeout       = "20m"
  vm_name           = "tdhtest"
  net_device        = "virtio-net"
  disk_interface    = "virtio"
  boot_wait         = "10s"
  boot_command      = ["<tab> text ks=http://{{ .HTTPIP }}:{{ .HTTPPort }}/centos6-ks.cfg<enter><wait>"]
}

build {
  sources = ["source.qemu.example"]
}
```

**JSON**

```json
{
  "builders": [
    {
      "type": "qemu",
      "iso_url": "http://mirror.raystedman.net/centos/6/isos/x86_64/CentOS-6.9-x86_64-minimal.iso",
      "iso_checksum": "md5:af4a1640c0c6f348c6c41f1ea9e192a2",
      "output_directory": "output_centos_tdhtest",
      "shutdown_command": "echo 'packer' | sudo -S shutdown -P now",
      "disk_size": "5000M",
      "format": "qcow2",
      "accelerator": "kvm",
      "http_directory": "path/to/httpdir",
      "ssh_username": "root",
      "ssh_password": "s0m3password",
      "ssh_timeout": "20m",
      "vm_name": "tdhtest",
      "net_device": "virtio-net",
      "disk_interface": "virtio",
      "boot_wait": "10s",
      "boot_command": [
        "<tab> text ks=http://{{ .HTTPIP }}:{{ .HTTPPort }}/centos6-ks.cfg<enter><wait>"
      ]
    }
  ]
}
```

This is an example only, and will time out waiting for SSH because we have not
provided a kickstart file. You must add a valid kickstart file to the
"http_directory" and then provide the file in the "boot_command" in order for
this build to run. We recommend you check out the
[Community Templates](https://developer.hashicorp.com/packer/docs/community-tools#templates)
for a practical usage example.

Note that you will need to set `"headless": true` if you are running Packer
on a Linux server without X11; or if you are connected via SSH to a remote
Linux server and have not enabled X11 forwarding (`ssh -X`).

## Qemu Specific Configuration Reference

There are many configuration options available for the builder. In addition to
the items listed here, you will want to look at the general configuration
references for [ISO](#iso-configuration),
[HTTP](#http-directory-configuration),
[Floppy](#floppy-configuration),
[Boot](#boot-configuration),
[Shutdown](#shutdown-configuration),
[Communicator](#communicator-configuration)
configuration references, which are
necessary for this build to succeed and can be found further down the page.

### Optional:

<!-- Code generated from the comments of the Config struct in builder/qemu/config.go; DO NOT EDIT MANUALLY -->

- `iso_skip_cache` (bool) - Use iso from provided url. Qemu must support
  curl block device. This defaults to `false`.

- `accelerator` (string) - The accelerator type to use when running the VM.
  This may be `none`, `kvm`, `tcg`, `hax`, `hvf`, `whpx`, or `xen`. The appropriate
  software must have already been installed on your build machine to use the
  accelerator you specified. When no accelerator is specified, Packer will try
  to use `kvm` if it is available but will default to `tcg` otherwise.
  
  ~> The `hax` accelerator has issues attaching CDROM ISOs. This is an
  upstream issue which can be tracked
  [here](https://github.com/intel/haxm/issues/20).
  
  ~> The `hvf` and `whpx` accelerator are new and experimental as of
  [QEMU 2.12.0](https://wiki.qemu.org/ChangeLog/2.12#Host_support).
  You may encounter issues unrelated to Packer when using these.  You may need to
  add [ "-global", "virtio-pci.disable-modern=on" ] to `qemuargs` depending on the
  guest operating system.
  
  ~> HAXM is discontinued, and as of Qemu 8.0, the option is deprecated,
  please consider using another accelerator.
  
  -> As an alternative to setting `accelerator`, you can set the `machine` and `accel` args
  directly using `qemuargs`. For example, to try potential accelerators in order, you could
  use the following:
  ```hcl
  qemuargs = [
    ["-machine", "type=q35,accel=hvf:kvm:whpx:tcg"],
  ]

- `disk_additional_size` ([]string) - Additional disks to create. Uses `vm_name` as the disk name template and
  appends `-#` where `#` is the position in the array. `#` starts at 1 since 0
  is the default disk. Each string represents the disk image size in bytes.
  Optional suffixes 'k' or 'K' (kilobyte, 1024), 'M' (megabyte, 1024k), 'G'
  (gigabyte, 1024M), 'T' (terabyte, 1024G), 'P' (petabyte, 1024T) and 'E'
  (exabyte, 1024P)  are supported. 'b' is ignored. Per qemu-img documentation.
  Each additional disk uses the same disk parameters as the default disk.
  Unset by default.

- `firmware` (string) - The firmware file to be used by QEMU.
  If unset, QEMU will load its default firmware.
  Also see the QEMU documentation.
  
  NOTE: when booting in UEFI mode, please use the `efi_` (see
  [EFI Boot Configuration](#efi-boot-configuration)) options to
  setup the firmware.

- `use_pflash` (bool) - If a firmware file option was provided, this option can be
  used to change how qemu will get it.
  If false (the default), then the firmware is provided through
  the -bios option, but if true, a pflash drive will be used
  instead.
  
  NOTE: when booting in UEFI mode, please use the `efi_` (see
  [EFI Boot Configuration](#efi-boot-configuration)) options to
  setup the firmware.

- `disk_interface` (string) - The interface to use for the disk. Allowed values include any of `ide`,
  `sata`, `scsi`, `virtio` or `virtio-scsi`^\*. Note also that any boot
  commands or kickstart type scripts must have proper adjustments for
  resulting device names. The Qemu builder uses `virtio` by default.
  
  ^\* Please be aware that use of the `scsi` disk interface has been
  disabled by Red Hat due to a bug described
  [here](https://bugzilla.redhat.com/show_bug.cgi?id=1019220). If you are
  running Qemu on RHEL or a RHEL variant such as CentOS, you *must* choose
  one of the other listed interfaces. Using the `scsi` interface under
  these circumstances will cause the build to fail.

- `disk_size` (string) - The size in bytes of the hard disk of the VM. Suffix with the first
  letter of common byte types. Use "k" or "K" for kilobytes, "M" for
  megabytes, G for gigabytes, and T for terabytes. If no value is provided
  for disk_size, Packer uses a default of `40960M` (40 GB). If a disk_size
  number is provided with no units, Packer will default to Megabytes.

- `skip_resize_disk` (bool) - Packer resizes the QCOW2 image using
  qemu-img resize.  Set this option to true to disable resizing.
  Defaults to false.

- `disk_cache` (string) - The cache mode to use for disk. Allowed values include any of
  `writethrough`, `writeback`, `none`, `unsafe` or `directsync`. By
  default, this is set to `writeback`.

- `disk_discard` (string) - The discard mode to use for disk. Allowed values
  include any of unmap or ignore. By default, this is set to ignore.

- `disk_detect_zeroes` (string) - The detect-zeroes mode to use for disk.
  Allowed values include any of unmap, on or off. Defaults to off.
  When the value is "off" we don't set the flag in the qemu command, so that
  Packer still works with old versions of QEMU that don't have this option.

- `skip_compaction` (bool) - Packer compacts the QCOW2 image using
  qemu-img convert.  Set this option to true to disable compacting.
  Defaults to false.

- `disk_compression` (bool) - Apply compression to the QCOW2 disk file
  using qemu-img convert. Defaults to false.

- `format` (string) - Either `qcow2` or `raw`, this specifies the output format of the virtual
  machine image. This defaults to `qcow2`. Due to a long-standing bug with
  `qemu-img convert` on OSX, sometimes the qemu-img convert call will
  create a corrupted image. If this is an issue for you, make sure that the
  the output format matches the input file's format, and Packer will
  perform a simple copy operation instead. See
  https://bugs.launchpad.net/qemu/+bug/1776920 for more details.

- `headless` (bool) - Packer defaults to building QEMU virtual machines by
  launching a GUI that shows the console of the machine being built. When this
  value is set to `true`, the machine will start without a console.
  
  You can still see the console if you make a note of the VNC display
  number chosen, and then connect using `vncviewer -Shared <host>:<display>`

- `disk_image` (bool) - Packer defaults to building from an ISO file, this parameter controls
  whether the ISO URL supplied is actually a bootable QEMU image. When
  this value is set to `true`, the machine will either clone the source or
  use it as a backing file (if `use_backing_file` is `true`); then, it
  will resize the image according to `disk_size` and boot it.

- `use_backing_file` (bool) - Only applicable when disk_image is true
  and format is qcow2, set this option to true to create a new QCOW2
  file that uses the file located at iso_url as a backing file. The new file
  will only contain blocks that have changed compared to the backing file, so
  enabling this option can significantly reduce disk usage. If true, Packer
  will force the `skip_compaction` also to be true as well to skip disk
  conversion which would render the backing file feature useless.

- `machine_type` (string) - The type of machine emulation to use. Run your qemu binary with the
  flags `-machine help` to list available types for your system. This
  defaults to `pc`.
  
  NOTE: when booting a UEFI machine with Secure Boot enabled, this has
  to be a q35 derivative.
  If the machine is not a q35 derivative, nothing will boot (not even
  an EFI shell).

- `memory` (int) - The amount of memory to use when building the VM
  in megabytes. This defaults to 512 megabytes.

- `net_device` (string) - The driver to use for the network interface. Allowed values `ne2k_pci`,
  `i82551`, `i82557b`, `i82559er`, `rtl8139`, `e1000`, `pcnet`, `virtio`,
  `virtio-net`, `virtio-net-pci`, `usb-net`, `i82559a`, `i82559b`,
  `i82559c`, `i82550`, `i82562`, `i82557a`, `i82557c`, `i82801`,
  `vmxnet3`, `i82558a` or `i82558b`. The Qemu builder uses `virtio-net` by
  default.

- `net_bridge` (string) - Connects the network to this bridge instead of using the user mode
  networking.
  
  **NB** This bridge must already exist. You can use the `virbr0` bridge
  as created by vagrant-libvirt.
  
  **NB** This will automatically enable the QMP socket (see QMPEnable).
  
  **NB** This only works in Linux based OSes.

- `output_directory` (string) - This is the path to the directory where the
  resulting virtual machine will be created. This may be relative or absolute.
  If relative, the path is relative to the working directory when packer
  is executed. This directory must not exist or be empty prior to running
  the builder. By default this is output-BUILDNAME where "BUILDNAME" is the
  name of the build.

- `qemuargs` ([][]string) - Allows complete control over the qemu command line (though not qemu-img).
  Each array of strings makes up a command line switch
  that overrides matching default switch/value pairs. Any value specified
  as an empty string is ignored. All values after the switch are
  concatenated with no separator.
  
  ~> **Warning:** The qemu command line allows extreme flexibility, so
  beware of conflicting arguments causing failures of your run.
  For instance adding a "--drive" or "--device" override will mean that
  none of the default configuration Packer sets will be used. To see the
  defaults that Packer sets, look in your packer.log
  file (set PACKER_LOG=1 to get verbose logging) and search for the
  qemu-system-x86 command. The arguments are all printed for review, and
  you can use those arguments along with the template engines allowed
  by qemu-args to set up a working configuration that includes both the
  Packer defaults and your extra arguments.
  
  Another pitfall could be setting arguments like --no-acpi, which could
  break the ability to send power signal type commands
  (e.g., shutdown -P now) to the virtual machine, thus preventing proper
  shutdown.
  
  The following shows a sample usage:
  
  In HCL2:
  ```hcl
    qemuargs = [
      [ "-m", "1024M" ],
      [ "--no-acpi", "" ],
      [
        "-netdev",
        "user,id=mynet0,",
        "hostfwd=hostip:hostport-guestip:guestport",
        ""
      ],
      [ "-device", "virtio-net,netdev=mynet0" ]
    ]
  ```
  
  In JSON:
  ```json
    "qemuargs": [
      [ "-m", "1024M" ],
      [ "--no-acpi", "" ],
      [
        "-netdev",
        "user,id=mynet0,",
        "hostfwd=hostip:hostport-guestip:guestport",
        ""
      ],
      [ "-device", "virtio-net,netdev=mynet0" ]
    ]
  ```
  
  would produce the following (not including other defaults supplied by
  the builder and not otherwise conflicting with the qemuargs):
  
  ```text
  qemu-system-x86 -m 1024m --no-acpi -netdev
  user,id=mynet0,hostfwd=hostip:hostport-guestip:guestport -device
  virtio-net,netdev=mynet0"
  ```
  
  ~> **Windows Users:** [QEMU for Windows](https://qemu.weilnetz.de/)
  builds are available though an environmental variable does need to be
  set for QEMU for Windows to redirect stdout to the console instead of
  stdout.txt.
  
  The following shows the environment variable that needs to be set for
  Windows QEMU support:
  
  ```text
  setx SDL_STDIO_REDIRECT=0
  ```
  
  You can also use the `SSHHostPort` template variable to produce a packer
  template that can be invoked by `make` in parallel:
  
  In HCL2:
  ```hcl
    qemuargs = [
      [ "-netdev", "user,hostfwd=tcp::{{ .SSHHostPort }}-:22,id=forward"],
      [ "-device", "virtio-net,netdev=forward,id=net0"]
    ]
  ```
  
  In JSON:
  ```json
    "qemuargs": [
      [ "-netdev", "user,hostfwd=tcp::{{ .SSHHostPort }}-:22,id=forward"],
      [ "-device", "virtio-net,netdev=forward,id=net0"]
    ]
  ```
  
  `make -j 3 my-awesome-packer-templates` spawns 3 packer processes, each
  of which will bind to their own SSH port as determined by each process.
  This will also work with WinRM, just change the port forward in
  `qemuargs` to map to WinRM's default port of `5985` or whatever value
  you have the service set to listen on.
  
  This is a template engine and allows access to the following variables:
  `{{ .HTTPIP }}`, `{{ .HTTPPort }}`, `{{ .HTTPDir }}`,
  `{{ .OutputDir }}`, `{{ .Name }}`, and `{{ .SSHHostPort }}`

- `qemu_img_args` (QemuImgArgs) - A map of custom arguments to pass to qemu-img commands, where the key
  is the subcommand, and the values are lists of strings for each flag.
  Example:
  
  In HCL:
  ```hcl
  qemu_img_args {
   convert = ["-o", "preallocation=full"]
   resize  = ["-foo", "bar"]
  }
  ```
  In JSON:
  ```json
  {
   "qemu_img_args": {
     "convert": ["-o", "preallocation=full"],
  	  "resize": ["-foo", "bar"]
   }
  ```
  Please note
  that unlike qemuargs, these commands are not split into switch-value
  sub-arrays, because the basic elements in qemu-img calls are  unlikely
  to need an actual override.
  The arguments will be constructed as follows:
  - Convert:
  	Default is `qemu-img convert -O $format $sourcepath $targetpath`. Adding
  	arguments ["-foo", "bar"] to qemu_img_args.convert will change this to
  	`qemu-img convert -foo bar -O $format $sourcepath $targetpath`
  - Create:
  	Default is `create -f $format $targetpath $size`. Adding arguments
  	["-foo", "bar"] to qemu_img_args.create will change this to
  	"create -f qcow2 -foo bar target.qcow2 1234M"
  - Resize:
  	Default is `qemu-img resize -f $format $sourcepath $size`. Adding
  	arguments ["-foo", "bar"] to qemu_img_args.resize will change this to
  	`qemu-img resize -f $format -foo bar $sourcepath $size`

- `qemu_binary` (string) - The name of the Qemu binary to look for. This
  defaults to qemu-system-x86_64, but may need to be changed for
  some platforms. For example qemu-kvm, or qemu-system-i386 may be a
  better choice for some systems.

- `qmp_enable` (bool) - Enable QMP socket. Location is specified by `qmp_socket_path`. Defaults
  to false.

- `qmp_socket_path` (string) - QMP Socket Path when `qmp_enable` is true. Defaults to
  `output_directory`/`vm_name`.monitor.

- `use_default_display` (bool) - If true, do not pass a -display option
  to qemu, allowing it to choose the default. This may be needed when running
  under macOS, and getting errors about sdl not being available.

- `vga` (string) - The type of VGA card to emulate. If undefined, this will not be included
  in the command-line, and the default qemu value for the emulated machine
  will be picked.

- `display` (string) - What QEMU -display option to use. Defaults to gtk, use none to not pass the
  -display option allowing QEMU to choose the default. This may be needed when
  running under macOS, and getting errors about sdl not being available.

- `vnc_bind_address` (string) - The IP address that should be
  binded to for VNC. By default packer will use 127.0.0.1 for this. If you
  wish to bind to all interfaces use 0.0.0.0.

- `vnc_password` (string) - The password to set when VNCUsePassword == true.

- `vnc_use_password` (bool) - Whether or not to set a password on the VNC server. This option
  automatically enables the QMP socket. See `qmp_socket_path`. Defaults to
  `false`.

- `vnc_port_min` (int) - The minimum and maximum port
  to use for VNC access to the virtual machine. The builder uses VNC to type
  the initial boot_command. Because Packer generally runs in parallel,
  Packer uses a randomly chosen port in this range that appears available. By
  default this is 5900 to 6000. The minimum and maximum ports are inclusive.
  The minimum port cannot be set below 5900 due to a quirk in how QEMU parses
  vnc display address.

- `vnc_port_max` (int) - VNC Port Max

- `vm_name` (string) - This is the name of the image (QCOW2 or IMG) file for
  the new virtual machine. By default this is packer-BUILDNAME, where
  "BUILDNAME" is the name of the build. Currently, no file extension will be
  used unless it is specified in this option.

- `cdrom_interface` (string) - The interface to use for the CDROM device which contains the ISO image.
  Allowed values include any of `ide`, `scsi`, `virtio` or
  `virtio-scsi`. The Qemu builder uses `virtio` by default.
  Some ARM64 images require `virtio-scsi`.

- `vtpm` (bool) - Use a virtual (emulated) TPM device to expose to the VM.

- `use_tpm1` (bool) - Use version 1.2 of the TPM specification for the emulated TPM.
  
  By default, we use version 2.0 of the TPM specs for the emulated TPM,
  if you want to force version 1.2, set this option to true.

- `tpm_device_type` (string) - The TPM device type to inject in the qemu command-line
  
  This is required to be specified for some platforms, as the device has to
  behave differently depending on the architecture.
  
  As per the docs:
  
   * x86: tpm-tis (default)
   * ARM: tpm-tis-device
   * PPC (p-series): tpm-spapr

- `boot_steps` ([][]string) - This is an array of tuples of boot commands, to type when the virtual
  machine is booted. The first element of the tuple is the actual boot
  command. The second element of the tuple, which is optional, is a
  description of what the boot command does. This is intended to be used for
  interactive installers that requires many commands to complete the
  installation. Both the command and the description will be printed when
  logging is enabled. When debug mode is enabled Packer will pause after
  typing each boot command. This will make it easier to follow along the
  installation process and make sure the Packer and the installer are in
  sync. `boot_steps` and `boot_commands` are mutually exclusive.
  
  Example:
  
  In HCL:
  ```hcl
  boot_steps = [
    ["1<enter><wait5>", "Install NetBSD"],
    ["a<enter><wait5>", "Installation messages in English"],
    ["a<enter><wait5>", "Keyboard type: unchanged"],
  
    ["a<enter><wait5>", "Install NetBSD to hard disk"],
    ["b<enter><wait5>", "Yes"]
  ]
  ```
  
  In JSON:
  ```json
  {
    "boot_steps": [
      ["1<enter><wait5>", "Install NetBSD"],
      ["a<enter><wait5>", "Installation messages in English"],
      ["a<enter><wait5>", "Keyboard type: unchanged"],
  
      ["a<enter><wait5>", "Install NetBSD to hard disk"],
      ["b<enter><wait5>", "Yes"]
    ]
  }
  ```

- `cpu_model` (string) - The CPU model is what will be used by qemu for booting the virtual machine
  and determine which features of a specific model/family of processors
  is supported.
  
  Any string is supported here, and to see which models are supported
  by qemu for a specific architecture, refer to `qemu -cpu help`
  
  The default value here is that no cpu option will be passed through to qemu,
  therefore it will default to whichever CPU model is the default for the
  targetted system (on x86_64 for example, it will be qemu64)
  
  NOTE: RHEL9 removed support for `qemu64` in their distributed qemu package,
  forcing users of RHEL9 on x86_64 systems to define this. "host" is a
  reasonable value if using an hypervisor.

<!-- End of code generated from the comments of the Config struct in builder/qemu/config.go; -->


## ISO Configuration

<!-- Code generated from the comments of the ISOConfig struct in multistep/commonsteps/iso_config.go; DO NOT EDIT MANUALLY -->

By default, Packer will symlink, download or copy image files to the Packer
cache into a "`hash($iso_url+$iso_checksum).$iso_target_extension`" file.
Packer uses [hashicorp/go-getter](https://github.com/hashicorp/go-getter) in
file mode in order to perform a download.

go-getter supports the following protocols:

* Local files
* Git
* Mercurial
* HTTP
* Amazon S3

Examples:
go-getter can guess the checksum type based on `iso_checksum` length, and it is
also possible to specify the checksum type.

In JSON:

```json

	"iso_checksum": "946a6077af6f5f95a51f82fdc44051c7aa19f9cfc5f737954845a6050543d7c2",
	"iso_url": "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"

```

```json

	"iso_checksum": "file:ubuntu.org/..../ubuntu-14.04.1-server-amd64.iso.sum",
	"iso_url": "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"

```

```json

	"iso_checksum": "file://./shasums.txt",
	"iso_url": "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"

```

```json

	"iso_checksum": "file:./shasums.txt",
	"iso_url": "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"

```

In HCL2:

```hcl

	iso_checksum = "946a6077af6f5f95a51f82fdc44051c7aa19f9cfc5f737954845a6050543d7c2"
	iso_url = "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"

```

```hcl

	iso_checksum = "file:ubuntu.org/..../ubuntu-14.04.1-server-amd64.iso.sum"
	iso_url = "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"

```

```hcl

	iso_checksum = "file://./shasums.txt"
	iso_url = "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"

```

```hcl

	iso_checksum = "file:./shasums.txt",
	iso_url = "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"

```

<!-- End of code generated from the comments of the ISOConfig struct in multistep/commonsteps/iso_config.go; -->


### Required:

<!-- Code generated from the comments of the ISOConfig struct in multistep/commonsteps/iso_config.go; DO NOT EDIT MANUALLY -->

- `iso_checksum` (string) - The checksum for the ISO file or virtual hard drive file. The type of
  the checksum is specified within the checksum field as a prefix, ex:
  "md5:{$checksum}". The type of the checksum can also be omitted and
  Packer will try to infer it based on string length. Valid values are
  "none", "{$checksum}", "md5:{$checksum}", "sha1:{$checksum}",
  "sha256:{$checksum}", "sha512:{$checksum}" or "file:{$path}". Here is a
  list of valid checksum values:
   * md5:090992ba9fd140077b0661cb75f7ce13
   * 090992ba9fd140077b0661cb75f7ce13
   * sha1:ebfb681885ddf1234c18094a45bbeafd91467911
   * ebfb681885ddf1234c18094a45bbeafd91467911
   * sha256:ed363350696a726b7932db864dda019bd2017365c9e299627830f06954643f93
   * ed363350696a726b7932db864dda019bd2017365c9e299627830f06954643f93
   * file:http://releases.ubuntu.com/20.04/SHA256SUMS
   * file:file://./local/path/file.sum
   * file:./local/path/file.sum
   * none
  Although the checksum will not be verified when it is set to "none",
  this is not recommended since these files can be very large and
  corruption does happen from time to time.

- `iso_url` (string) - A URL to the ISO containing the installation image or virtual hard drive
  (VHD or VHDX) file to clone.

<!-- End of code generated from the comments of the ISOConfig struct in multistep/commonsteps/iso_config.go; -->


### Optional:

<!-- Code generated from the comments of the ISOConfig struct in multistep/commonsteps/iso_config.go; DO NOT EDIT MANUALLY -->

- `iso_urls` ([]string) - Multiple URLs for the ISO to download. Packer will try these in order.
  If anything goes wrong attempting to download or while downloading a
  single URL, it will move on to the next. All URLs must point to the same
  file (same checksum). By default this is empty and `iso_url` is used.
  Only one of `iso_url` or `iso_urls` can be specified.

- `iso_target_path` (string) - The path where the iso should be saved after download. By default will
  go in the packer cache, with a hash of the original filename and
  checksum as its name.

- `iso_target_extension` (string) - The extension of the iso file after download. This defaults to `iso`.

<!-- End of code generated from the comments of the ISOConfig struct in multistep/commonsteps/iso_config.go; -->


## Http directory configuration

<!-- Code generated from the comments of the HTTPConfig struct in multistep/commonsteps/http_config.go; DO NOT EDIT MANUALLY -->

Packer will create an http server serving `http_directory` when it is set, a
random free port will be selected and the architecture of the directory
referenced will be available in your builder.

Example usage from a builder:

```
wget http://{{ .HTTPIP }}:{{ .HTTPPort }}/foo/bar/preseed.cfg
```

<!-- End of code generated from the comments of the HTTPConfig struct in multistep/commonsteps/http_config.go; -->


### Optional:

<!-- Code generated from the comments of the HTTPConfig struct in multistep/commonsteps/http_config.go; DO NOT EDIT MANUALLY -->

- `http_directory` (string) - Path to a directory to serve using an HTTP server. The files in this
  directory will be available over HTTP that will be requestable from the
  virtual machine. This is useful for hosting kickstart files and so on.
  By default this is an empty string, which means no HTTP server will be
  started. The address and port of the HTTP server will be available as
  variables in `boot_command`. This is covered in more detail below.

- `http_content` (map[string]string) - Key/Values to serve using an HTTP server. `http_content` works like and
  conflicts with `http_directory`. The keys represent the paths and the
  values contents, the keys must start with a slash, ex: `/path/to/file`.
  `http_content` is useful for hosting kickstart files and so on. By
  default this is empty, which means no HTTP server will be started. The
  address and port of the HTTP server will be available as variables in
  `boot_command`. This is covered in more detail below.
  Example:
  ```hcl
    http_content = {
      "/a/b"     = file("http/b")
      "/foo/bar" = templatefile("${path.root}/preseed.cfg", { packages = ["nginx"] })
    }
  ```

- `http_port_min` (int) - These are the minimum and maximum port to use for the HTTP server
  started to serve the `http_directory`. Because Packer often runs in
  parallel, Packer will choose a randomly available port in this range to
  run the HTTP server. If you want to force the HTTP server to be on one
  port, make this minimum and maximum port the same. By default the values
  are `8000` and `9000`, respectively.

- `http_port_max` (int) - HTTP Port Max

- `http_bind_address` (string) - This is the bind address for the HTTP server. Defaults to 0.0.0.0 so that
  it will work with any network interface.

<!-- End of code generated from the comments of the HTTPConfig struct in multistep/commonsteps/http_config.go; -->


## Floppy configuration

<!-- Code generated from the comments of the FloppyConfig struct in multistep/commonsteps/floppy_config.go; DO NOT EDIT MANUALLY -->

A floppy can be made available for your build. This is most useful for
unattended Windows installs, which look for an Autounattend.xml file on
removable media. By default, no floppy will be attached. All files listed in
this setting get placed into the root directory of the floppy and the floppy
is attached as the first floppy device. The summary size of the listed files
must not exceed 1.44 MB. The supported ways to move large files into the OS
are using `http_directory` or [the file
provisioner](/packer/docs/provisioner/file).

<!-- End of code generated from the comments of the FloppyConfig struct in multistep/commonsteps/floppy_config.go; -->


### Optional:

<!-- Code generated from the comments of the FloppyConfig struct in multistep/commonsteps/floppy_config.go; DO NOT EDIT MANUALLY -->

- `floppy_files` ([]string) - A list of files to place onto a floppy disk that is attached when the VM
  is booted. Currently, no support exists for creating sub-directories on
  the floppy. Wildcard characters (\\*, ?, and \[\]) are allowed. Directory
  names are also allowed, which will add all the files found in the
  directory to the floppy.

- `floppy_dirs` ([]string) - A list of directories to place onto the floppy disk recursively. This is
  similar to the `floppy_files` option except that the directory structure
  is preserved. This is useful for when your floppy disk includes drivers
  or if you just want to organize it's contents as a hierarchy. Wildcard
  characters (\\*, ?, and \[\]) are allowed. The maximum summary size of
  all files in the listed directories are the same as in `floppy_files`.

- `floppy_content` (map[string]string) - Key/Values to add to the floppy disk. The keys represent the paths, and
  the values contents. It can be used alongside `floppy_files` or
  `floppy_dirs`, which is useful to add large files without loading them
  into memory. If any paths are specified by both, the contents in
  `floppy_content` will take precedence.
  
  Usage example (HCL):
  
  ```hcl
  floppy_files = ["vendor-data"]
  floppy_content = {
    "meta-data" = jsonencode(local.instance_data)
    "user-data" = templatefile("user-data", { packages = ["nginx"] })
  }
  floppy_label = "cidata"
  ```

- `floppy_label` (string) - Floppy Label

<!-- End of code generated from the comments of the FloppyConfig struct in multistep/commonsteps/floppy_config.go; -->


### CD configuration

<!-- Code generated from the comments of the CDConfig struct in multistep/commonsteps/extra_iso_config.go; DO NOT EDIT MANUALLY -->

An iso (CD) containing custom files can be made available for your build.

By default, no extra CD will be attached. All files listed in this setting
get placed into the root directory of the CD and the CD is attached as the
second CD device.

This config exists to work around modern operating systems that have no
way to mount floppy disks, which was our previous go-to for adding files at
boot time.

<!-- End of code generated from the comments of the CDConfig struct in multistep/commonsteps/extra_iso_config.go; -->


#### Optional:

<!-- Code generated from the comments of the CDConfig struct in multistep/commonsteps/extra_iso_config.go; DO NOT EDIT MANUALLY -->

- `cd_files` ([]string) - A list of files to place onto a CD that is attached when the VM is
  booted. This can include either files or directories; any directories
  will be copied onto the CD recursively, preserving directory structure
  hierarchy. Symlinks will have the link's target copied into the directory
  tree on the CD where the symlink was. File globbing is allowed.
  
  Usage example (JSON):
  
  ```json
  "cd_files": ["./somedirectory/meta-data", "./somedirectory/user-data"],
  "cd_label": "cidata",
  ```
  
  Usage example (HCL):
  
  ```hcl
  cd_files = ["./somedirectory/meta-data", "./somedirectory/user-data"]
  cd_label = "cidata"
  ```
  
  The above will create a CD with two files, user-data and meta-data in the
  CD root. This specific example is how you would create a CD that can be
  used for an Ubuntu 20.04 autoinstall.
  
  Since globbing is also supported,
  
  ```hcl
  cd_files = ["./somedirectory/*"]
  cd_label = "cidata"
  ```
  
  Would also be an acceptable way to define the above cd. The difference
  between providing the directory with or without the glob is whether the
  directory itself or its contents will be at the CD root.
  
  Use of this option assumes that you have a command line tool installed
  that can handle the iso creation. Packer will use one of the following
  tools:
  
    * xorriso
    * mkisofs
    * hdiutil (normally found in macOS)
    * oscdimg (normally found in Windows as part of the Windows ADK)

- `cd_content` (map[string]string) - Key/Values to add to the CD. The keys represent the paths, and the values
  contents. It can be used alongside `cd_files`, which is useful to add large
  files without loading them into memory. If any paths are specified by both,
  the contents in `cd_content` will take precedence.
  
  Usage example (HCL):
  
  ```hcl
  cd_files = ["vendor-data"]
  cd_content = {
    "meta-data" = jsonencode(local.instance_data)
    "user-data" = templatefile("user-data", { packages = ["nginx"] })
  }
  cd_label = "cidata"
  ```

- `cd_label` (string) - CD Label

<!-- End of code generated from the comments of the CDConfig struct in multistep/commonsteps/extra_iso_config.go; -->


## Shutdown configuration

### Optional:

<!-- Code generated from the comments of the ShutdownConfig struct in shutdowncommand/config.go; DO NOT EDIT MANUALLY -->

- `shutdown_command` (string) - The command to use to gracefully shut down the machine once all
  provisioning is complete. By default this is an empty string, which
  tells Packer to just forcefully shut down the machine. This setting can
  be safely omitted if for example, a shutdown command to gracefully halt
  the machine is configured inside a provisioning script. If one or more
  scripts require a reboot it is suggested to leave this blank (since
  reboots may fail) and instead specify the final shutdown command in your
  last script.

- `shutdown_timeout` (duration string | ex: "1h5m2s") - The amount of time to wait after executing the shutdown_command for the
  virtual machine to actually shut down. If the machine doesn't shut down
  in this time it is considered an error. By default, the time out is "5m"
  (five minutes).

<!-- End of code generated from the comments of the ShutdownConfig struct in shutdowncommand/config.go; -->


## Communicator configuration

### Optional common fields:

<!-- Code generated from the comments of the Config struct in communicator/config.go; DO NOT EDIT MANUALLY -->

- `communicator` (string) - Packer currently supports three kinds of communicators:
  
  -   `none` - No communicator will be used. If this is set, most
      provisioners also can't be used.
  
  -   `ssh` - An SSH connection will be established to the machine. This
      is usually the default.
  
  -   `winrm` - A WinRM connection will be established.
  
  In addition to the above, some builders have custom communicators they
  can use. For example, the Docker builder has a "docker" communicator
  that uses `docker exec` and `docker cp` to execute scripts and copy
  files.

- `pause_before_connecting` (duration string | ex: "1h5m2s") - We recommend that you enable SSH or WinRM as the very last step in your
  guest's bootstrap script, but sometimes you may have a race condition
  where you need Packer to wait before attempting to connect to your
  guest.
  
  If you end up in this situation, you can use the template option
  `pause_before_connecting`. By default, there is no pause. For example if
  you set `pause_before_connecting` to `10m` Packer will check whether it
  can connect, as normal. But once a connection attempt is successful, it
  will disconnect and then wait 10 minutes before connecting to the guest
  and beginning provisioning.

<!-- End of code generated from the comments of the Config struct in communicator/config.go; -->


<!-- Code generated from the comments of the CommConfig struct in builder/qemu/comm_config.go; DO NOT EDIT MANUALLY -->

- `host_port_min` (int) - The minimum port to use for the Communicator port on the host machine which is forwarded
  to the SSH or WinRM port on the guest machine. By default this is 2222.

- `host_port_max` (int) - The maximum port to use for the Communicator port on the host machine which is forwarded
  to the SSH or WinRM port on the guest machine. Because Packer often runs in parallel,
  Packer will choose a randomly available port in this range to use as the
  host port. By default this is 4444.

- `skip_nat_mapping` (bool) - Defaults to false. When enabled, Packer
  does not setup forwarded port mapping for communicator (SSH or WinRM) requests and uses ssh_port or winrm_port
  on the host to communicate to the virtual machine.

<!-- End of code generated from the comments of the CommConfig struct in builder/qemu/comm_config.go; -->


### Optional SSH fields:

<!-- Code generated from the comments of the SSH struct in communicator/config.go; DO NOT EDIT MANUALLY -->

- `ssh_host` (string) - The address to SSH to. This usually is automatically configured by the
  builder.

- `ssh_port` (int) - The port to connect to SSH. This defaults to `22`.

- `ssh_username` (string) - The username to connect to SSH with. Required if using SSH.

- `ssh_password` (string) - A plaintext password to use to authenticate with SSH.

- `ssh_ciphers` ([]string) - This overrides the value of ciphers supported by default by Golang.
  The default value is [
    "aes128-gcm@openssh.com",
    "chacha20-poly1305@openssh.com",
    "aes128-ctr", "aes192-ctr", "aes256-ctr",
  ]
  
  Valid options for ciphers include:
  "aes128-ctr", "aes192-ctr", "aes256-ctr", "aes128-gcm@openssh.com",
  "chacha20-poly1305@openssh.com",
  "arcfour256", "arcfour128", "arcfour", "aes128-cbc", "3des-cbc",

- `ssh_clear_authorized_keys` (bool) - If true, Packer will attempt to remove its temporary key from
  `~/.ssh/authorized_keys` and `/root/.ssh/authorized_keys`. This is a
  mostly cosmetic option, since Packer will delete the temporary private
  key from the host system regardless of whether this is set to true
  (unless the user has set the `-debug` flag). Defaults to "false";
  currently only works on guests with `sed` installed.

- `ssh_key_exchange_algorithms` ([]string) - If set, Packer will override the value of key exchange (kex) algorithms
  supported by default by Golang. Acceptable values include:
  "curve25519-sha256@libssh.org", "ecdh-sha2-nistp256",
  "ecdh-sha2-nistp384", "ecdh-sha2-nistp521",
  "diffie-hellman-group14-sha1", and "diffie-hellman-group1-sha1".

- `ssh_certificate_file` (string) - Path to user certificate used to authenticate with SSH.
  The `~` can be used in path and will be expanded to the
  home directory of current user.

- `ssh_pty` (bool) - If `true`, a PTY will be requested for the SSH connection. This defaults
  to `false`.

- `ssh_timeout` (duration string | ex: "1h5m2s") - The time to wait for SSH to become available. Packer uses this to
  determine when the machine has booted so this is usually quite long.
  Example value: `10m`.
  This defaults to `5m`, unless `ssh_handshake_attempts` is set.

- `ssh_disable_agent_forwarding` (bool) - If true, SSH agent forwarding will be disabled. Defaults to `false`.

- `ssh_handshake_attempts` (int) - The number of handshakes to attempt with SSH once it can connect.
  This defaults to `10`, unless a `ssh_timeout` is set.

- `ssh_bastion_host` (string) - A bastion host to use for the actual SSH connection.

- `ssh_bastion_port` (int) - The port of the bastion host. Defaults to `22`.

- `ssh_bastion_agent_auth` (bool) - If `true`, the local SSH agent will be used to authenticate with the
  bastion host. Defaults to `false`.

- `ssh_bastion_username` (string) - The username to connect to the bastion host.

- `ssh_bastion_password` (string) - The password to use to authenticate with the bastion host.

- `ssh_bastion_interactive` (bool) - If `true`, the keyboard-interactive used to authenticate with bastion host.

- `ssh_bastion_private_key_file` (string) - Path to a PEM encoded private key file to use to authenticate with the
  bastion host. The `~` can be used in path and will be expanded to the
  home directory of current user.

- `ssh_bastion_certificate_file` (string) - Path to user certificate used to authenticate with bastion host.
  The `~` can be used in path and will be expanded to the
  home directory of current user.

- `ssh_file_transfer_method` (string) - `scp` or `sftp` - How to transfer files, Secure copy (default) or SSH
  File Transfer Protocol.
  
  **NOTE**: Guests using Windows with Win32-OpenSSH v9.1.0.0p1-Beta, scp
  (the default protocol for copying data) returns a a non-zero error code since the MOTW
  cannot be set, which cause any file transfer to fail. As a workaround you can override the transfer protocol
  with SFTP instead `ssh_file_transfer_method = "sftp"`.

- `ssh_proxy_host` (string) - A SOCKS proxy host to use for SSH connection

- `ssh_proxy_port` (int) - A port of the SOCKS proxy. Defaults to `1080`.

- `ssh_proxy_username` (string) - The optional username to authenticate with the proxy server.

- `ssh_proxy_password` (string) - The optional password to use to authenticate with the proxy server.

- `ssh_keep_alive_interval` (duration string | ex: "1h5m2s") - How often to send "keep alive" messages to the server. Set to a negative
  value (`-1s`) to disable. Example value: `10s`. Defaults to `5s`.

- `ssh_read_write_timeout` (duration string | ex: "1h5m2s") - The amount of time to wait for a remote command to end. This might be
  useful if, for example, packer hangs on a connection after a reboot.
  Example: `5m`. Disabled by default.

- `ssh_remote_tunnels` ([]string) - 

- `ssh_local_tunnels` ([]string) - 

<!-- End of code generated from the comments of the SSH struct in communicator/config.go; -->


- `ssh_private_key_file` (string) - Path to a PEM encoded private key file to use to authenticate with SSH.
  The `~` can be used in path and will be expanded to the home directory
  of current user.


<!-- Code generated from the comments of the SSHTemporaryKeyPair struct in communicator/config.go; DO NOT EDIT MANUALLY -->

- `temporary_key_pair_type` (string) - `dsa` | `ecdsa` | `ed25519` | `rsa` ( the default )
  
  Specifies the type of key to create. The possible values are 'dsa',
  'ecdsa', 'ed25519', or 'rsa'.
  
  NOTE: DSA is deprecated and no longer recognized as secure, please
  consider other alternatives like RSA or ED25519.

- `temporary_key_pair_bits` (int) - Specifies the number of bits in the key to create. For RSA keys, the
  minimum size is 1024 bits and the default is 4096 bits. Generally, 3072
  bits is considered sufficient. DSA keys must be exactly 1024 bits as
  specified by FIPS 186-2. For ECDSA keys, bits determines the key length
  by selecting from one of three elliptic curve sizes: 256, 384 or 521
  bits. Attempting to use bit lengths other than these three values for
  ECDSA keys will fail. Ed25519 keys have a fixed length and bits will be
  ignored.
  
  NOTE: DSA is deprecated and no longer recognized as secure as specified
  by FIPS 186-5, please consider other alternatives like RSA or ED25519.

<!-- End of code generated from the comments of the SSHTemporaryKeyPair struct in communicator/config.go; -->


### Optional WinRM fields:

<!-- Code generated from the comments of the WinRM struct in communicator/config.go; DO NOT EDIT MANUALLY -->

- `winrm_username` (string) - The username to use to connect to WinRM.

- `winrm_password` (string) - The password to use to connect to WinRM.

- `winrm_host` (string) - The address for WinRM to connect to.
  
  NOTE: If using an Amazon EBS builder, you can specify the interface
  WinRM connects to via
  [`ssh_interface`](/packer/integrations/hashicorp/amazon/latest/components/builder/ebs#ssh_interface)

- `winrm_no_proxy` (bool) - Setting this to `true` adds the remote
  `host:port` to the `NO_PROXY` environment variable. This has the effect of
  bypassing any configured proxies when connecting to the remote host.
  Default to `false`.

- `winrm_port` (int) - The WinRM port to connect to. This defaults to `5985` for plain
  unencrypted connection and `5986` for SSL when `winrm_use_ssl` is set to
  true.

- `winrm_timeout` (duration string | ex: "1h5m2s") - The amount of time to wait for WinRM to become available. This defaults
  to `30m` since setting up a Windows machine generally takes a long time.

- `winrm_use_ssl` (bool) - If `true`, use HTTPS for WinRM.

- `winrm_insecure` (bool) - If `true`, do not check server certificate chain and host name.

- `winrm_use_ntlm` (bool) - If `true`, NTLMv2 authentication (with session security) will be used
  for WinRM, rather than default (basic authentication), removing the
  requirement for basic authentication to be enabled within the target
  guest. Further reading for remote connection authentication can be found
  [here](https://msdn.microsoft.com/en-us/library/aa384295(v=vs.85).aspx).

<!-- End of code generated from the comments of the WinRM struct in communicator/config.go; -->


## Boot Configuration

<!-- Code generated from the comments of the VNCConfig struct in bootcommand/config.go; DO NOT EDIT MANUALLY -->

The boot command "typed" character for character over a VNC connection to
the machine, simulating a human actually typing the keyboard.

Keystrokes are typed as separate key up/down events over VNC with a default
100ms delay. The delay alleviates issues with latency and CPU contention.
You can tune this delay on a per-builder basis by specifying
"boot_key_interval" in your Packer template.

<!-- End of code generated from the comments of the VNCConfig struct in bootcommand/config.go; -->


<!-- Code generated from the comments of the BootConfig struct in bootcommand/config.go; DO NOT EDIT MANUALLY -->

The boot configuration is very important: `boot_command` specifies the keys
to type when the virtual machine is first booted in order to start the OS
installer. This command is typed after boot_wait, which gives the virtual
machine some time to actually load.

The boot_command is an array of strings. The strings are all typed in
sequence. It is an array only to improve readability within the template.

There are a set of special keys available. If these are in your boot
command, they will be replaced by the proper key:

-   `<bs>` - Backspace

-   `<del>` - Delete

-   `<enter> <return>` - Simulates an actual "enter" or "return" keypress.

-   `<esc>` - Simulates pressing the escape key.

-   `<tab>` - Simulates pressing the tab key.

-   `<f1> - <f12>` - Simulates pressing a function key.

-   `<up> <down> <left> <right>` - Simulates pressing an arrow key.

-   `<spacebar>` - Simulates pressing the spacebar.

-   `<insert>` - Simulates pressing the insert key.

-   `<home> <end>` - Simulates pressing the home and end keys.

  - `<pageUp> <pageDown>` - Simulates pressing the page up and page down
    keys.

-   `<menu>` - Simulates pressing the Menu key.

-   `<leftAlt> <rightAlt>` - Simulates pressing the alt key.

-   `<leftCtrl> <rightCtrl>` - Simulates pressing the ctrl key.

-   `<leftShift> <rightShift>` - Simulates pressing the shift key.

-   `<leftSuper> <rightSuper>` - Simulates pressing the ⌘ or Windows key.

  - `<wait> <wait5> <wait10>` - Adds a 1, 5 or 10 second pause before
    sending any additional keys. This is useful if you have to generally
    wait for the UI to update before typing more.

  - `<waitXX>` - Add an arbitrary pause before sending any additional keys.
    The format of `XX` is a sequence of positive decimal numbers, each with
    optional fraction and a unit suffix, such as `300ms`, `1.5h` or `2h45m`.
    Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`. For
    example `<wait10m>` or `<wait1m20s>`.

  - `<XXXOn> <XXXOff>` - Any printable keyboard character, and of these
    "special" expressions, with the exception of the `<wait>` types, can
    also be toggled on or off. For example, to simulate ctrl+c, use
    `<leftCtrlOn>c<leftCtrlOff>`. Be sure to release them, otherwise they
    will be held down until the machine reboots. To hold the `c` key down,
    you would use `<cOn>`. Likewise, `<cOff>` to release.

  - `{{ .HTTPIP }} {{ .HTTPPort }}` - The IP and port, respectively of an
    HTTP server that is started serving the directory specified by the
    `http_directory` configuration parameter. If `http_directory` isn't
    specified, these will be blank!

-   `{{ .Name }}` - The name of the VM.

Example boot command. This is actually a working boot command used to start an
CentOS 6.4 installer:

In JSON:

```json
"boot_command": [

	   "<tab><wait>",
	   " ks=http://{{ .HTTPIP }}:{{ .HTTPPort }}/centos6-ks.cfg<enter>"
	]

```

In HCL2:

```hcl
boot_command = [

	   "<tab><wait>",
	   " ks=http://{{ .HTTPIP }}:{{ .HTTPPort }}/centos6-ks.cfg<enter>"
	]

```

The example shown below is a working boot command used to start an Ubuntu
12.04 installer:

In JSON:

```json
"boot_command": [

	"<esc><esc><enter><wait>",
	"/install/vmlinuz noapic ",
	"preseed/url=http://{{ .HTTPIP }}:{{ .HTTPPort }}/preseed.cfg ",
	"debian-installer=en_US auto locale=en_US kbd-chooser/method=us ",
	"hostname={{ .Name }} ",
	"fb=false debconf/frontend=noninteractive ",
	"keyboard-configuration/modelcode=SKIP keyboard-configuration/layout=USA ",
	"keyboard-configuration/variant=USA console-setup/ask_detect=false ",
	"initrd=/install/initrd.gz -- <enter>"

]
```

In HCL2:

```hcl
boot_command = [

	"<esc><esc><enter><wait>",
	"/install/vmlinuz noapic ",
	"preseed/url=http://{{ .HTTPIP }}:{{ .HTTPPort }}/preseed.cfg ",
	"debian-installer=en_US auto locale=en_US kbd-chooser/method=us ",
	"hostname={{ .Name }} ",
	"fb=false debconf/frontend=noninteractive ",
	"keyboard-configuration/modelcode=SKIP keyboard-configuration/layout=USA ",
	"keyboard-configuration/variant=USA console-setup/ask_detect=false ",
	"initrd=/install/initrd.gz -- <enter>"

]
```

For more examples of various boot commands, see the sample projects from our
[community templates page](https://packer.io/community-tools#templates).

<!-- End of code generated from the comments of the BootConfig struct in bootcommand/config.go; -->


### Optional:

<!-- Code generated from the comments of the VNCConfig struct in bootcommand/config.go; DO NOT EDIT MANUALLY -->

- `disable_vnc` (bool) - Whether to create a VNC connection or not. A boot_command cannot be used
  when this is true. Defaults to false.

- `boot_key_interval` (duration string | ex: "1h5m2s") - Time in ms to wait between each key press

<!-- End of code generated from the comments of the VNCConfig struct in bootcommand/config.go; -->


<!-- Code generated from the comments of the BootConfig struct in bootcommand/config.go; DO NOT EDIT MANUALLY -->

- `boot_keygroup_interval` (duration string | ex: "1h5m2s") - Time to wait after sending a group of key pressses. The value of this
  should be a duration. Examples are `5s` and `1m30s` which will cause
  Packer to wait five seconds and one minute 30 seconds, respectively. If
  this isn't specified, a sensible default value is picked depending on
  the builder type.

- `boot_wait` (duration string | ex: "1h5m2s") - The time to wait after booting the initial virtual machine before typing
  the `boot_command`. The value of this should be a duration. Examples are
  `5s` and `1m30s` which will cause Packer to wait five seconds and one
  minute 30 seconds, respectively. If this isn't specified, the default is
  `10s` or 10 seconds. To set boot_wait to 0s, use a negative number, such
  as "-1s"

- `boot_command` ([]string) - This is an array of commands to type when the virtual machine is first
  booted. The goal of these commands should be to type just enough to
  initialize the operating system installer. Special keys can be typed as
  well, and are covered in the section below on the boot command. If this
  is not specified, it is assumed the installer will start itself.

<!-- End of code generated from the comments of the BootConfig struct in bootcommand/config.go; -->


## EFI Boot Configuration

<!-- Code generated from the comments of the QemuEFIBootConfig struct in builder/qemu/config.go; DO NOT EDIT MANUALLY -->

Booting in EFI mode

Use these options if wanting to boot on a UEFI firmware, as the options to
do so are different from what BIOS (default) booting will require.

<!-- End of code generated from the comments of the QemuEFIBootConfig struct in builder/qemu/config.go; -->


### Optional

<!-- Code generated from the comments of the QemuEFIBootConfig struct in builder/qemu/config.go; DO NOT EDIT MANUALLY -->

- `efi_boot` (bool) - Boot in EFI mode instead of BIOS. This is required for more modern
  guest OS. If either or both of `efi_firmware_code` or
  `efi_firmware_vars` are defined, this will implicitely be set to `true`.
  
  NOTE: when using a Secure-Boot enabled firmware, the machine type has
  to be q35, otherwise qemu will not boot.

- `efi_firmware_code` (string) - Path to the CODE part of OVMF (or other compatible firmwares)
  The OVMF_CODE.fd file contains the bootstrap code for booting in EFI
  mode, and requires a separate VARS.fd file to be able to persist data
  between boot cycles.
  
  Default: `/usr/share/OVMF/OVMF_CODE.fd`

- `efi_firmware_vars` (string) - Path to the VARS corresponding to the OVMF code file.
  
  Default: `/usr/share/OVMF/OVMF_VARS.fd`

- `efi_drop_efivars` (bool) - Drop the efivars.fd file in the exported artifact.
  
  In addition to the disks created by the builder, we also expose the
  `efivars.fd` file if the image was booted with UEFI enabled.
  
  However, if the output is consumed by a post-processor (like AWS,
  GCP, etc.), this may not be supported by the code, and since the file
  is not a disk image, this will error.
  This option can then be used to remove the `efivars.fd` from the
  artifact produced by the builder, so it only lists the disks produced
  instead.

<!-- End of code generated from the comments of the QemuEFIBootConfig struct in builder/qemu/config.go; -->


## SMP Configuration

<!-- Code generated from the comments of the QemuSMPConfig struct in builder/qemu/config.go; DO NOT EDIT MANUALLY -->

QemuSMPConfig sets the smp configuration option for the Qemu command-line

The smp option sets the number of vCPUs to expose to the VM, the final
number of available vCPUs is `sockets * cores * threads`.

<!-- End of code generated from the comments of the QemuSMPConfig struct in builder/qemu/config.go; -->


### Optional

<!-- Code generated from the comments of the QemuSMPConfig struct in builder/qemu/config.go; DO NOT EDIT MANUALLY -->

- `cpus` (int) - The number of virtual cpus to use when building the VM.
  
  If undefined, the value will either be `1`, or the product of
  `sockets * cores * threads`
  
  If this is defined in conjunction with any topology specifier (sockets,
  cores and/or threads), the smallest of the two will be used.
  
  If the cpu count is the only thing specified, qemu's default behaviour
  regarding topology will be applied.
  The behaviour depends on the version of qemu; before version 6.2, sockets
  were preferred to cores, from version 6.2 onwards, cores are preferred.

- `sockets` (int) - The number of sockets to use when building the VM.
   The default is `1` socket.
   The socket count must not be higher than the CPU count.

- `cores` (int) - The number of cores per CPU to use when building the VM.
   The default is `1` core per CPU.

- `threads` (int) - The number of threads per core to use when building the VM.
   The default is `1` thread per core.

<!-- End of code generated from the comments of the QemuSMPConfig struct in builder/qemu/config.go; -->


### Communicator Configuration

#### Optional:

<!-- Code generated from the comments of the Config struct in communicator/config.go; DO NOT EDIT MANUALLY -->

- `communicator` (string) - Packer currently supports three kinds of communicators:
  
  -   `none` - No communicator will be used. If this is set, most
      provisioners also can't be used.
  
  -   `ssh` - An SSH connection will be established to the machine. This
      is usually the default.
  
  -   `winrm` - A WinRM connection will be established.
  
  In addition to the above, some builders have custom communicators they
  can use. For example, the Docker builder has a "docker" communicator
  that uses `docker exec` and `docker cp` to execute scripts and copy
  files.

- `pause_before_connecting` (duration string | ex: "1h5m2s") - We recommend that you enable SSH or WinRM as the very last step in your
  guest's bootstrap script, but sometimes you may have a race condition
  where you need Packer to wait before attempting to connect to your
  guest.
  
  If you end up in this situation, you can use the template option
  `pause_before_connecting`. By default, there is no pause. For example if
  you set `pause_before_connecting` to `10m` Packer will check whether it
  can connect, as normal. But once a connection attempt is successful, it
  will disconnect and then wait 10 minutes before connecting to the guest
  and beginning provisioning.

<!-- End of code generated from the comments of the Config struct in communicator/config.go; -->


### SSH key pair automation

The QEMU builder can inject the current SSH key pair's public key into
the template using the `SSHPublicKey` template engine. This is the SSH public
key as a line in OpenSSH authorized_keys format.

When a private key is provided using `ssh_private_key_file`, the key's
corresponding public key can be accessed using the above engine.

- `ssh_private_key_file` (string) - Path to a PEM encoded private key file to use to authenticate with SSH.
  The `~` can be used in path and will be expanded to the home directory
  of current user.


If `ssh_password` and `ssh_private_key_file` are not specified, Packer will
automatically generate en ephemeral key pair. The key pair's public key can
be accessed using the template engine.

For example, the public key can be provided in the boot command as a URL
encoded string by appending `| urlquery` to the variable:

In JSON:

```json
"boot_command": [
  "<up><wait><tab> text ks=http://{{ .HTTPIP }}:{{ .HTTPPort }}/ks.cfg PACKER_USER={{ user `username` }} PACKER_AUTHORIZED_KEY={{ .SSHPublicKey | urlquery }}<enter>"
]
```

In HCL2:

```hcl
boot_command = [
  "<up><wait><tab> text ks=http://{{ .HTTPIP }}:{{ .HTTPPort }}/ks.cfg PACKER_USER={{ user `username` }} PACKER_AUTHORIZED_KEY={{ .SSHPublicKey | urlquery }}<enter>"
]
```

A kickstart could then leverage those fields from the kernel command line by
decoding the URL-encoded public key:

```shell
%post

# Newly created users need the file/folder framework for SSH key authentication.
umask 0077
mkdir /etc/skel/.ssh
touch /etc/skel/.ssh/authorized_keys

# Loop over the command line. Set interesting variables.
for x in $(cat /proc/cmdline)
do
  case $x in
    PACKER_USER=*)
      PACKER_USER="${x#*=}"
      ;;
    PACKER_AUTHORIZED_KEY=*)
      # URL decode $encoded into $PACKER_AUTHORIZED_KEY
      encoded=$(echo "${x#*=}" | tr '+' ' ')
      printf -v PACKER_AUTHORIZED_KEY '%b' "${encoded//%/\\x}"
      ;;
  esac
done

# Create/configure packer user, if any.
if [ -n "$PACKER_USER" ]
then
  useradd $PACKER_USER
  echo "%$PACKER_USER ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers.d/$PACKER_USER
  [ -n "$PACKER_AUTHORIZED_KEY" ] && echo $PACKER_AUTHORIZED_KEY >> $(eval echo ~"$PACKER_USER")/.ssh/authorized_keys
fi

%end
```

### Troubleshooting

#### Invalid Keymaps

Some users have experienced errors complaining about invalid keymaps. This
seems to be related to having a `common` directory or file in the directory
they've run Packer in, like the Packer source directory. This appears to be an
upstream bug with qemu, and the best solution for now is to remove the
file/directory or run in another directory.

Some users have reported issues with incorrect keymaps using qemu version 2.11.
This is a bug with qemu, and the solution is to upgrade, or downgrade to 2.10.1
or earlier.

#### Corrupted image after Packer calls qemu-img convert on OSX

Due to an upstream bug with `qemu-img convert` on OSX, sometimes the qemu-img
convert call will create a corrupted image. If this is an issue for you, make
sure that the the output format (provided using the option `format`) matches
the input file's format and file extension, and Packer will
perform a simple copy operation instead. You will also want to set
`"skip_compaction": true,` and `"disk_compression": false` to skip a final
image conversion at the end of the build. See
https://bugs.launchpad.net/qemu/+bug/1776920 for more details.
