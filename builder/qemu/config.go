//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type Config,QemuImgArgs

package qemu

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/bootcommand"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/shutdowncommand"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

var accels = map[string]struct{}{
	"none": {},
	"kvm":  {},
	"tcg":  {},
	"xen":  {},
	"hax":  {},
	"hvf":  {},
	"whpx": {},
}

var diskInterface = map[string]bool{
	"ide":         true,
	"sata":        true,
	"scsi":        true,
	"virtio":      true,
	"virtio-scsi": true,
}

var diskCache = map[string]bool{
	"writethrough": true,
	"writeback":    true,
	"none":         true,
	"unsafe":       true,
	"directsync":   true,
}

var diskDiscard = map[string]bool{
	"unmap":  true,
	"ignore": true,
}

var diskDZeroes = map[string]bool{
	"unmap": true,
	"on":    true,
	"off":   true,
}

type QemuImgArgs struct {
	Convert []string `mapstructure:"convert" required:"false"`
	Create  []string `mapstructure:"create" required:"false"`
	Resize  []string `mapstructure:"resize" required:"false"`
}

// QemuSMPConfig sets the smp configuration option for the Qemu command-line
//
// The smp option sets the number of vCPUs to expose to the VM, the final
// number of available vCPUs is sockets*cores*threads.
type QemuSMPConfig struct {
	// The number of virtual cpus to use when building the VM.
	//
	// If undefined, the value will either be `1`, or the product of
	// `sockets * cpus * threads`
	//
	// If this is defined in conjunction with any topology specifier (sockets,
	// cores and/or threads), the smallest of the two will be used.
	//
	// If the cpu count is the only thing specified, qemu's default behaviour
	// regarding topology will be applied.
	// The behaviour depends on the version of qemu; before version 6.2, sockets
	// were preferred to cores, from version 6.2 onwards, cores are preferred.
	CpuCount int `mapstructure:"cpus" required:"false"`
	// The number of sockets to use when building the VM.
	//  The default is `1` socket.
	//  The socket count must not be higher than the CPU count.
	SocketCount int `mapstructure:"sockets" required:"false"`
	// The number of cores per CPU to use when building the VM.
	//  The default is `1` core per CPU.
	CoreCount int `mapstructure:"cores" required:"false"`
	// The number of threads per core to use when building the VM.
	//  The default is `1` thread per core.
	ThreadCount int `mapstructure:"threads" required:"false"`
}

// getDefaultCmdLine returns the smp command-line argument if defined.
func (c QemuSMPConfig) getDefaultCmdLine() string {
	totalVCpus := c.getMaxCPUs()
	if totalVCpus == 0 && c.CpuCount == 0 {
		return "1"
	}

	cpuCount := c.CpuCount

	if cpuCount == 0 {
		log.Printf("CPU count at default value, setting to topology maximum: %d", totalVCpus)
		cpuCount = totalVCpus
	}

	if totalVCpus > cpuCount {
		log.Print("CPU count lower than what's described in topology." +
			"This will negatively impact performance.")
	}

	if cpuCount > totalVCpus && totalVCpus != 0 {
		log.Printf("CPU count is greater than what topology allows, setting to max CPU count of the provided topology: %d", totalVCpus)
		cpuCount = totalVCpus
	}

	smpStr := fmt.Sprintf("%d", cpuCount)
	if c.SocketCount > 0 {
		smpStr = fmt.Sprintf("%s,sockets=%d", smpStr, c.SocketCount)
	}
	if c.CoreCount > 0 {
		smpStr = fmt.Sprintf("%s,cores=%d", smpStr, c.CoreCount)
	}
	if c.ThreadCount > 0 {
		smpStr = fmt.Sprintf("%s,threads=%d", smpStr, c.ThreadCount)
	}

	return smpStr
}

// getMaxCPUs infers the maximum number of CPUs compatible with the optional topology
//
// If nothing is defined, 0 will be returned, otherwise this is the product of
// all non-zero terms in `sockets`, `cores` and `threads`.
func (c QemuSMPConfig) getMaxCPUs() int {
	totalVCPUs := c.SocketCount

	if c.CoreCount > 0 && totalVCPUs > 0 {
		totalVCPUs *= c.CoreCount
	}

	// If number of sockets were not provided take the number of cores
	if totalVCPUs == 0 {
		totalVCPUs = c.CoreCount
	}

	if c.ThreadCount > 0 && totalVCPUs != 0 {
		totalVCPUs *= c.ThreadCount
	}

	// If nothing else was provided, return the thread count
	if totalVCPUs == 0 {
		totalVCPUs = c.ThreadCount
	}

	return totalVCPUs
}

