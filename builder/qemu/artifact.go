// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: MPL-2.0

package qemu

import (
	"fmt"
	"log"
	"os"

	registryimage "github.com/hashicorp/packer-plugin-sdk/packer/registry/image"
)

// Artifact is the result of running the Qemu builder, namely a set
// of files associated with the resulting machine.
type Artifact struct {
	dir   string
	f     []string
	state map[string]interface{}
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Files() []string {
	return a.f
}

func (*Artifact) Id() string {
	return "VM"
}

func (a *Artifact) String() string {
	return fmt.Sprintf("VM files in directory: %s", a.dir)
}

func (a *Artifact) State(name string) interface{} {
	if name == registryimage.ArtifactStateURI {
		diskName, _ := a.state["diskName"].(string)
		opts := []registryimage.ArtifactOverrideFunc{
			registryimage.WithProvider("qemu"),
			registryimage.WithID(diskName),
			registryimage.WithRegion(a.dir),
		}
		if sourceImage, ok := a.state["sourceImage"].(string); ok {
			opts = append(opts, registryimage.WithSourceID(sourceImage))
		}
		img, err := registryimage.FromArtifact(a, opts...)
		if err != nil {
			log.Printf("[DEBUG] error encountered when creating a registry image %v", err)
			return nil
		}

		return img
	}

	return a.state[name]
}

func (a *Artifact) Destroy() error {
	return os.RemoveAll(a.dir)
}
