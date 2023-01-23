package qemu

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// stepPrepareEfivars copies the EFIVars file to the output, so we can boot
// and use it as a RW flash drive
type stepPrepareEfivars struct {
	EFIEnabled bool
	OutputDir  string
	SourcePath string
}

const efivarStateKey string = "EFI_VARS_FILE_PATH"

func (s *stepPrepareEfivars) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)

	if !s.EFIEnabled {
		return multistep.ActionContinue
	}

	dstPath := filepath.Join(s.OutputDir, "efivars.fd")
	outFile, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY, 0660)
	if err != nil {
		errMsg := fmt.Sprintf("failed to create local efivars file at %s: %s", dstPath, err)
		ui.Error(errMsg)
		return multistep.ActionHalt
	}
	defer outFile.Close()

	state.Put(efivarStateKey, dstPath)

	inFile, err := os.Open(s.SourcePath)
	if err != nil {
		errMsg := fmt.Sprintf("failed to read from efivars file at %s: %s", s.SourcePath, err)
		ui.Error(errMsg)
		return multistep.ActionHalt
	}

	_, err = io.Copy(outFile, inFile)
	if err != nil {
		errMsg := fmt.Sprintf("failed to copy efivars data: %s", err)
		ui.Error(errMsg)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepPrepareEfivars) Cleanup(state multistep.StateBag) {
	if !s.EFIEnabled {
		return
	}

	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if cancelled || halted {
		efiVarFile, ok := state.GetOk(efivarStateKey)
		// If the path isn't in state, we can assume it's not been created and
		// therefore we have nothing to cleanup
		if !ok {
			return
		}

		os.Remove(efiVarFile.(string))
	}
}