// Booting in EFI mode
//
// Use these options if wanting to boot on a UEFI firmware, as the options to
// do so are different from what BIOS (default) booting will require.
type QemuEFIBootConfig struct {
	// Boot in EFI mode instead of BIOS. This is required for more modern
	// guest OS. If either or both of `efi_firmware_code` or
	// `efi_firmware_vars` are defined, this will implicitely be set to `true`.
	//
	// NOTE: when using a Secure-Boot enabled firmware, the machine type has
	// to be q35, otherwise qemu will not boot.
	EnableEFI bool `mapstructure:"efi_boot" required:"false"`
	// Path to the CODE part of OVMF (or other compatible firmwares)
	// The OVMF_CODE.fd file contains the bootstrap code for booting in EFI
	// mode, and requires a separate VARS.fd file to be able to persist data
	// between boot cycles.
	//
	// Default: /usr/share/OVMF/OVMF_CODE.fd
	OVMFCode string `mapstructure:"efi_firmware_code" required:"false"`
	// Path to the VARS corresponding to the OVMF code file.
	//
	// Default: /usr/share/OVMF/OVMF_VARS.fd
	OVMFVars string `mapstructure:"efi_firmware_vars" required:"false"`
}

func (efiCfg *QemuEFIBootConfig) loadDefaults() {
	// Auto enable EFI if either of the Code/Vars path is set
	if efiCfg.OVMFCode != "" || efiCfg.OVMFVars != "" {
		efiCfg.EnableEFI = true
	}

	if !efiCfg.EnableEFI {
		return
	}

	if efiCfg.OVMFCode == "" {
		efiCfg.OVMFCode = "/usr/share/OVMF/OVMF_CODE.fd"
	}

	if efiCfg.OVMFVars == "" {
		efiCfg.OVMFVars = "/usr/share/OVMF/OVMF_VARS.fd"
	}
}

