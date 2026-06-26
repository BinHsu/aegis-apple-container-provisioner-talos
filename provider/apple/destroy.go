// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package apple

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/siderolabs/talos/pkg/provision"
)

// Destroy tears down a provisioned cluster. It is idempotent (stop/remove, volume deletes, and the
// state-dir removal all ignore "not found"/not-exist) and, per the G4 acceptance criterion, leaves
// `container ls -a` clean.
//
// Node IDs and names come from the cluster's recorded state (cluster.StatePath() + Info().Nodes),
// so teardown does not depend on `container ls` label filtering (which the CLI does not support),
// and it works in a fresh process via Reflect/ReadState.
//
// Critical side effect of persistent state volumes: deleting each node's /var and /system/state
// NAMED VOLUMES is mandatory. Skipping it would leave old etcd data + machine config + PKI behind,
// so recreating a same-named cluster would hit the create-time stale-state guard (or boot onto stale
// state). The volumes are container-managed (not under statePath), so they are deleted explicitly per
// node; the separate RemoveAll(statePath) only sweeps the provisioner's own state dir (state.yaml).
//
// Known gap: this reflects recorded state, so it can only clean a cluster whose Create reached
// saveState. A Create that FAILED earlier (e.g. a stuck container before state.yaml was written)
// leaves no record for Destroy to act on. A label-based sweep (talos.owned=true) would let Destroy
// clean half-created clusters — tracked in docs/VERIFICATION.md G5.
func (p *provisioner) Destroy(ctx context.Context, cluster provision.Cluster, opts ...provision.Option) error {
	options := provision.DefaultOptions()

	for _, opt := range opts {
		if err := opt(&options); err != nil {
			return err
		}
	}

	info := cluster.Info()

	// statePath is the provisioner's own state dir (state.yaml). It may be unset on a hand-built
	// Cluster; if so we skip the state-dir sweep but still stop/remove containers and delete volumes.
	statePath, statePathErr := cluster.StatePath()

	var errs []error

	for _, node := range info.Nodes {
		fmt.Fprintln(options.LogWriter, "destroying node", node.Name)

		if err := p.stop(ctx, node.ID); err != nil {
			errs = append(errs, err)
		}

		if err := p.remove(ctx, node.ID); err != nil {
			errs = append(errs, err)
		}

		// Delete the node's persistent /var + /system/state named volumes. volumeDelete ignores
		// "not found", so this is idempotent. Names derive from the same nodeVolumeNames create used.
		varVol, systemStateVol := nodeVolumeNames(info.ClusterName, node.Name)

		for _, vol := range []string{varVol, systemStateVol} {
			if err := p.volumeDelete(ctx, vol); err != nil {
				errs = append(errs, err)
			}
		}
	}

	if err := p.destroyNetwork(ctx, info.Network.Name); err != nil {
		errs = append(errs, err)
	}

	// Per-cluster sweep: removes the provisioner state dir (state.yaml). Volumes are container-managed,
	// not under statePath, so they are deleted per node above — not by this RemoveAll.
	if statePathErr == nil {
		if err := os.RemoveAll(statePath); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
