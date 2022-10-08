package qemu

import (
	"context"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

const KeyLeftShift uint32 = 0xFFE1

type bootCommandTemplateData struct {
	HTTPIP   string
	HTTPPort int
	Name     string
}

// This step "types" the boot command into the VM over VNC.
//
// Uses:
//   config *config
//   http_port int
//   ui     packersdk.Ui
//   vnc_port int
//
// Produces:
//   <nothing>
type stepTypeBootCommand struct{}

func (s *stepTypeBootCommand) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	command := config.VNCConfig.FlatBootCommand()
	bootSteps := [][]string{{command}}

	return typeBootCommands(ctx, state, bootSteps)
}

func (*stepTypeBootCommand) Cleanup(multistep.StateBag) {}
