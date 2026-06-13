# Runbook — reproduce the spike from a clean machine

A forker should be able to replay every manual step from here.

**Discipline: only verified steps live in this file.** A step appears once it has actually
been run and its output checked (cross-referenced in `VERIFICATION.md`). Un-run gates are
stubs — we do **not** pre-write unverified procedure. A runbook of untested commands is the
AI-comprehensive-without-verification trap this spike explicitly avoids.

## Prerequisites (verified 2026-06-13)

- macOS 26+, Apple Silicon — here: macOS 26.5.1 / arm64
- `talosctl` v1.13.3 · `go` 1.26.3 · `kind` · OrbStack (docker) · `jq` — all present
- Pin the Talos node image to **v1.13.3** to match `talosctl`
- `kubectl` — verify present before G4 (`kubectl version --client`)

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

container --version
container system start
# Smoke — confirm exact run flags against `container run --help` on first use
container run --rm docker.io/library/alpine echo ok      # expect: ok
```

Signature **verified 2026-06-13** → `Developer ID Installer: Apple Inc. - Containerization
(UPBK2H6LZM)`, notarized, timestamp 2026-06-09 (matches 1.0.0 release). The sudo install +
smoke are **PENDING** (run by Bin — host-mutating, needs password).

## G1 — kernel feature matrix — STUB
> Fill with the actual `container run` + in-VM inspection commands once executed and verified.

## G2 — machined under vminitd — STUB
## G3 — networking — STUB
## G4 — manual five-step cluster — STUB
## G5 — aegis provider build + run — STUB

Each STUB is filled with the real, reproduced commands as its gate runs — never before.
