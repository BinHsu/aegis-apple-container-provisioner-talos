# Runbook — reproduce the spike from a clean machine

A forker should be able to replay every manual step from here.

**Discipline: only verified steps live in this file.** A step appears once it has actually
been run and its output checked (cross-referenced in `VERIFICATION.md`). Un-run gates are
stubs — we do **not** pre-write unverified procedure. A runbook of untested commands is the
AI-comprehensive-without-verification trap this spike explicitly avoids.

## Requirements (a forker needs these)

| Tool | Version (tested) | For | When |
|---|---|---|---|
| macOS + Apple Silicon | 26.5.1 / arm64 | the whole runtime (Virtualization.framework) | always |
| Apple `container` | 1.0.0 | exec target — runs Talos nodes as micro-VMs | G0+ |
| └ default kernel | kata-containers 3.28.0 arm64 (`--recommended`) | the guest VM kernel | set at G0 |
| `talosctl` | v1.13.3 | gen-config / apply-config / bootstrap / health | G4+ |
| Talos node image | `ghcr.io/siderolabs/talos:v1.13.3` | the node OS — pin to `talosctl` | G1+ |
| `kubectl` | any recent | G4 acceptance — nodes `Ready` | G4 |
| `go` | 1.26.3 | build the aegis provider | G5 |
| `siderolabs/talos` Go module | matches v1.13.3 | compile against `pkg/provision` | G5 |
| `golangci-lint` | bundles gocyclo/gocognit/cyclop/funlen/maintidx | provider lint + complexity/BVA gates | G5 |
| `jq` | any | parse `container` / `talosctl` JSON | G3+ |
| OrbStack (or Docker Desktop / Colima) | — | Linux Docker socket the `orbstack + talos` fallback rides on (`talosctl cluster create --provisioner docker`) | fallback only |

> `kind` is **not** part of this fallback — it is kubeadm-in-Docker, not Talos. It appears in the
> spike only as a shared-kernel comparison point, never as a substrate to keep.

## G0 — install Apple `container` (verified procedure)

Official signed pkg from the apple/container release. **Vet the signature before installing.**

```bash
cd /tmp
curl -fsSL -o container-1.0.0-installer-signed.pkg \
  https://github.com/apple/container/releases/download/1.0.0/container-1.0.0-installer-signed.pkg

# Supply-chain gate — expect: "Apple Inc. - Containerization", notarized
pkgutil --check-signature container-1.0.0-installer-signed.pkg

# Install (system service → needs sudo / password)
sudo installer -pkg container-1.0.0-installer-signed.pkg -target /

container --version                                      # 1.0.0
# No default kernel ships. `container system start` prompts [Y/n] to download one
# (interactive — fails in a headless/no-tty session). Use the non-interactive flag:
container system kernel set --recommended                # downloads kata-containers 3.28.0 arm64
container run --rm docker.io/library/alpine echo ok      # smoke — prints: ok
```

Signature **verified 2026-06-13** → `Developer ID Installer: Apple Inc. - Containerization
(UPBK2H6LZM)`, notarized, timestamp 2026-06-09. Install (sudo, by Bin) + `kernel set
--recommended` + smoke → **`ok`, G0 PASSED 2026-06-13**. The default kernel is
**kata-containers 3.28.0** — empirical confirmation of the Kata-derived-kernel premise, and the
exact kernel G1 must inspect.

## G1 — kernel feature matrix (verified 2026-06-13)

One ephemeral VM on the default kata kernel; read its feature set. `/proc/config.gz` is present,
so the kernel config is the authoritative source (no guessing from module probes).

```bash
container run --rm docker.io/library/alpine sh -c '
  uname -a                                  # kernel version
  cat /proc/filesystems                     # overlay present?
  cat /sys/fs/cgroup/cgroup.controllers     # exists => cgroup v2 unified
  zcat /proc/config.gz | grep -E \
    "CONFIG_(OVERLAY_FS|BRIDGE_NETFILTER|BRIDGE=|NF_CONNTRACK|NF_TABLES|IP_NF_IPTABLES|VXLAN|VETH|NF_NAT)"
'
```

**Result — kernel `6.18.15` (kata 3.28.0 default), every k8s-required feature built-in (`=y`):**