type Config struct {
	common.PackerConfig            `mapstructure:",squash"`
	commonsteps.HTTPConfig         `mapstructure:",squash"`
	commonsteps.ISOConfig          `mapstructure:",squash"`
	bootcommand.VNCConfig          `mapstructure:",squash"`
	shutdowncommand.ShutdownConfig `mapstructure:",squash"`
	CommConfig                     CommConfig `mapstructure:",squash"`
	commonsteps.FloppyConfig       `mapstructure:",squash"`
	commonsteps.CDConfig           `mapstructure:",squash"`
	QemuSMPConfig                  `mapstructure:",squash"`
	QemuEFIBootConfig              `mapstructure:",squash"`
	// Use iso from provided url. Qemu must support
	// curl block device. This defaults to `false`.
	ISOSkipCache bool `mapstructure:"iso_skip_cache" required:"false"`
	// The accelerator type to use when running the VM.
	// This may be `none`, `kvm`, `tcg`, `hax`, `hvf`, `whpx`, or `xen`. The appropriate
	// software must have already been installed on your build machine to use the
	// accelerator you specified. When no accelerator is specified, Packer will try
	// to use `kvm` if it is available but will default to `tcg` otherwise.
	//
	// ~> The `hax` accelerator has issues attaching CDROM ISOs. This is an
	// upstream issue which can be tracked
	// [here](https://github.com/intel/haxm/issues/20).
	//
	// ~> The `hvf` and `whpx` accelerator are new and experimental as of
	// [QEMU 2.12.0](https://wiki.qemu.org/ChangeLog/2.12#Host_support).
	// You may encounter issues unrelated to Packer when using these.  You may need to
	// add [ "-global", "virtio-pci.disable-modern=on" ] to `qemuargs` depending on the
	// guest operating system.
	//
	// ~> For `whpx`, note that [Stefan Weil's QEMU for Windows distribution](https://qemu.weilnetz.de/w64/)
	// does not include WHPX support and users may need to compile or source a
	// build of QEMU for Windows themselves with WHPX support.
	Accelerator string `mapstructure:"accelerator" required:"false"`
	// Additional disks to create. Uses `vm_name` as the disk name template and
	// appends `-#` where `#` is the position in the array. `#` starts at 1 since 0
	// is the default disk. Each string represents the disk image size in bytes.
	// Optional suffixes 'k' or 'K' (kilobyte, 1024), 'M' (megabyte, 1024k), 'G'
	// (gigabyte, 1024M), 'T' (terabyte, 1024G), 'P' (petabyte, 1024T) and 'E'
	// (exabyte, 1024P)  are supported. 'b' is ignored. Per qemu-img documentation.
	// Each additional disk uses the same disk parameters as the default disk.
	// Unset by default.
	AdditionalDiskSize []string `mapstructure:"disk_additional_size" required:"false"`
	// The firmware file to be used by QEMU.
	// If unset, QEMU will load its default firmware.
	// Also see the QEMU documentation.
	//
	// NOTE: when booting in UEFI mode, please use the `efi_` options to
	// setup the firmware.
	Firmware string `mapstructure:"firmware" required:"false"`
	// If a firmware file option was provided, this option can be
	// used to change how qemu will get it.
	// If false (the default), then the firmware is provided through
	// the -bios option, but if true, a pflash drive will be used
	// instead.
	//
	// NOTE: when booting in UEFI mode, please use the `efi_` options to
	// setup the firmware.
	PFlash bool `mapstructure:"use_pflash" required:"false"`
	// The interface to use for the disk. Allowed values include any of `ide`,
	// `sata`, `scsi`, `virtio` or `virtio-scsi`^\*. Note also that any boot
	// commands or kickstart type scripts must have proper adjustments for
	// resulting device names. The Qemu builder uses `virtio` by default.
	//
	// ^\* Please be aware that use of the `scsi` disk interface has been
	// disabled by Red Hat due to a bug described
	// [here](https://bugzilla.redhat.com/show_bug.cgi?id=1019220). If you are
	// running Qemu on RHEL or a RHEL variant such as CentOS, you *must* choose
	// one of the other listed interfaces. Using the `scsi` interface under
	// these circumstances will cause the build to fail.
	DiskInterface string `mapstructure:"disk_interface" required:"false"`
	// The size in bytes of the hard disk of the VM. Suffix with the first
	// letter of common byte types. Use "k" or "K" for kilobytes, "M" for
	// megabytes, G for gigabytes, and T for terabytes. If no value is provided
	// for disk_size, Packer uses a default of `40960M` (40 GB). If a disk_size
	// number is provided with no units, Packer will default to Megabytes.
	DiskSize string `mapstructure:"disk_size" required:"false"`
	// Packer resizes the QCOW2 image using
	// qemu-img resize.  Set this option to true to disable resizing.
	// Defaults to false.
	SkipResizeDisk bool `mapstructure:"skip_resize_disk" required:"false"`
	// The cache mode to use for disk. Allowed values include any of
	// `writethrough`, `writeback`, `none`, `unsafe` or `directsync`. By
	// default, this is set to `writeback`.
	DiskCache string `mapstructure:"disk_cache" required:"false"`
	// The discard mode to use for disk. Allowed values
	// include any of unmap or ignore. By default, this is set to ignore.
	DiskDiscard string `mapstructure:"disk_discard" required:"false"`
	// The detect-zeroes mode to use for disk.
	// Allowed values include any of unmap, on or off. Defaults to off.
	// When the value is "off" we don't set the flag in the qemu command, so that
	// Packer still works with old versions of QEMU that don't have this option.
	DetectZeroes string `mapstructure:"disk_detect_zeroes" required:"false"`
	// Packer compacts the QCOW2 image using
	// qemu-img convert.  Set this option to true to disable compacting.
	// Defaults to false.
	SkipCompaction bool `mapstructure:"skip_compaction" required:"false"`
	// Apply compression to the QCOW2 disk file
	// using qemu-img convert. Defaults to false.
	DiskCompression bool `mapstructure:"disk_compression" required:"false"`
	// Either `qcow2` or `raw`, this specifies the output format of the virtual
	// machine image. This defaults to `qcow2`. Due to a long-standing bug with
	// `qemu-img convert` on OSX, sometimes the qemu-img convert call will
	// create a corrupted image. If this is an issue for you, make sure that the
	// the output format matches the input file's format, and Packer will
	// perform a simple copy operation instead. See
	// https://bugs.launchpad.net/qemu/+bug/1776920 for more details.
	Format string `mapstructure:"format" required:"false"`
	// Packer defaults to building QEMU virtual machines by
	// launching a GUI that shows the console of the machine being built. When this
	// value is set to `true`, the machine will start without a console.
	//
	// You can still see the console if you make a note of the VNC display
	// number chosen, and then connect using `vncviewer -Shared <host>:<display>`
	Headless bool `mapstructure:"headless" required:"false"`
	// Packer defaults to building from an ISO file, this parameter controls
	// whether the ISO URL supplied is actually a bootable QEMU image. When
	// this value is set to `true`, the machine will either clone the source or
	// use it as a backing file (if `use_backing_file` is `true`); then, it
	// will resize the image according to `disk_size` and boot it.
	DiskImage bool `mapstructure:"disk_image" required:"false"`
	// Only applicable when disk_image is true
	// and format is qcow2, set this option to true to create a new QCOW2
	// file that uses the file located at iso_url as a backing file. The new file
	// will only contain blocks that have changed compared to the backing file, so
	// enabling this option can significantly reduce disk usage. If true, Packer
	// will force the `skip_compaction` also to be true as well to skip disk
	// conversion which would render the backing file feature useless.
	UseBackingFile bool `mapstructure:"use_backing_file" required:"false"`
	// The type of machine emulation to use. Run your qemu binary with the
	// flags `-machine help` to list available types for your system. This
	// defaults to `pc`.
	//
	// NOTE: when booting a UEFI machine with Secure Boot enabled, this has
	// to be a q35 derivative.
	// If the machine is not a q35 derivative, nothing will boot (not even
	// an EFI shell).
	MachineType string `mapstructure:"machine_type" required:"false"`
	// The amount of memory to use when building the VM
	// in megabytes. This defaults to 512 megabytes.
	MemorySize int `mapstructure:"memory" required:"false"`
	// The driver to use for the network interface. Allowed values `ne2k_pci`,
	// `i82551`, `i82557b`, `i82559er`, `rtl8139`, `e1000`, `pcnet`, `virtio`,
	// `virtio-net`, `virtio-net-pci`, `usb-net`, `i82559a`, `i82559b`,
	// `i82559c`, `i82550`, `i82562`, `i82557a`, `i82557c`, `i82801`,
	// `vmxnet3`, `i82558a` or `i82558b`. The Qemu builder uses `virtio-net` by
	// default.
	NetDevice string `mapstructure:"net_device" required:"false"`
	// Connects the network to this bridge instead of using the user mode
	// networking.
	//
	// **NB** This bridge must already exist. You can use the `virbr0` bridge
	// as created by vagrant-libvirt.
	//
	// **NB** This will automatically enable the QMP socket (see QMPEnable).
	//
	// **NB** This only works in Linux based OSes.
	NetBridge string `mapstructure:"net_bridge" required:"false"`
	// This is the path to the directory where the
	// resulting virtual machine will be created. This may be relative or absolute.
	// If relative, the path is relative to the working directory when packer
	// is executed. This directory must not exist or be empty prior to running
	// the builder. By default this is output-BUILDNAME where "BUILDNAME" is the
	// name of the build.
	OutputDir string `mapstructure:"output_directory" required:"false"`
	// Allows complete control over the qemu command line (though not qemu-img).
	// Each array of strings makes up a command line switch
	// that overrides matching default switch/value pairs. Any value specified
	// as an empty string is ignored. All values after the switch are
	// concatenated with no separator.
	//
	// ~> **Warning:** The qemu command line allows extreme flexibility, so
	// beware of conflicting arguments causing failures of your run.
	// For instance adding a "--drive" or "--device" override will mean that
	// none of the default configuration Packer sets will be used. To see the
	// defaults that Packer sets, look in your packer.log
	// file (set PACKER_LOG=1 to get verbose logging) and search for the
	// qemu-system-x86 command. The arguments are all printed for review, and
	// you can use those arguments along with the template engines allowed
	// by qemu-args to set up a working configuration that includes both the
	// Packer defaults and your extra arguments.
	//
	// Another pitfall could be setting arguments like --no-acpi, which could
	// break the ability to send power signal type commands
	// (e.g., shutdown -P now) to the virtual machine, thus preventing proper
	// shutdown.
	//
	// The following shows a sample usage:
	//
	// In JSON:
	// ```json
	//   "qemuargs": [
	//     [ "-m", "1024M" ],
	//     [ "--no-acpi", "" ],
	//     [
	//       "-netdev",
	//       "user,id=mynet0,",
	//       "hostfwd=hostip:hostport-guestip:guestport",
	//       ""
	//     ],
	//     [ "-device", "virtio-net,netdev=mynet0" ]
	//   ]
	// ```
	//
	// In HCL2:
	// ```hcl
	//   qemuargs = [
	//     [ "-m", "1024M" ],
	//     [ "--no-acpi", "" ],
	//     [
	//       "-netdev",
	//       "user,id=mynet0,",
	//       "hostfwd=hostip:hostport-guestip:guestport",
	//       ""
	//     ],
	//     [ "-device", "virtio-net,netdev=mynet0" ]
	//   ]
	// ```
	//
	// would produce the following (not including other defaults supplied by
	// the builder and not otherwise conflicting with the qemuargs):
	//
	// ```text
	// qemu-system-x86 -m 1024m --no-acpi -netdev
	// user,id=mynet0,hostfwd=hostip:hostport-guestip:guestport -device
	// virtio-net,netdev=mynet0"
	// ```
	//
	// ~> **Windows Users:** [QEMU for Windows](https://qemu.weilnetz.de/)
	// builds are available though an environmental variable does need to be
	// set for QEMU for Windows to redirect stdout to the console instead of
	// stdout.txt.
	//
	// The following shows the environment variable that needs to be set for
	// Windows QEMU support:
	//
	// ```text
	// setx SDL_STDIO_REDIRECT=0
	// ```
	//
	// You can also use the `SSHHostPort` template variable to produce a packer
	// template that can be invoked by `make` in parallel:
	//
	// In JSON:
	// ```json
	//   "qemuargs": [
	//     [ "-netdev", "user,hostfwd=tcp::{{ .SSHHostPort }}-:22,id=forward"],
	//     [ "-device", "virtio-net,netdev=forward,id=net0"]
	//   ]
	// ```
	//
	// In HCL2:
	// ```hcl
	//   qemuargs = [
	//     [ "-netdev", "user,hostfwd=tcp::{{ .SSHHostPort }}-:22,id=forward"],
	//     [ "-device", "virtio-net,netdev=forward,id=net0"]
	//   ]
	//
	// `make -j 3 my-awesome-packer-templates` spawns 3 packer processes, each
	// of which will bind to their own SSH port as determined by each process.
	// This will also work with WinRM, just change the port forward in
	// `qemuargs` to map to WinRM's default port of `5985` or whatever value
	// you have the service set to listen on.
	//
	// This is a template engine and allows access to the following variables:
	// `{{ .HTTPIP }}`, `{{ .HTTPPort }}`, `{{ .HTTPDir }}`,
	// `{{ .OutputDir }}`, `{{ .Name }}`, and `{{ .SSHHostPort }}`
	QemuArgs [][]string `mapstructure:"qemuargs" required:"false"`
	// A map of custom arguments to pass to qemu-img commands, where the key
	// is the subcommand, and the values are lists of strings for each flag.
	// Example:
	//
	// In HCL:
	// ```hcl
	//qemu_img_args {
	//  convert = ["-o", "preallocation=full"]
	//  resize  = ["-foo", "bar"]
	//}
	// ```
	// In JSON:
	// ```json
	// {
	//  "qemu_img_args": {
	//    "convert": ["-o", "preallocation=full"],
	//	  "resize": ["-foo", "bar"]
	//  }
	// ```
	// Please note
	// that unlike qemuargs, these commands are not split into switch-value
	// sub-arrays, because the basic elements in qemu-img calls are  unlikely
	// to need an actual override.
	// The arguments will be constructed as follows:
	// - Convert:
	// 	Default is `qemu-img convert -O $format $sourcepath $targetpath`. Adding
	// 	arguments ["-foo", "bar"] to qemu_img_args.convert will change this to
	// 	`qemu-img convert -foo bar -O $format $sourcepath $targetpath`
	// - Create:
	// 	Default is `create -f $format $targetpath $size`. Adding arguments
	// 	["-foo", "bar"] to qemu_img_args.create will change this to
	// 	"create -f qcow2 -foo bar target.qcow2 1234M"
	// - Resize:
	// 	Default is `qemu-img resize -f $format $sourcepath $size`. Adding
	// 	arguments ["-foo", "bar"] to qemu_img_args.resize will change this to
	// 	`qemu-img resize -f $format -foo bar $sourcepath $size`
	QemuImgArgs QemuImgArgs `mapstructure:"qemu_img_args" required:"false"`
	// The name of the Qemu binary to look for. This
	// defaults to qemu-system-x86_64, but may need to be changed for
	// some platforms. For example qemu-kvm, or qemu-system-i386 may be a
	// better choice for some systems.
	QemuBinary string `mapstructure:"qemu_binary" required:"false"`
	// Enable QMP socket. Location is specified by `qmp_socket_path`. Defaults
	// to false.
	QMPEnable bool `mapstructure:"qmp_enable" required:"false"`
	// QMP Socket Path when `qmp_enable` is true. Defaults to
	// `output_directory`/`vm_name`.monitor.
	QMPSocketPath string `mapstructure:"qmp_socket_path" required:"false"`
	// If true, do not pass a -display option
	// to qemu, allowing it to choose the default. This may be needed when running
	// under macOS, and getting errors about sdl not being available.
	UseDefaultDisplay bool `mapstructure:"use_default_display" required:"false"`
	// What QEMU -display option to use. Defaults to gtk, use none to not pass the
	// -display option allowing QEMU to choose the default. This may be needed when
	// running under macOS, and getting errors about sdl not being available.
	Display string `mapstructure:"display" required:"false"`
	// The IP address that should be
	// binded to for VNC. By default packer will use 127.0.0.1 for this. If you
	// wish to bind to all interfaces use 0.0.0.0.
	VNCBindAddress string `mapstructure:"vnc_bind_address" required:"false"`
	// Whether or not to set a password on the VNC server. This option
	// automatically enables the QMP socket. See `qmp_socket_path`. Defaults to
	// `false`.
	VNCUsePassword bool `mapstructure:"vnc_use_password" required:"false"`
	// The minimum and maximum port
	// to use for VNC access to the virtual machine. The builder uses VNC to type
	// the initial boot_command. Because Packer generally runs in parallel,
	// Packer uses a randomly chosen port in this range that appears available. By
	// default this is 5900 to 6000. The minimum and maximum ports are inclusive.
	// The minimum port cannot be set below 5900 due to a quirk in how QEMU parses
	// vnc display address.
	VNCPortMin int `mapstructure:"vnc_port_min" required:"false"`
	VNCPortMax int `mapstructure:"vnc_port_max"`
	// This is the name of the image (QCOW2 or IMG) file for
	// the new virtual machine. By default this is packer-BUILDNAME, where
	// "BUILDNAME" is the name of the build. Currently, no file extension will be
	// used unless it is specified in this option.
	VMName string `mapstructure:"vm_name" required:"false"`
	// The interface to use for the CDROM device which contains the ISO image.
	// Allowed values include any of `ide`, `scsi`, `virtio` or
	// `virtio-scsi`. The Qemu builder uses `virtio` by default.
	// Some ARM64 images require `virtio-scsi`.
	CDROMInterface string `mapstructure:"cdrom_interface" required:"false"`
	// Use a virtual (emulated) TPM device to expose to the VM.
	VTPM bool `mapstructure:"vtpm" required:"false"`
	// Use version 1.2 of the TPM specification for the emulated TPM.
	//
	// By default, we use version 2.0 of the TPM specs for the emulated TPM,
	// if you want to force version 1.2, set this option to true.
	VTPMUseTPM1 bool `mapstructure:"use_tpm1" required:"false"`
	// The TPM device type to inject in the qemu command-line
	//
	// This is required to be specified for some platforms, as the device has to
	// behave differently depending on the architecture.
	//
	// As per the docs:
	//
	//  * x86: tpm-tis (default)
	//  * ARM: tpm-tis-device
	//  * PPC (p-series): tpm-spapr
	TPMType string `mapstructure:"tpm_device_type" required:"false"`
	// TODO(mitchellh): deprecate
	RunOnce bool `mapstructure:"run_once"`

	ctx interpolate.Context
}

