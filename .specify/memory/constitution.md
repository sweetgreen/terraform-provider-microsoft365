<!--
SpecKit project constitution for terraform-provider-microsoft365
(Sweetgreen fork of deploymenttheory/terraform-provider-microsoft365).
The Specify / Plan / Analyze stages read this file and check work against it.
-->

# Project Constitution

## Core Principles

### I. Clarity over cleverness
Code and specs are written to be read. Prefer the obvious solution; justify any
non-obvious one in the spec.

### II. Test the contract
Every feature defines how we will know it works (acceptance criteria) before it
is implemented. No feature is "done" without a checkable validation.

### III. Small, reversible steps
Ship the smallest change that delivers value. Avoid speculative generality.

### IV. Upstream first
Provider fixes belong upstream. A fix may land on the fork's `sg` line first to
unblock Sweetgreen, but each fork-only patch must have an upstream PR (or a
recorded reason why not). Upstream PRs branch from `upstream/main`, contain only
the change at hand, and follow upstream conventions (conventional-commit PR
titles enforced by `pr-title-validation.yml`, the upstream PR template).

### V. Graph behaviour is evidence, not assumption
Changes that depend on how Microsoft Graph answers a request (accepted
properties, status codes, null vs default values) must cite the observed
behaviour — a probe, an API response, or Microsoft documentation — in the spec
and in the PR description.

### VI. Unit tests with mocks prove the fix
Every bug fix adds a unit test that runs offline against the package's
`httpmock` responders (or a pure-function test) and fails without the fix.
Acceptance tests against a live tenant are optional and never the only proof.

### VII. Fork patches are retired after upstream release
When upstream releases a fix the fork carries, record the upstream version on
the tracking issue; the next `sg` rebase drops the fork patch.

## Constraints
- Follow the existing architecture and conventions of the repository.
- Do not introduce new top-level dependencies without recording the reason in
  the plan's research notes.
- Never change upstream behaviour beyond the bug being fixed in an upstream PR.

## Governance
This constitution supersedes ad-hoc preferences. When a spec or plan conflicts
with it, the constitution wins or the conflict is recorded as an open question.

**Version**: 1.0.0 (replaced the vibe-kanban-specific principles IV–VII with
provider principles; 2026-10-15)