| Feature | Talos/k8s needs it for | Verdict |
|---|---|---|
| `overlay` | container rootfs (overlayfs) | ✓ `/proc/filesystems` + `CONFIG_OVERLAY_FS=y` (+redirect_dir, index, metacopy) |
| cgroup v2 unified | kubelet resource control | ✓ mounted; controllers: `cpuset cpu io memory hugetlb pids` |
| `CONFIG_BRIDGE` + `BRIDGE_NETFILTER` | pod bridge + `br_netfilter` | ✓ `=y` |
| `NF_CONNTRACK` | kube-proxy / NAT conntrack | ✓ `=y` |
| `NF_TABLES` (+inet/ipv4/ipv6/bridge) | nftables backend | ✓ `=y` |
| `IP_NF_IPTABLES` (+LEGACY) | iptables backend | ✓ `=y` |
| `NF_NAT` (+masquerade) | service NAT / SNAT | ✓ `=y` |
| `VXLAN` | CNI overlay (flannel/cilium vxlan) | ✓ `=y` |
| `VETH` | pod veth pairs | ✓ `=y` |

All built-in, **nothing modular** (`/proc/modules` empty) — so the guest never needs `modprobe`,
sidestepping the "can an unprivileged micro-VM load `br_netfilter`?" failure mode that bites
minimal kernels. **G1 PASS → G2.**

## G2 — machined under vminitd (verified 2026-06-13)

Run the pinned Talos image directly; watch whether `machined` (Talos PID1 normally) tolerates being
a child of Apple's `vminitd`. Two runs — the contrast IS the finding.

```bash
# (a) default, unprivileged — FAILS at a privilege wall
container run --rm --name g2 ghcr.io/siderolabs/talos:v1.13.3
# (b) with full Linux capabilities — machined comes up clean
container run --rm --name g2-caps --cap-add ALL ghcr.io/siderolabs/talos:v1.13.3
container stop g2-caps        # machined idles waiting for config; stop to tear down
```

**Result:**
- `machined` **tolerates not being PID 1** — Talos detects `"mode": "container"`, runs early-startup,
  `phase machined`, starts containerd/machined/apid. The PID1 unknown is answered: **not a wall.**
- **(a) unprivileged hits a privilege wall** — fatal:
  `unix.Fsopen fstype="tmpfs" failed: operation not permitted` (needs CAP_SYS_ADMIN), plus containerd
  restart-loop: `write /proc/N/oom_score_adj: permission denied` (CAP_SYS_RESOURCE) and cgroup-remove
  failures. This is the apple/container micro-VM's no-`Privileged` model, empirically, at the syscall layer.
- **(b) `--cap-add ALL` clears all of it** — no fatal; controller-runtime comes up fully (resolvers,
  time servers, `iptables-nft`/`KUBE-IPTABLES-HINT`, nftables chains), containerd stable (PID 10),
  `machined` health-check OK, then idles waiting for `apply-config`. Only non-fatal noise: ethtool
  netlink unavailable on the container NIC.

**G2 PASS (with `--cap-add ALL`) → G3.** Design consequence for G5: the aegis provider must launch
nodes with `--cap-add ALL` — the apple/container analog of the docker provisioner's `Privileged: true`.
## G3 — networking (verified 2026-06-13)

Default network `default` = `192.168.64.0/24` (vmnet plugin); every container gets an IP automatically.

```bash
container run -d --name n1 docker.io/library/alpine sleep 600
container run -d --name n2 docker.io/library/alpine sleep 600
container ls                                   # IP column: n1 .6, n2 .7
container inspect n1 | jq '.[0].status.networks[0]'   # ipv4Address/Gateway, mac, mtu 1280

# cross-container reachability (listener in n2, probe from n1)
container exec -d n2 sh -c 'while true; do echo OKPONG | nc -l -p 6443 2>/dev/null; done'
container exec n1 sh -c 'echo hi | nc -w3 192.168.64.7 6443'   # -> OKPONG
container exec n1 ping -c2 192.168.64.7                        # -> 0% loss, ~0.6ms

# IP stability across restart
container stop n1 && container start n1
container inspect n1 | jq -r '.[0].status.networks[0].ipv4Address'   # CHANGED .6 -> .8
```

**Result:**
- **Per-node IPs ✓** — auto-assigned from `192.168.64.0/24`, each with ipv4/ipv6/MAC, gateway `.1`.
- **Cross-node reachability ✓** — TCP `:6443` returned `OKPONG`; ICMP 2/2, 0% loss, ~0.6 ms.
- **IP NOT stable across stop/start ✗** — restart moved n1 `.6 → .8` (next-free, old lease released).
  No `--ip` static flag on `run`/`create`; `container network create` has `--subnet` but no per-container
  reservation. **Pinning MAC (`--network default,mac=...`) does NOT help** — MAC held across restart but
  IP still moved `.9 → .10`; the vmnet DHCP does not reserve by MAC. Hypothesis tested and disproven.

