// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package apple

import (
	"context"
	"errors"

	"github.com/siderolabs/talos/pkg/provision"
)

// Create provisions a Talos cluster on Apple's `container` runtime.
//
// TODO(G5): implement, mirroring docker/create.go + node.go. The node-launch step must
// emit the recipe verified by hand in G4 (docs/runbook.md "G4"):
//   - `container run -d --cap-add ALL`                       (G2: machined needs full caps)
//   - --tmpfs /run /tmp /system /system/state /var /etc/cni /etc/kubernetes /usr/libexec/kubernetes
//     but NOT /opt                                           (G4: /opt shadows shipped /opt/cni/bin)
//   - control-plane memory >= ~2GB (-m)                      (G4: 512Mi apiserver OOMs at 1GB)
//   - env PLATFORM=container, TALOSSKU=..., USERDATA=<base64 machine config>
//   - after start: `container inspect` to discover the DHCP-assigned vmnet IP (G3),
//     then populate ClusterInfo.Nodes[].IPs from it (no static --ip available).
func (p *provisioner) Create(ctx context.Context, request provision.ClusterRequest, opts ...provision.Option) (provision.Cluster, error) {
	return nil, errors.New("apple-container provisioner: Create not yet implemented (G5 in progress)")
}