func (c *Config) Prepare(raws ...interface{}) ([]string, error) {
	err := config.Decode(c, &config.DecodeOpts{
		PluginType:         BuilderId,
		Interpolate:        true,
		InterpolateContext: &c.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"boot_command",
				"qemuargs",
			},
		},
	}, raws...)
	if err != nil {
		return nil, err
	}

	// Accumulate any errors and warnings
	var errs *packersdk.MultiError
	warnings := make([]string, 0)

	errs = packersdk.MultiErrorAppend(errs, c.ShutdownConfig.Prepare(&c.ctx)...)

	if c.DiskSize == "" || c.DiskSize == "0" {
		c.DiskSize = "40960M"
	} else {
		// Make sure supplied disk size is valid
		// (digits, plus an optional valid unit character). e.g. 5000, 40G, 1t
		re := regexp.MustCompile(`^[\d]+(b|k|m|g|t){0,1}$`)
		matched := re.MatchString(strings.ToLower(c.DiskSize))
		if !matched {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("Invalid disk size."))
		} else {
			// Okay, it's valid -- if it doesn't alreay have a suffix, then
			// append "M" as the default unit.
			re = regexp.MustCompile(`^[\d]+$`)
			matched = re.MatchString(strings.ToLower(c.DiskSize))
			if matched {
				// Needs M added.
				c.DiskSize = fmt.Sprintf("%sM", c.DiskSize)
			}
		}
	}

	if c.DiskCache == "" {
		c.DiskCache = "writeback"
	}

	if c.DiskDiscard == "" {
		c.DiskDiscard = "ignore"
	}

	if c.DetectZeroes == "" {
		c.DetectZeroes = "off"
	}

	if c.Accelerator == "" {
		if runtime.GOOS == "windows" {
			c.Accelerator = "tcg"
		} else {
			// /dev/kvm is a kernel module that may be loaded if kvm is
			// installed and the host supports VT-x extensions. To make sure
			// this will actually work we need to os.Open() it. If os.Open fails
			// the kernel module was not installed or loaded correctly.
			if fp, err := os.Open("/dev/kvm"); err != nil {
				c.Accelerator = "tcg"
			} else {
				fp.Close()
				c.Accelerator = "kvm"
			}
		}
		log.Printf("use detected accelerator: %s", c.Accelerator)
	} else {
		log.Printf("use specified accelerator: %s", c.Accelerator)
	}

	if c.MachineType == "" {
		c.MachineType = "pc"
	}

	if c.OutputDir == "" {
		c.OutputDir = fmt.Sprintf("output-%s", c.PackerBuildName)
	}

	if c.QemuBinary == "" {
		c.QemuBinary = "qemu-system-x86_64"
	}

	if c.MemorySize < 10 {
		log.Printf("MemorySize %d is too small, using default: 512", c.MemorySize)
		c.MemorySize = 512
	}

	if c.VNCBindAddress == "" {
		c.VNCBindAddress = "127.0.0.1"
	}

	if c.VNCPortMin == 0 {
		c.VNCPortMin = 5900
	}

	if c.VNCPortMax == 0 {
		c.VNCPortMax = 6000
	}

	if c.VMName == "" {
		c.VMName = fmt.Sprintf("packer-%s", c.PackerBuildName)
	}

	if c.Format == "" {
		c.Format = "qcow2"
	}

	if c.TPMType == "" {
		c.TPMType = "tpm-tis"
	}

	c.QemuEFIBootConfig.loadDefaults()

	errs = packersdk.MultiErrorAppend(errs, c.FloppyConfig.Prepare(&c.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, c.CDConfig.Prepare(&c.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, c.VNCConfig.Prepare(&c.ctx)...)

	if c.NetDevice == "" {
		c.NetDevice = "virtio-net"
	}

	if c.DiskInterface == "" {
		c.DiskInterface = "virtio"
	}

	if c.ISOSkipCache {
		c.ISOChecksum = "none"
	}
	isoWarnings, isoErrs := c.ISOConfig.Prepare(&c.ctx)
	warnings = append(warnings, isoWarnings...)
	errs = packersdk.MultiErrorAppend(errs, isoErrs...)

	errs = packersdk.MultiErrorAppend(errs, c.HTTPConfig.Prepare(&c.ctx)...)
	commConfigWarnings, es := c.CommConfig.Prepare(&c.ctx)
	if len(es) > 0 {
		errs = packersdk.MultiErrorAppend(errs, es...)
	}
	warnings = append(warnings, commConfigWarnings...)

	if !(c.Format == "qcow2" || c.Format == "raw") {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("invalid format, only 'qcow2' or 'raw' are allowed"))
	}

	if c.Format != "qcow2" {
		c.SkipCompaction = true
		c.DiskCompression = false
	}

	if c.UseBackingFile {
		c.SkipCompaction = true
		if !(c.DiskImage && c.Format == "qcow2") {
			errs = packersdk.MultiErrorAppend(
				errs, errors.New("use_backing_file can only be enabled for QCOW2 images and when disk_image is true"))
		}
	}

	if c.SkipResizeDisk && !(c.DiskImage) {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("skip_resize_disk can only be used when disk_image is true"))
	}

	if _, ok := accels[c.Accelerator]; !ok {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("invalid accelerator, only 'kvm', 'tcg', 'xen', 'hax', 'hvf', 'whpx', or 'none' are allowed"))
	}

	if _, ok := diskInterface[c.DiskInterface]; !ok {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("unrecognized disk interface type"))
	}

	if _, ok := diskCache[c.DiskCache]; !ok {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("unrecognized disk cache type"))
	}

	if _, ok := diskDiscard[c.DiskDiscard]; !ok {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("unrecognized disk discard type"))
	}

	if _, ok := diskDZeroes[c.DetectZeroes]; !ok {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("unrecognized disk detect zeroes setting"))
	}

	if c.CpuCount < 0 {
		errs = packersdk.MultiErrorAppend(errs, errors.New("cpus must be a positive number"))
	}

	if c.SocketCount < 0 {
		errs = packersdk.MultiErrorAppend(errs, errors.New("sockets must be a positive number"))
	}
	if c.CoreCount < 0 {
		errs = packersdk.MultiErrorAppend(errs, errors.New("cores must be a positive number"))
	}
	if c.ThreadCount < 0 {
		errs = packersdk.MultiErrorAppend(errs, errors.New("threads must be a positive number"))
	}

	if !c.PackerForce {
		if _, err := os.Stat(c.OutputDir); err == nil {
			errs = packersdk.MultiErrorAppend(
				errs,
				fmt.Errorf("Output directory '%s' already exists. It must not exist.", c.OutputDir))
		}
	}

	if c.VNCPortMin < 5900 {
		errs = packersdk.MultiErrorAppend(
			errs, fmt.Errorf("vnc_port_min cannot be below 5900"))
	}

	if c.VNCPortMin > 65535 || c.VNCPortMax > 65535 {
		errs = packersdk.MultiErrorAppend(
			errs, fmt.Errorf("vmc_port_min and vnc_port_max must both be below 65535 to be valid TCP ports"))
	}

	if c.VNCPortMin > c.VNCPortMax {
		errs = packersdk.MultiErrorAppend(
			errs, fmt.Errorf("vnc_port_min must be less than vnc_port_max"))
	}

	if c.NetBridge != "" && runtime.GOOS != "linux" {
		errs = packersdk.MultiErrorAppend(
			errs, fmt.Errorf("net_bridge is only supported in Linux based OSes"))
	}

	if c.NetBridge != "" || c.VNCUsePassword {
		c.QMPEnable = true
	}

	if c.QMPEnable && c.QMPSocketPath == "" {
		socketName := fmt.Sprintf("%s.monitor", c.VMName)
		c.QMPSocketPath = filepath.Join(c.OutputDir, socketName)
	}

	if c.QemuArgs == nil {
		c.QemuArgs = make([][]string, 0)
	}

	if errs != nil && len(errs.Errors) > 0 {
		return warnings, errs
	}

	return warnings, nil

}
