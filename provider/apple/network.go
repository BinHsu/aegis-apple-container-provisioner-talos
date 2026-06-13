// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package apple

import (
	"context"
	"strings"

	"github.com/siderolabs/talos/pkg/provision"
)

// defaultNetwork is apple/container's built-in network; it always exists and needs no creation.
const defaultNetwork = "default"

// ensureNetwork creates the cluster network if it does not already exist. The built-in
// "default" network is used as-is (verified working for cluster bring-up in G3/G4).
func (p *provisioner) ensureNetwork(ctx context.Context, req provision.NetworkRequest) error {
	if req.Name == "" || req.Name == defaultNetwork {
		return nil
	}

	args := []string{"network", "create"}
	if len(req.CIDRs) > 0 {
		args = append(args, "--subnet", req.CIDRs[0].String())
	}

	args = append(args, req.Name)

	_, err := p.run(ctx, args...)
	if err != nil && strings.Contains(err.Error(), "already") {
		return nil // idempotent: re-use an existing network
	}

	return err
}

// destroyNetwork removes the cluster network, ignoring the built-in default and "not found".
func (p *provisioner) destroyNetwork(ctx context.Context, name string) error {
	if name == "" || name == defaultNetwork {
		return nil
	}

	_, err := p.run(ctx, "network", "delete", name)
	if err != nil && strings.Contains(err.Error(), "not found") {
		return nil
	}

	return err
}
