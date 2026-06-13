// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package apple

import (
	"context"

	"github.com/siderolabs/talos/pkg/provision"
)

// Reflect reconstructs a Cluster from saved state, for callers (e.g. `cluster destroy`/`show`)
// that did not run Create in this process.
//
// Unlike the docker provider — which rebuilds from the live runtime by listing containers by
// label — apple/container's `container ls` has no label filter, so we read the state.yaml that
// Create persisted (provision.ReadState). *provision.State already satisfies provision.Cluster
// (Provisioner/Info/StatePath), and it carries the node IDs and recorded IPs, which is what
// Destroy needs. Node IDs are stable (we launch with --name), so teardown is reliable even
// though a node's DHCP IP may have changed since Create (the G3 cold-restart gap); a future
// refinement could refresh IPs via `container inspect`.
func (p *provisioner) Reflect(ctx context.Context, clusterName, stateDirectory string) (provision.Cluster, error) {
	state, err := provision.ReadState(ctx, clusterName, stateDirectory)
	if err != nil {
		return nil, err
	}

	return state, nil
}
