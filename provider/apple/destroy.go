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

// Destroy tears down a provisioned cluster. It is idempotent (stop/remove ignore
// "not found") and, per the G4 acceptance criterion, leaves `container ls -a` clean.
//
// Node IDs come from the cluster's recorded state, so teardown does not depend on
// `container ls` label filtering (which the CLI does not support).
func (p *provisioner) Destroy(ctx context.Context, cluster provision.Cluster, opts ...provision.Option) error {
	options := provision.DefaultOptions()

	for _, opt := range opts {
		if err := opt(&options); err != nil {
			return err
		}
	}

	info := cluster.Info()

	var errs []error

	for _, node := range info.Nodes {
		fmt.Fprintln(options.LogWriter, "destroying node", node.Name)

		if err := p.stop(ctx, node.ID); err != nil {
			errs = append(errs, err)
		}

		if err := p.remove(ctx, node.ID); err != nil {
			errs = append(errs, err)
		}
	}

	if err := p.destroyNetwork(ctx, info.Network.Name); err != nil {
		errs = append(errs, err)
	}

	if statePath, err := cluster.StatePath(); err == nil {
		if err := os.RemoveAll(statePath); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
