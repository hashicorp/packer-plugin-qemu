source "qemu" "debian_efi" {
	iso_url          = "https://cdimage.debian.org/debian-cd/current/amd64/iso-cd/debian-11.5.0-amd64-netinst.iso"
	iso_checksum     = "sha256:e307d0e583b4a8f7e5b436f8413d4707dd4242b70aea61eb08591dc0378522f3"
	communicator     = "ssh"
	ssh_username     = "root"
	ssh_password     = "root"
	ssh_timeout      = "30m"
	output_directory = "./out"
	memory           = "1024"
	disk_size        = "6G"
	cpus             = 4
	format           = "qcow2"
	accelerator      = "kvm"
	vm_name          = "debian_efi"
	# headless         = "false" # uncomment to see the boot process in a qemu window
	machine_type     = "q35" # As of now, q35 is required for secure boot to be enabled
	boot_steps     = [
		["<enter>FS0:<enter>EFI\\boot\\bootx64.efi<enter>", "boot from EFI shell"],
		["<wait><down><down><enter>", "manual install"],
		["<wait><down><down><down><down><down><enter>", "automatic install"],
		["<wait30>", "wait 30s for preseed prompt"],
		["http://{{.HTTPIP}}:{{.HTTPPort}}/preseed.cfg<tab><enter>", "select preseed medium"],
		["<wait><enter>", "select English as language/locale"],
		["<wait><enter>", "select English as language"],
		["<wait><enter>", "set English-US as keyboard layout"],
		["<wait><wait><wait>root<enter>", "set root password"],
		["<wait>root<enter>", "confirm root password"],
		["<wait>debian<enter>", "set machine name to debian"],
		["<wait><enter>", "set user to debian"],
		["<wait>debian<enter>", "set password to debian"],
		["<wait>debian<enter>", "confirm password to debian"],
		["<wait180>", "wait 3m for system to install"],
		["root<enter>root<enter>sed -Ei 's/^#.*PermitRootLogin.*$/PermitRootLogin yes/' /etc/ssh/sshd_config<enter>systemctl restart sshd<enter>exit<enter>", "configure sshd to allow root connection"],
	]
	http_directory = "http"
	boot_wait     = "3s"
	qemuargs          = [
		["-cpu", "host"],
		["-vga","virtio"] # if vga is not virtio, output is garbled for some reason
	]
	vtpm              = true
	efi_firmware_code = "./efi_data/OVMF_CODE_4M.ms.fd"
	efi_firmware_vars = "./efi_data/OVMF_VARS_4M.ms.fd"
}

build {
	sources = ["source.qemu.debian_efi"]

	provisioner "shell" {
		inline = [ "dmesg | grep -qi 'Secure boot enabled' && echo \"Secure Boot is on!\"" ]
	}
}
