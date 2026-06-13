// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package configmaker

import (
	"github.com/siderolabs/talos/cmd/talosctl/cmd/mgmt/cluster/create/clusterops"
	"github.com/siderolabs/talos/cmd/talosctl/cmd/mgmt/cluster/create/clusterops/configmaker/internal/makers"
)

// AppleOptions are the options for provisioning an apple-container based Talos cluster.
type AppleOptions makers.MakerOptions[clusterops.Apple]

// GetAppleConfigs returns the cluster configs for apple-container.
func GetAppleConfigs(options AppleOptions) (clusterops.ClusterConfigs, error) {
	maker, err := makers.NewApple(options)
	if err != nil {
		return clusterops.ClusterConfigs{}, err
	}

	return maker.GetClusterConfigs()
}
