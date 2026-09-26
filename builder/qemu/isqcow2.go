// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qemu

import (
	"os"
)

// Check for the magic bytes at the start of the file
// which would indicate this is a qcow2 disk image
func isQCOW2(path string) (bool, error) {
	const qcow2Magic = "QFI\xfb"
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	magic := make([]byte, 4)
	_, err = file.Read(magic)
	if err != nil {
		return false, err
	}
	return string(magic) == qcow2Magic, nil
}
