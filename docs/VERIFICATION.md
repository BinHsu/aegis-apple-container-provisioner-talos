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

## G0 smoke (`container run ... echo ok`) — PENDING (Bin)
> Expected: `ok`. Record what actually printed, and anything about boot time / first-run
> kernel download that surprised you.

## G1 kernel feature matrix — PENDING
## G2 machined under vminitd — PENDING
## G3 networking — PENDING
## G4 manual five-step cluster — PENDING
## G5 aegis provider — PENDING

Fill each first-person as the gate runs. Surprises and dead-ends are the most valuable
entries — they are what a reviewer reads as a human having actually done the work.
