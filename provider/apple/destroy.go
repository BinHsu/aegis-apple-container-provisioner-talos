// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package apple

import (
	"context"
	"errors"

	"github.com/siderolabs/talos/pkg/provision"
)

// Destroy tears down a provisioned cluster. Must be idempotent and leave
// `container ls -a` clean (G4 acceptance criterion).
//
// TODO(G5): implement, mirroring docker/destroy.go: `container stop`/`container rm`
// each node (no-op if already gone), remove the network, then os.RemoveAll(stateDir).
func (p *provisioner) Destroy(ctx context.Context, cluster provision.Cluster, opts ...provision.Option) error {
	return errors.New("apple-container provisioner: Destroy not yet implemented (G5 in progress)")
}
