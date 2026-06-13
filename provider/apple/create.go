// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package apple

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"strconv"

	"github.com/siderolabs/talos/pkg/machinery/constants"
	"github.com/siderolabs/talos/pkg/provision"
)

// Create provisions a Talos cluster on Apple's `container` runtime.
//
// Flow mirrors the in-tree docker provider: init state -> ensure network -> launch
// control-plane then worker nodes -> record ClusterInfo -> save state. The node-launch
// recipe (caps, tmpfs set sans /opt, memory, USERDATA env) lives in node.go and is the
// contract verified by hand in G1-G4.
//
// `container run` pulls the image on demand, so there is no explicit image-pull step.
func (p *provisioner) Create(ctx context.Context, request provision.ClusterRequest, opts ...provision.Option) (provision.Cluster, error) {
	options := provision.DefaultOptions()

	for _, opt := range opts {
		if err := opt(&options); err != nil {
			return nil, err
		}
	}

	statePath := filepath.Join(request.StateDirectory, request.Name)

	fmt.Fprintf(options.LogWriter, "creating state directory in %q\n", statePath)

	state, err := provision.NewState(statePath, ProviderName, request.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize provisioner state: %w", err)
	}

	fmt.Fprintln(options.LogWriter, "ensuring network", request.Network.Name)

	if err = p.ensureNetwork(ctx, request.Network); err != nil {
		return nil, fmt.Errorf("failed to ensure network: %w", err)
	}

	fmt.Fprintln(options.LogWriter, "creating controlplane nodes")

	nodes, err := p.createNodes(ctx, request, request.Nodes.ControlPlaneNodes())
	if err != nil {
		return nil, err
	}

	fmt.Fprintln(options.LogWriter, "creating worker nodes")

	workers, err := p.createNodes(ctx, request, request.Nodes.WorkerNodes())
	if err != nil {
		return nil, err
	}

	nodes = append(nodes, workers...)

	// Kubernetes endpoint uses the discovered control-plane IP. apple/container assigns IPs via
	// DHCP (no static --ip; G3), so unlike docker this cannot be computed from the CIDR upfront.
	var kubernetesEndpoint string

	if len(nodes) > 0 && len(nodes[0].IPs) > 0 {
		kubernetesEndpoint = "https://" + net.JoinHostPort(nodes[0].IPs[0].String(), strconv.Itoa(constants.DefaultControlPlanePort))
	}

	state.ClusterInfo = provision.ClusterInfo{
		ClusterName: request.Name,
		Network: provision.NetworkInfo{
			Name:         request.Network.Name,
			CIDRs:        request.Network.CIDRs,
			GatewayAddrs: request.Network.GatewayAddrs,
			MTU:          request.Network.MTU,
		},
		Nodes:              nodes,
		KubernetesEndpoint: kubernetesEndpoint,
	}

	if err = state.Save(); err != nil {
		return nil, err
	}

	return &result{
		clusterInfo: state.ClusterInfo,
		statePath:   statePath,
	}, nil
}