**G3 PASS for cluster bring-up → G4** (per-node IPs + reachability are what G4 needs). **Documented
limitation:** dynamic IP, unstable across cold restart. Consequence for G5 + the blog: the provider must
capture each node's IP *after* launch (mirrors the docker provider) and treat cold-restart IP change as a
known gap — a candidate apple/container feature request (static IP / DHCP reservation), not a Talos bug.
## G4 — manual five-step cluster (verified 2026-06-13) ✅ FULLY GREEN

A 2-node Talos cluster (1 control-plane + 1 worker) reaches all-Ready on apple/container, by hand.
**This is the known-good recipe the G5 provider automates.** The three non-obvious launch requirements
are below — a vanilla `container run` of the Talos image does NOT boot a working node.

### Node launch recipe (the load-bearing part)

```bash
# control-plane: needs >= ~2GB (apiserver requests 512Mi; a 1GB node OOM-kills it silently)
container run -d --name talos-cp --cap-add ALL -m 4096MB \
  --tmpfs /run --tmpfs /tmp --tmpfs /system --tmpfs /system/state \
  --tmpfs /var --tmpfs /etc/cni --tmpfs /etc/kubernetes --tmpfs /usr/libexec/kubernetes \
  ghcr.io/siderolabs/talos:v1.13.3
# worker: 2GB is plenty
container run -d --name talos-worker --cap-add ALL -m 2048MB \
  --tmpfs /run --tmpfs /tmp --tmpfs /system --tmpfs /system/state \
  --tmpfs /var --tmpfs /etc/cni --tmpfs /etc/kubernetes --tmpfs /usr/libexec/kubernetes \
  ghcr.io/siderolabs/talos:v1.13.3
```

Three requirements, each learned from a failure (see `VERIFICATION.md` G4):
1. **`--cap-add ALL`** — without it, `machined` dies on `fsopen(tmpfs)` EPERM (G2).
2. **tmpfs on `/run /tmp /system /system/state /var /etc/cni /etc/kubernetes /usr/libexec/kubernetes`** —
   makes Talos's `setupSharedFilesystems` propagation targets real mount points (else EINVAL, G2/G4).
   **Do NOT tmpfs `/opt`** — it shadows the image's shipped `/opt/cni/bin`, and CNI sandbox creation then
   fails with `failed to find plugin "flannel"/"loopback"` (coredns stuck `ContainerCreating`).
3. **control-plane memory >= ~2GB** — on a 1GB node the `kube-apiserver` static pod (512Mi request) is
   OOM-killed at create with no log and never appears; CM/scheduler then CrashLoop on `127.0.0.1:7445` EOF.

### Five steps

```bash
CP_IP=$(container inspect talos-cp | jq -r '.[0].status.networks[0].ipv4Address' | cut -d/ -f1)
W_IP=$(container inspect talos-worker | jq -r '.[0].status.networks[0].ipv4Address' | cut -d/ -f1)

talosctl gen config aegis https://$CP_IP:6443 --output-dir /tmp/talos-g4 --force   # 1. config + PKI
export TALOSCONFIG=/tmp/talos-g4/talosconfig
talosctl apply-config --insecure -n $CP_IP -f /tmp/talos-g4/controlplane.yaml      # 2. cp  (maintenance apid :50000)
talosctl apply-config --insecure -n $W_IP  -f /tmp/talos-g4/worker.yaml            # 3. worker
talosctl config endpoint $CP_IP && talosctl config node $CP_IP
talosctl -n $CP_IP bootstrap                                                       # 4. bootstrap etcd
talosctl -n $CP_IP kubeconfig /tmp/talos-g4/kubeconfig --force                     # 5. kubeconfig
```

### Acceptance — all three met

```bash
talosctl -n $CP_IP health --control-plane-nodes $CP_IP --worker-nodes $W_IP   # all green
KUBECONFIG=/tmp/talos-g4/kubeconfig kubectl get nodes -o wide                  # both Ready
# teardown -> clean
container stop talos-cp talos-worker && container rm talos-cp talos-worker
container ls -a    # EMPTY — no orphan VMs; default network intact
```

**Verified result:** both nodes `Ready` (v1.36.1, Talos v1.13.3, kernel 6.18.15, containerd 2.2.4); all
control-plane + flannel + kube-proxy + coredns pods `1/1 Running`; `talosctl health` green; teardown leaves
`container ls -a` empty. **G4 PASS → G5.** The hypothesis holds: apple/container runs a real Talos cluster.
## G5 — aegis provider build + run — STUB

Each STUB is filled with the real, reproduced commands as its gate runs — never before.
