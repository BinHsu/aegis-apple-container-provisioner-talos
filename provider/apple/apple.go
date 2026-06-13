// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package apple implements the Talos provision.Provisioner interface for Apple's
// `container` runtime (Virtualization.framework micro-VMs, one micro-VM per node).
//
// The layout mirrors the in-tree providers (pkg/provision/providers/docker and
// .../qemu) so an eventual upstream move is a directory copy. Structurally this is
// the docker case (Talos runs as a container on a foreign, Kata-derived kernel with
// Apple's vminitd as PID 1), and it follows the qemu pattern of exec-ing a CLI
// (`container`) rather than calling a daemon API.
//
// The verified node-launch recipe (caps, tmpfs set, memory) lives in node.go; it was
// established by hand in the G1-G4 spike (see docs/runbook.md, docs/VERIFICATION.md).
package apple

import (
	"context"

	"github.com/siderolabs/talos/pkg/machinery/config"
	"github.com/siderolabs/talos/pkg/machinery/config/bundle"
	"github.com/siderolabs/talos/pkg/machinery/config/generate"
	"github.com/siderolabs/talos/pkg/machinery/config/types/v1alpha1"
	"github.com/siderolabs/talos/pkg/machinery/nethelpers"
	"github.com/siderolabs/talos/pkg/provision"
)

// ProviderName is the name this provisioner registers under (mirrors docker/qemu).
const ProviderName = "apple-container"

// firstInterface is the in-VM NIC name Talos sees, same as the docker provider.
const firstInterface = "eth0"

// Compile-time proof we satisfy the real upstream interface. This assertion failing to
// build is exactly the signal that our provider has drifted from pkg/provision — the
// "directory move, not rewrite" contract this repo exists to keep.
var _ provision.Provisioner = (*provisioner)(nil)

// provisioner drives Apple's `container` CLI to provision Talos nodes.
//
// Unlike the docker provisioner it holds no mapped host ports: apple/container nodes get
// vmnet IPs that are directly reachable from the host (verified G4: host -> 192.168.64.x:6443),
// so there is no localhost port-forwarding to track.
type provisioner struct {
	// containerCLI is the `container` binary we exec (the qemu-provider exec pattern).
	containerCLI string
}

// NewProvisioner initializes the apple/container provisioner.
//
//nolint:revive // ctx kept for signature parity with docker.NewProvisioner / future use.
func NewProvisioner(ctx context.Context) (provision.Provisioner, error) {
	return &provisioner{
		containerCLI: "container",
	}, nil
}

// Close releases resources. The `container` daemon is long-lived and host-managed
// (launchd), and we exec a CLI rather than holding a client handle, so there is
// nothing to close.
func (p *provisioner) Close() error {
	return nil
}

// GenOptions provides additional config-generate options. Mirrors the docker provider:
// Talos runs in container mode here too, so the same host-DNS and single-doc-network
// adjustments apply.
func (p *provisioner) GenOptions(_ provision.NetworkRequest, contract *config.VersionContract) ([]generate.Option, []bundle.Option) {
	var genOptions []generate.Option

	if contract.HostDNSEnabled() {
		genOptions = append(genOptions,
			generate.WithHostDNSForwardKubeDNSToHost(true),
		)
	}

	if !contract.MultidocNetworkConfigSupported() {
		genOptions = append(genOptions,
			generate.WithNetworkOptions(
				v1alpha1.WithNetworkInterfaceIgnore(v1alpha1.IfaceByName(firstInterface)),
			),
		)
	}

	return genOptions, nil
}

// GetInClusterKubernetesControlPlaneEndpoint returns the in-cluster Kubernetes API endpoint.
//
// TODO(G5): apple/container assigns node IPs via vmnet DHCP (no static --ip; verified G3),
// so the control-plane IP is NOT predictable from the CIDR the way docker's .Next().Next()
// assumes. Create() must discover the control-plane node IP after launch and this endpoint
// must be derived from it. The CIDR-based value below is a placeholder to keep the build green.
func (p *provisioner) GetInClusterKubernetesControlPlaneEndpoint(networkReq provision.NetworkRequest, controlPlanePort int) string {
	return "https://" + nethelpers.JoinHostPort(networkReq.CIDRs[0].Addr().Next().Next().String(), controlPlanePort)
}

// GetExternalKubernetesControlPlaneEndpoint returns the externally reachable Kubernetes API
// endpoint. Unlike docker, apple/container needs no localhost port mapping — the vmnet node IP
// is reachable from the host directly (verified G4).
//
// TODO(G5): same DHCP caveat as the in-cluster endpoint — derive from the discovered CP IP.
func (p *provisioner) GetExternalKubernetesControlPlaneEndpoint(networkReq provision.NetworkRequest, controlPlanePort int) string {
	return p.GetInClusterKubernetesControlPlaneEndpoint(networkReq, controlPlanePort)
}

// GetTalosAPIEndpoints returns the Talos API (apid) endpoints, reachable on the vmnet node IPs.
//
// TODO(G5): derive from discovered node IPs (DHCP). Placeholder uses the first CIDR address.
func (p *provisioner) GetTalosAPIEndpoints(networkReq provision.NetworkRequest) []string {
	return []string{networkReq.CIDRs[0].Addr().Next().Next().String()}
}

// GetFirstInterface returns the first network interface selector.
func (p *provisioner) GetFirstInterface() v1alpha1.IfaceSelector {
	return v1alpha1.IfaceByName(firstInterface)
}

// GetFirstInterfaceName returns the first network interface name.
func (p *provisioner) GetFirstInterfaceName() string {
	return firstInterface
}

// UserDiskName is not applicable to the container model (no user disks); mirrors docker.
func (p *provisioner) UserDiskName(_ int) string {
	return ""
}
