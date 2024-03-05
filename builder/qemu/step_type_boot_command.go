// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qemu

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/bootcommand"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/mitchellh/go-vnc"
)

const KeyLeftShift uint32 = 0xFFE1

type bootCommandTemplateData struct {
	HTTPIP       string
	HTTPPort     int
	Name         string
	SSHPublicKey string
}

// This step "types" the boot command into the VM over VNC.
//
// Uses:
//
//	config *config
//	http_port int
//	ui     packersdk.Ui
//	vnc_port int
//
// Produces:
//
//	<nothing>
type stepTypeBootCommand struct{}

func (s *stepTypeBootCommand) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	command := config.VNCConfig.FlatBootCommand()
	bootSteps := config.BootSteps

	if len(command) > 0 {
		bootSteps = [][]string{{command}}
	}

	return typeBootCommands(ctx, state, bootSteps)
}

func (*stepTypeBootCommand) Cleanup(multistep.StateBag) {}

func typeBootCommands(ctx context.Context, state multistep.StateBag, bootSteps [][]string) multistep.StepAction {
	config := state.Get("config").(*Config)
	debug := state.Get("debug").(bool)
	httpPort := state.Get("http_port").(int)
	ui := state.Get("ui").(packersdk.Ui)
	vncPort := state.Get("vnc_port").(int)
	vncIP := config.VNCBindAddress
	vncPassword := state.Get("vnc_password")

	if config.VNCConfig.DisableVNC {
		log.Println("Skipping boot command step...")
		return multistep.ActionContinue
	}

	// Wait the for the vm to boot.
	if int64(config.BootWait) > 0 {
		ui.Say(fmt.Sprintf("Waiting %s for boot...", config.BootWait))
		select {
		case <-time.After(config.BootWait):
			break
		case <-ctx.Done():
			return multistep.ActionHalt
		}
	}

	var pauseFn multistep.DebugPauseFn
	if debug {
		pauseFn = state.Get("pauseFn").(multistep.DebugPauseFn)
	}

	// Connect to VNC
	ui.Say(fmt.Sprintf("Connecting to VM via VNC (%s:%d)", vncIP, vncPort))

	nc, err := net.Dial("tcp", fmt.Sprintf("%s:%d", vncIP, vncPort))
	if err != nil {
		err := fmt.Errorf("Error connecting to VNC: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	defer nc.Close()

	var auth []vnc.ClientAuth

	if vncPassword != nil && len(vncPassword.(string)) > 0 {
		auth = []vnc.ClientAuth{&vnc.PasswordAuth{Password: vncPassword.(string)}}
	} else {
		auth = []vnc.ClientAuth{new(vnc.ClientAuthNone)}
	}

	c, err := vnc.Client(nc, &vnc.ClientConfig{Auth: auth, Exclusive: false})
	if err != nil {
		err := fmt.Errorf("Error handshaking with VNC: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	defer c.Close()

	log.Printf("Connected to VNC desktop: %s", c.DesktopName)

	hostIP := state.Get("http_ip").(string)
	SSHPublicKey := string(config.CommConfig.Comm.SSHPublicKey)
	configCtx := config.ctx
	configCtx.Data = &bootCommandTemplateData{
		hostIP,
		httpPort,
		config.VMName,
		SSHPublicKey,
	}

	d := bootcommand.NewVNCDriver(c, config.VNCConfig.BootKeyInterval)

	ui.Say("Typing the boot commands over VNC...")

	for _, step := range bootSteps {
		if len(step) == 0 {
			continue
		}

		var description string

		if len(step) >= 2 {
			description = step[1]
		} else {
			description = ""
		}

		if len(description) > 0 {
			ui.Say(fmt.Sprintf("Typing boot command for: %s", description))
		}

		command, err := interpolate.Render(step[0], &configCtx)

		if err != nil {
			err := fmt.Errorf("Error preparing boot command: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		seq, err := bootcommand.GenerateExpressionSequence(command)
		if err != nil {
			err := fmt.Errorf("Error generating boot command: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		if err := seq.Do(ctx, d); err != nil {
			err := fmt.Errorf("Error running boot command: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		if pauseFn != nil {
			var message string

			if len(description) > 0 {
				message = fmt.Sprintf("boot description: \"%s\", command: %s", description, command)
			} else {
				message = fmt.Sprintf("boot_command: %s", command)
			}

			pauseFn(multistep.DebugLocationAfterRun, message, state)
		}
	}

	return multistep.ActionContinue
}
