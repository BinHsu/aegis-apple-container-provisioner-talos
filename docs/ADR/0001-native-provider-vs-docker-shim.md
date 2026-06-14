# ADR 0001 — A native pkg/provision provider, not a Docker-API shim

**Status:** Accepted (2026-06-14)

## Context

Apple's `container` runtime exposes no Docker API. There are two ways to make Talos run on it:

- **(A) Native provider** — implement Talos's `pkg/provision.Provisioner` interface for
  apple/container (what this repo does). Merges into `siderolabs/talos` as a new provider.
- **(B) Docker-API shim** — put a Docker-compatible API in front of apple/container
  (socktainer-style) so Talos's existing `--provisioner docker` drives it unchanged.

(B) is attractive because it needs zero Talos changes. We chose (A) anyway.

## Decision

Build the native provider. The deciding factor is **how apple/container assigns IPs**, and **which
component owns the node-creation flow**.

apple/container assigns node IPs via vmnet **DHCP** — there is no static-IP option (verified, G3).
That single fact breaks the Docker route:

- The Talos **docker provisioner's contract is "static IP + config-at-create"**: the config maker
  computes each node's IP up front (`.2`, `.3`), the docker provider creates the container with that
  IP pinned (`IPAMConfig.IPv4Address`) and the machine config — which bakes that same IP into
  `cluster.controlPlane.endpoint` and the apiserver cert SANs — injected as the `USERDATA` env var.
  The node is expected to **boot already holding that IP**.
- A shim receiving "create container with IP `.2` and this `USERDATA`" **cannot honor the IP**
  (DHCP hands out, say, `.8`). The node boots with a config that says `.2` while its real address is
  `.8` → etcd/apiserver bind `.8`, certs/endpoint say `.2` → the cluster never forms. **This is the
  exact wall socktainer hits.**
- To rescue it, a shim would have to intercept the create call, defer the start, discover the DHCP
  address, and rewrite the base64 `USERDATA` config — but the docker provisioner's inject-at-create
  flow leaves **no clean seam** for that reconciliation.

The **native provider owns `Create`**, so it does the reconciliation cleanly: launch the node bare
into maintenance mode, discover its DHCP IP, patch `cluster.controlPlane.endpoint`, then apply the
config over the maintenance API. **That is precisely why the native provider works where the shim
does not** — the framework is never changed; the DHCP workaround lives entirely inside the provider.

## Other factors (all favour the native provider)

| | Native provider | Docker-API shim |
|---|---|---|
| DHCP reconciliation | owns `Create`, does it cleanly | no seam to insert it (the socktainer wall) |
| Maintenance surface | ~1000 lines against a **stable Go interface** | emulate a large slice of the **Docker API**, chase its changes |
| Upstream story | merges into Talos; verified via `talosctl cluster create apple-container` | a third-party apple/container tool, not a Talos contribution |
| Semantics | models the micro-VM reality directly (no host port-map, `/opt` not tmpfs, `--cap-add ALL`) | must fake shared-kernel Docker semantics everywhere |

The shim's only edge — zero Talos changes — does not pay off, because it does not actually work
cleanly and the Docker-API emulation is a heavier long-term burden than a small native provider.

## Honest correction

Early in the spike we attributed socktainer's failure to "apple/container has no privileged concept."
That is imprecise: `container run --cap-add ALL` **is** the `Privileged: true` equivalent and works
(verified, G2). The real blocker is the **static-IP + config-at-create contract**, not privileges.
This sharpens the rationale: the issue is *where the IP is decided and config is injected*, not capabilities.

## Consequence

The provider is `provider/apple/`; it implements the real interface and is verified end to end,
including a full networking stack (MetalLB host-reachable LoadBalancer; L7 via ingress-nginx and
Gateway API). The Docker-shim path is documented here as considered-and-rejected, with the reason.
