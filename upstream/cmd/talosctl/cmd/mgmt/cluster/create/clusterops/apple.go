// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package clusterops

import (
	"github.com/siderolabs/talos/cmd/talosctl/pkg/mgmt/helpers"
	"github.com/siderolabs/talos/pkg/images"
)

// Apple are the apple-container-specific cluster options.
//
// Minimal by design: the apple provider assigns no static IPs and exposes no host ports
// (vmnet node IPs are reachable from the host directly), so unlike Docker there are no
// port/host-IP/IPv6/mount knobs here.
type Apple struct {
	TalosImage string
}

// GetApple returns the default apple-container options.
func GetApple() Apple {
	return Apple{
		TalosImage: helpers.DefaultImage(images.DefaultTalosImageRepository),
	}
}
