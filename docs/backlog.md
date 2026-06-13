# Backlog — toward enforced GitOps / DevSecOps / FinOps

**Principle: anything a tool can enforce must not live in a prompt, instruction, or memory.**
A documented rule drifts and is invisible to a reviewer; a CI gate or hook is mechanical and
auditable — a *hard* signal, not a *soft* one. Each item below is a **gate**, not a guideline.
(Sharpens the CLAUDE.md pain-driven hook-upgrade pattern: toolify proactively where the tool is
cheap, not only after the lapse.)

Status (2026-06-13): the provider exists (G5 done), so the code gates **landed**:
- **CI** (`.github/workflows/ci.yml`): build + vet + `go test -race` (the BVA suites) + `golangci-lint`
  + `govulncheck` + gitleaks, on push/PR. Verified locally green (lint 0 issues, tests pass,
  govulncheck clean after the Go 1.26.4 bump that closed GO-2026-5037/5039).
- **golangci-lint** (`.golangci.yml`): the complexity/length linters below, generous thresholds.
- **Pre-commit** (`lefthook.yml`): local mirror (gofmt + golangci-lint pre-commit; vet + test pre-push),
  linters via `go run` (non-host-mutating).
Already enforced server-side (via `gh`): secret scanning + push protection, Dependabot.
Still backlog: SHA-pin action versions (note in ci.yml), CodeQL, release flow, branch protection
(deferred until pushed/PR'd).

## Pre-commit hooks (local gate — fail before a bad commit leaves the machine)
- `gofmt` / `goimports` — formatting, zero debate
- `golangci-lint` — vet + complexity/BVA gates (gocyclo, gocognit, cyclop, funlen, maintidx)
- `gitleaks` — block secrets locally (defence in depth with server push protection)
- conventional-commit message lint
- Tooling: pinned, repo-local (pre-commit framework or lefthook) — non-host-mutating

## GitHub Actions CI (remote gate — DevSecOps)
- **lint:** golangci-lint + `actionlint`
- **test:** `go test ./...` (the BVA suites)
- **build:** `go build` against the real `siderolabs/talos` `pkg/provision` — the compile *is*
  the proof of "directory move, not rewrite"
- **security:** `govulncheck`, `gitleaks`, CodeQL (Go), `dependency-review` on PRs
- **supply-chain:** pin action SHAs, OIDC over long-lived secrets, SBOM (`syft`) on release
- Note: CI workflows are also the gate that justifies branch protection (below)

## GitOps
- The spike's *subject* is a Talos provisioner. For THIS repo's own delivery, `runbook.md` +
  `VERIFICATION.md` are the reproducibility contract today. A release flow (tag → CI build →
  signed artifact) lands here once there is an artifact to ship.

## FinOps
- This spike is local, **zero cloud spend by design** — FinOps is N/A here. Placeholder for any
  future gate that touches cloud: `infracost` on IaC PRs, budget alarms, and teardown
  verification (no orphan billable resources — already a G5 acceptance criterion).

## Branch protection (post-CI governance)
- `main`: block force-push + deletion; require the CI status checks above once they exist.
  Deferred until CI exists, so it doesn't gate a solo direct-push spike prematurely (and would
  have blocked the 2026-06-13 history scrub).
