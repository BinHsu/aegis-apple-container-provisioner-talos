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
| `kind` + OrbStack/docker | — | fallback substrate (`orbstack + talos`) | fallback only |

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

## G1 — kernel feature matrix — STUB
> Fill with the actual `container run` + in-VM inspection commands once executed and verified.

## G2 — machined under vminitd — STUB
## G3 — networking — STUB
## G4 — manual five-step cluster — STUB
## G5 — aegis provider build + run — STUB

Each STUB is filled with the real, reproduced commands as its gate runs — never before.
