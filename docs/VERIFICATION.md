# Verification log — first person: what was run, what was seen

The proof that a **human verified**, not just that artifacts exist. This is the committed,
forker-facing distillation of the loop's raw `_out/notes/` (gitignored). One entry per
verification event: *what I ran · what I expected · what I saw · what surprised me · verdict.*

**Don't pre-fill gates not yet run.** An entry exists only after the thing was actually
observed. Empty-but-claimed verification is the exact failure this spike is built to avoid.

---

## 2026-06-13 — `container` pkg supply-chain check ✅
- **Ran:** `pkgutil --check-signature container-1.0.0-installer-signed.pkg`
- **Expected:** an Apple-signed, notarized installer.
- **Saw:** `Developer ID Installer: Apple Inc. - Containerization (UPBK2H6LZM)`; "trusted by
  the Apple notary service"; trusted timestamp 2026-06-09 — matches the 1.0.0 release date.
- **Verdict:** legitimate official artifact, cleared to install.
- *(Performed in-session via Claude as operator, at Bin's direction. The hands-on gate
  verifications below are Bin's.)*

## 2026-06-13 — toolchain state ✅
- **Saw:** talosctl v1.13.3, go 1.26.3, kind, OrbStack (docker), jq `/usr/bin/jq` — present.
  `container`: pending install.

---

## 2026-06-13 — G0: container install + smoke ✅ PASSED
- **Ran:** (sudo install by Bin in his own terminal) `container --version` → 1.0.0; then
  `container system kernel set --recommended`; then `container run --rm docker.io/library/alpine echo ok`.
- **Expected:** `ok` from a booted micro-VM.
- **Saw:** image fetched + unpacked → init image (vminitd, ~64 MB) fetched → `[6/6] Starting
  container` → **`ok`**. The Virtualization.framework micro-VM boot path works on this machine.
- **Surprised me:** (1) no default kernel ships — `container system start`/`run` fails until one
  is set, and the prompt is interactive (no-tty headless fails); `--recommended` is the
  non-interactive path. (2) the default kernel is **kata-containers 3.28.0 arm64** — confirms the
  Kata-derived-kernel premise empirically. Carry into G1: this exact kernel's feature set is what
  G1 inspects.
- **Verdict:** G0 PASSED → current gate G1.

## G1 kernel feature matrix — PENDING
## G2 machined under vminitd — PENDING
## G3 networking — PENDING
## G4 manual five-step cluster — PENDING
## G5 aegis provider — PENDING

Fill each first-person as the gate runs. Surprises and dead-ends are the most valuable
entries — they are what a reviewer reads as a human having actually done the work.
