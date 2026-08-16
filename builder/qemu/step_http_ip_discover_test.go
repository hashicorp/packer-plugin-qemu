// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: MPL-2.0

package qemu

import (
	"bytes"
	"context"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func TestStepHTTPIPDiscover_Run(t *testing.T) {
	testcases := []struct {
		name        string
		httpAddress string
		wantIP      string
	}{
		{
			name:   "user mode NAT default",
			wantIP: "10.0.2.2",
		},
		{
			name:        "specific http_bind_address",
			httpAddress: "1.2.3.4",
			wantIP:      "1.2.3.4",
		},
		{
			name:        "wildcard IPv4 bind still uses NAT address",
			httpAddress: "0.0.0.0",
			wantIP:      "10.0.2.2",
		},
		{
			name:        "wildcard IPv6 bind still uses NAT address",
			httpAddress: "::",
			wantIP:      "10.0.2.2",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			state := new(multistep.BasicStateBag)
			state.Put("ui", &packersdk.BasicUi{
				Reader: new(bytes.Buffer),
				Writer: new(bytes.Buffer),
			})
			config := &Config{}
			config.HTTPAddress = tc.httpAddress
			state.Put("config", config)
			step := new(stepHTTPIPDiscover)

			if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
				t.Fatalf("bad action: %#v", action)
			}
			if _, ok := state.GetOk("error"); ok {
				t.Fatal("should NOT have error")
			}
			httpIp := state.Get("http_ip").(string)
			if httpIp != tc.wantIP {
				t.Fatalf("bad: Http ip is %s but was supposed to be %s", httpIp, tc.wantIP)
			}
		})
	}
}
