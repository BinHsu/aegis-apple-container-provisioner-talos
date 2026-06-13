# Upstream PR delta — `apple-container` provisioner for `siderolabs/talos`

This directory is the **exact set of changes** that add an Apple `container` provisioner to
talosctl, laid out at their in-tree paths so the merge is a mechanical copy. Everything here was
**built and run inside a real `talos` v1.13.3 checkout** — `talosctl cluster create apple-container`
brought up a fully healthy 2-node cluster and `talosctl cluster destroy` tore it down clean (see
`docs/VERIFICATION.md`, G5/upstream entry). These files reference talos `internal/` packages, so
they only compile inside the talos tree — that is why they live here as the delta, not in this
repo's own module.

## What goes where (directory move)

| This repo | talos path |
|---|---|
| `provider/apple/*.go` (the provisioner, the repo's module) | `pkg/provision/providers/apple/` |
| `upstream/.../create/cmd_apple.go` | `cmd/talosctl/cmd/mgmt/cluster/create/cmd_apple.go` |
| `upstream/.../create/create_apple.go` | `cmd/talosctl/cmd/mgmt/cluster/create/create_apple.go` |
| `upstream/.../create/clusterops/apple.go` | `.../create/clusterops/apple.go` |
| `upstream/.../create/clusterops/configmaker/apple.go` | `.../configmaker/apple.go` |
| `upstream/.../create/clusterops/configmaker/internal/makers/apple.go` | `.../internal/makers/apple.go` |
| `upstream/pkg/provision/providers/factory.go.diff` | applied to `pkg/provision/providers/factory.go` |

The provider package (`provider/apple/`) is the only piece that compiles standalone in this repo
(against the real `pkg/provision` interface — the "directory move, not rewrite" proof). The
cmd-layer files plug into talosctl's internal config-maker framework.

## To reproduce the PR

```bash
git clone --branch v1.13.3 https://github.com/siderolabs/talos
cp -r <this-repo>/provider/apple talos/pkg/provision/providers/apple
cp -r <this-repo>/upstream/cmd talos/cmd
cp <this-repo>/upstream/.../clusterops/... talos/...           # per the table above
( cd talos && git apply <this-repo>/upstream/pkg/provision/providers/factory.go.diff )
( cd talos && go build ./cmd/talosctl )
./talosctl cluster create apple-container --memory-controlplanes 4GiB
```

## Design notes for reviewers

- **Self-contained DHCP reconciliation, no framework change.** apple/container assigns node IPs via
  vmnet DHCP (no static `--ip`), so the provider cannot bake the config in at launch the way the
  docker provider does (USERDATA). Instead `Create` launches nodes bare into maintenance mode,
  discovers IPs, patches `cluster.controlPlane.endpoint`, and applies config over the maintenance
  API. `pkg/provision` is untouched — the divergence lives entirely in the provider.
- **No host port-mapping.** vmnet node IPs are reachable from the host directly, so unlike docker
  there is no localhost port-forwarding; `postCreate` bootstraps against the discovered node IPs.
- **Launch recipe** (learned empirically, see `docs/runbook.md` G1–G4): `--cap-add ALL`; tmpfs the
  propagation/runtime paths **except `/opt`** (a tmpfs would shadow the image's `/opt/cni/bin`,
  unlike a docker volume which copies up); control-plane memory ≥ ~2GB.
- **Open follow-up (a discussion point, not required):** an in-process machinery-client
  `ApplyConfiguration` would be cleaner than re-exec-ing talosctl; and an upstream "dynamic-IP
  provider" affordance would let DHCP providers avoid the post-launch patch entirely.
