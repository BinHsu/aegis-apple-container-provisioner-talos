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

// Config holds optional provisioner parameters for the apple/container provider.
type Config struct {
	// DNSDomain, when set, names every container as <node>.<domain> so Apple's container
	// DNS forwarding resolves it from the host as a stable FQDN. The cluster control-plane
	// endpoint and certificate SANs are set to this FQDN rather than the volatile DHCP IP,
	// so kubeconfig and talosctl survive cold restarts. Default: empty (IP-based, v0.1.x
	// behaviour). Prerequisite: run `sudo container system dns create <domain>` once; the
	// entry does not survive a macOS reboot and must be re-created after one.
	DNSDomain string
}

// provisioner drives Apple's `container` CLI to provision Talos nodes.
//
// Unlike the docker provisioner it holds no mapped host ports: apple/container nodes get
// vmnet IPs that are directly reachable from the host (verified G4: host -> 192.168.64.x:6443),
// so there is no localhost port-forwarding to track.
type provisioner struct {
	// containerCLI is the `container` binary we exec (the qemu-provider exec pattern).
	containerCLI string
	// dnsDomain is the Apple container DNS domain for stable FQDN container naming (see Config).
	dnsDomain string
}

// NewProvisioner initializes the apple/container provisioner.
//
//nolint:revive // ctx kept for signature parity with docker.NewProvisioner / future use.
func NewProvisioner(ctx context.Context, cfg Config) (provision.Provisioner, error) {
	return &provisioner{
		containerCLI: "container",
		dnsDomain:    cfg.DNSDomain,
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

// GetInClusterKubernetesControlPlaneEndpoint returns the CIDR-derived endpoint used as a
// generation-time seed for the machine-config bundle. vmnet DHCP assigns the real IP after
// node launch (no static-IP option; verified G3): Create() reads it from container inspect,
// then overwrites this endpoint in every node's config via reconcileConfigs / patchConfig.
func (p *provisioner) GetInClusterKubernetesControlPlaneEndpoint(networkReq provision.NetworkRequest, controlPlanePort int) string {
	return "https://" + nethelpers.JoinHostPort(networkReq.CIDRs[0].Addr().Next().Next().String(), controlPlanePort)
}

// GetExternalKubernetesControlPlaneEndpoint returns the externally reachable Kubernetes API
// endpoint. Unlike the docker provider, apple/container needs no host port-mapping: vmnet
// node IPs are directly reachable from the host (verified G4). This delegates to the
// in-cluster endpoint; Create() overwrites the real value in every node's config via reconcileConfigs.
func (p *provisioner) GetExternalKubernetesControlPlaneEndpoint(networkReq provision.NetworkRequest, controlPlanePort int) string {
	return p.GetInClusterKubernetesControlPlaneEndpoint(networkReq, controlPlanePort)
}

// GetTalosAPIEndpoints returns the Talos API (apid) endpoints seeded into the talosconfig
// bundle at generation time. The CIDR-derived value is provisional: after Create() runs,
// reconcileConfigs applies the real machine configs and the driver prints the actual
// DHCP-assigned node IPs for the operator to supply to talosctl config endpoint / node.
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
