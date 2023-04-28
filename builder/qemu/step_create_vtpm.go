// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qemu

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepCreatevTPM struct {
	enableVTPM bool
	vtpmType   string
	isTPM1     bool
}

const (
	qemuVTPM        string = "qemu_vtpm"
	swtpmProcess    string = "qemu_swtpm_process"
	swtpmTmpDir     string = "qemu_swtpm_dir"
	swtpmSocketPath string = "qemu_swtpm_socket_path"
)

func (s *stepCreatevTPM) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if !s.enableVTPM {
		return multistep.ActionContinue
	}

	ui := state.Get("ui").(packersdk.Ui)

	if runtime.GOOS == "windows" {
		ui.Error("vTPM is only supported on UNIX OSes for now")
		return multistep.ActionHalt
	}

	swtpmPath, err := exec.LookPath("swtpm")
	if err != nil {
		ui.Error(fmt.Sprintf(
			"failed to locate swtpm (%s), this is required for vTPM support",
			err))
		return multistep.ActionHalt
	}

	vtpmDeviceDir, err := os.MkdirTemp("", "")
	if err != nil {
		ui.Error(fmt.Sprintf("failed to create vtpm state directory: %s", err))
		return multistep.ActionHalt
	}

	state.Put(swtpmTmpDir, vtpmDeviceDir)

	sockPath := fmt.Sprintf("%s/vtpm.sock", vtpmDeviceDir)

	state.Put(swtpmSocketPath, sockPath)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"failed to create swtpm communication: %s",
			err))
		return multistep.ActionHalt
	}

	args := []string{
		"socket",
		"--tpmstate", fmt.Sprintf("dir=%s", vtpmDeviceDir),
		"--ctrl", fmt.Sprintf("type=unixio,path=%s", sockPath),
	}

	if !s.isTPM1 {
		args = append(args, "--tpm2")
	}

	swtpm := exec.Command(swtpmPath, args...)
	swtpm.Stdout = os.Stdout
	swtpm.Stderr = os.Stderr

	state.Put(qemuVTPM, true)

	log.Printf("Executing swtpm: %+v", args)
	err = swtpm.Start()
	if err != nil {
		ui.Error(fmt.Sprintf(
			"failed to start swtpm: %s", err))
		return multistep.ActionHalt
	}

	state.Put(swtpmProcess, swtpm.Process)

	return multistep.ActionContinue
}

func (s *stepCreatevTPM) Cleanup(state multistep.StateBag) {
	process, ok := state.GetOk(swtpmProcess)
	if !ok {
		return
	}

	log.Printf("killing swtpm with PID %d", process.(*os.Process).Pid)
	err := process.(*os.Process).Kill()
	if err != nil {
		log.Printf("failed to kill swtpm: %s", err)
	}

	tmpDir := state.Get(swtpmTmpDir).(string)
	os.RemoveAll(tmpDir)
}
