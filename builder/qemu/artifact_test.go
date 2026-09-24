// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: MPL-2.0

package qemu

import (
	"testing"

	registryimage "github.com/hashicorp/packer-plugin-sdk/packer/registry/image"
)

func TestArtifactState_RegistryImage(t *testing.T) {
	a := &Artifact{
		dir:   "output-rhel",
		state: map[string]interface{}{"diskName": "rhel.qcow2"},
	}

	img, ok := a.State(registryimage.ArtifactStateURI).(*registryimage.Image)
	if !ok {
		t.Fatalf("State(%q) did not return a *registryimage.Image", registryimage.ArtifactStateURI)
	}
	if img.ImageID != "rhel.qcow2" || img.ProviderRegion != "output-rhel" || img.ProviderName != BuilderId {
		t.Errorf("unexpected image: %+v", img)
	}
	if err := img.Validate(); err != nil {
		t.Errorf("image failed validation: %v", err)
	}
}
