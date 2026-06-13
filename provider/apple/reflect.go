// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package apple

import (
	"context"
	"errors"

	"github.com/siderolabs/talos/pkg/provision"
)

// Reflect reconstructs a Cluster from the live runtime + saved state, without a prior
// Create in this process (used by `cluster destroy`/`show`).
//
// TODO(G5): implement, mirroring docker/reflect.go: read state.yaml via provision.ReadState,
// then `container inspect` / `container network inspect` to refresh node IPs and network info.
func (p *provisioner) Reflect(ctx context.Context, clusterName, stateDirectory string) (provision.Cluster, error) {
	return nil, errors.New("apple-container provisioner: Reflect not yet implemented (G5 in progress)")
}
