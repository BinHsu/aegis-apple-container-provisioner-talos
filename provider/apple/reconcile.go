// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package apple

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/siderolabs/talos/pkg/machinery/config"
	"github.com/siderolabs/talos/pkg/machinery/config/configpatcher"
	"github.com/siderolabs/talos/pkg/machinery/constants"
	"github.com/siderolabs/talos/pkg/provision"
)

// reconcileConfigs is the apple/container-specific divergence from the docker provider.
//
// Why it exists: the docker provider bakes the machine config in at launch (USERDATA env) because
// it assigns each node a static IP, so cluster.controlPlane.endpoint is known up front.
// apple/container assigns IPs via vmnet DHCP with no static-IP option (verified G3), so the
// control-plane IP is only known after launch. Nodes therefore boot bare into maintenance mode;
// here we patch each config's endpoint with the discovered control-plane IP and apply it over the
// insecure maintenance API. This keeps the entire DHCP workaround inside the provider — the
// upstream pkg/provision framework is unchanged. (The flow mirrors the G4 manual procedure that
// proved a cluster comes up this way.)
//
// When the provisioner has a DNS domain configured, the persisted endpoint is the FQDN
// (nodes[0].ID, set by createNode) rather than the ephemeral DHCP IP. The FQDN is also added
// to cluster.apiServer.certSANs and machine.certSANs so that kubeconfig and talosctl remain
// valid after a cold restart that moves the DHCP IP. apply-config itself is still sent to the
// node's current IP over the maintenance API (maintenance mode does not know the FQDN yet).
//
// orderedReqs and nodes are index-aligned (control-plane first, then workers).
func (p *provisioner) reconcileConfigs(
	ctx context.Context,
	request provision.ClusterRequest,
	orderedReqs []provision.NodeRequest,
	nodes []provision.NodeInfo,
	options *provision.Options,
) error {
	controlPlaneIP := nodes[0].IPs[0]

	// Determine the persisted endpoint. When a DNS domain is set, use the FQDN
	// (nodes[0].ID == "<cluster>-controlplane-1.<domain>") for stable cross-restart
	// connectivity. Otherwise fall back to the discovered IP (v0.1.x behaviour).
	var endpoint, cpFQDN string
	if p.dnsDomain != "" {
		cpFQDN = nodes[0].ID // e.g. "aegis-controlplane-1.aegis"
		endpoint = "https://" + net.JoinHostPort(cpFQDN, strconv.Itoa(constants.DefaultControlPlanePort))
	} else {
		endpoint = "https://" + net.JoinHostPort(controlPlaneIP.String(), strconv.Itoa(constants.DefaultControlPlanePort))
	}

	for i := range orderedReqs {
		req := orderedReqs[i]

		if req.SkipInjectingConfig || req.Config == nil {
			continue
		}

		patched, err := patchConfig(req.Config, endpoint, cpFQDN)
		if err != nil {
			return fmt.Errorf("patching config for %q: %w", req.Name, err)
		}

		fmt.Fprintln(options.LogWriter, "applying config to", req.Name, "at", nodes[i].IPs[0])

		// apply-config runs against the node's current IP over the maintenance API. The
		// FQDN is not resolvable at this point (the node is in maintenance mode and has not
		// yet registered with container DNS). Only the PERSISTED endpoint + certSANs in the
		// config use the FQDN.
		if err = p.applyConfig(ctx, request.SelfExecutable, nodes[i].IPs[0], patched); err != nil {
			return fmt.Errorf("applying config to %q: %w", req.Name, err)
		}
	}

	return nil
}

// patchConfigYAML generates the strategic-merge YAML for patchConfig.
// Extracted as a pure function so the YAML shape is unit-testable without a real config.Provider.
//
// When cpFQDN is non-empty, the patch also adds it to cluster.apiServer.certSANs (kube-apiserver
// certificate) and machine.certSANs (Talos apid certificate). Both SANs are needed so that clients
// that connect via the FQDN after a cold restart receive a certificate that includes that hostname
// and can validate TLS without re-pointing or regenerating credentials.
func patchConfigYAML(endpoint, cpFQDN string) string {
	yaml := fmt.Sprintf("cluster:\n  controlPlane:\n    endpoint: %s\n", endpoint)

	if cpFQDN != "" {
		yaml += fmt.Sprintf("  apiServer:\n    certSANs:\n      - %s\nmachine:\n  certSANs:\n    - %s\n", cpFQDN, cpFQDN)
	}

	return yaml
}

// patchConfig rewrites cluster.controlPlane.endpoint in an already-generated machine config.
// When cpFQDN is non-empty (hostname-endpoint feature), it also adds the FQDN to
// cluster.apiServer.certSANs and machine.certSANs (see patchConfigYAML for rationale).
func patchConfig(cfg config.Provider, endpoint, cpFQDN string) ([]byte, error) {
	patches, err := configpatcher.LoadPatches([]string{patchConfigYAML(endpoint, cpFQDN)})
	if err != nil {
		return nil, err
	}

	out, err := configpatcher.Apply(configpatcher.WithConfig(cfg), patches)
	if err != nil {
		return nil, err
	}

	return out.Bytes()
}

// applyConfigTimeout bounds retries while the node's maintenance apid becomes reachable.
const applyConfigTimeout = 60 * time.Second

// applyConfig applies a machine config to a maintenance-mode node by re-execing talosctl
// (request.SelfExecutable) — the same re-exec convention the in-tree qemu provider uses. An
// in-process machinery client (client.ApplyConfiguration in maintenance mode) is a cleaner
// upstream refinement; the exec path is used here because it is already verified end-to-end (G4)
// and consistent with the provider exec-ing the `container` CLI.
func (p *provisioner) applyConfig(ctx context.Context, talosctl string, ip netip.Addr, cfg []byte) error {
	if talosctl == "" {
		talosctl = "talosctl"
	}

	f, err := os.CreateTemp("", "aegis-talos-cfg-*.yaml")
	if err != nil {
		return err
	}

	defer os.Remove(f.Name()) //nolint:errcheck

	if _, err = f.Write(cfg); err != nil {
		f.Close() //nolint:errcheck

		return err
	}

	if err = f.Close(); err != nil {
		return err
	}

	deadline := time.Now().Add(applyConfigTimeout)

	for {
		cmd := exec.CommandContext(ctx, talosctl, "apply-config", "--insecure", "--nodes", ip.String(), "--file", f.Name())

		out, runErr := cmd.CombinedOutput()
		if runErr == nil {
			return nil
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("%w: %s", runErr, out)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}
