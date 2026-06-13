// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package makers

import (
	"slices"

	"github.com/siderolabs/talos/cmd/talosctl/cmd/mgmt/cluster/create/clusterops"
)

var _ ConfigMaker = &(Apple{})

// Apple is the maker for the apple-container provider.
type Apple struct {
	*Maker[clusterops.Apple]
}

// NewApple returns a new initialized Apple maker.
func NewApple(ops MakerOptions[clusterops.Apple]) (Apple, error) {
	maker, err := New(ops)
	if err != nil {
		return Apple{}, err
	}

	m := Apple{Maker: &maker}

	m.SetExtraOptionsProvider(&m)

	if err := m.Init(); err != nil {
		return Apple{}, err
	}

	return m, nil
}

// InitExtra implements ExtraOptionsProvider.
func (m *Apple) InitExtra() error { return nil }

// AddExtraConfigBundleOpts implements ExtraOptionsProvider.
func (m *Apple) AddExtraConfigBundleOpts() error { return nil }

// AddExtraGenOps adds the control-plane endpoints to the apiserver cert SANs (as Docker does).
func (m *Apple) AddExtraGenOps() error {
	m.GenOps = slices.Concat(m.GenOps, getWithAdditionalSubjectAltNamesGenOps(m.Endpoints))

	return nil
}

// AddExtraProvisionOpts implements ExtraOptionsProvider.
func (m *Apple) AddExtraProvisionOpts() error { return nil }

// ModifyClusterRequest sets the Talos image. The apple provider discovers node IPs at runtime
// (vmnet DHCP), so unlike Docker there is no static-IP or IPv6 toggle to apply here.
func (m *Apple) ModifyClusterRequest() error {
	m.ClusterRequest.Image = m.EOps.TalosImage

	return nil
}

// ModifyNodes implements ExtraOptionsProvider.
func (m *Apple) ModifyNodes() error { return nil }
