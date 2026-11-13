# ADR 0006: release-please opens a draft release PR; the owner decides when a release exists

Status: accepted (2026-11-13)

## Context

`release-please.yml` runs on every push to `main` and opens the release PR (`CHANGELOG.md`, `.release-please-manifest.json`).
Merging that PR creates the tag `vX.Y.Z`, which triggers `release.yml`: it builds five binaries, writes `SHA256SUMS`, attests the
provenance and creates the GitHub release. A release is public and cannot be taken back cleanly, so it stays the owner's decision.

Two facts shape the setup:

- The manifest starts at `0.0.0` and only `feat`, `fix` and breaking commits are releasable. `chore`, `ci`, `docs` and
  `refactor` commits open nothing, so the first release PR does not appear on its own.
- A pull request opened with the default `GITHUB_TOKEN` does not start other workflows, so the required check `ci` does not run on
  the release PR until a person reopens it (the `reopened` event is in `ci.yml`).

## Decision

- `release-please-config.json` sets `draft-pull-request: true`: the release PR is a draft and cannot be merged (nor auto-merged)
  before the owner marks it ready. Nothing is published until then.
- The first version is requested by a `Release-As: 0.1.0` footer in the body of a squashed commit, so the milestone
  v0.1.0 of the README is the first release. The footer is written once, in the body of one pull request.
- To get `ci` on the release PR: close and reopen it, then mark it ready, then merge with squash. No token or secret is added.
- `bump-minor-pre-major` stays on: before 1.0.0 a `feat` raises the minor version, a breaking change too.

## Consequences

- One manual step (reopen) per release PR. It is accepted instead of storing a personal access token.
- The workflow, the actions (pinned by SHA) and the permissions (`contents: write`, `pull-requests: write` on the one job) are unchanged.

## Alternatives rejected

- A personal access token or a GitHub App to make `ci` run on the release PR: a long-lived secret for a solo project.
- A repository variable that disables `release.yml`: needs a GitHub write and hides the state in the settings.
- `release-as` in the config file: it stays in force after the release and forces the same version again.

## Main risk and mitigation

Risk: the footer is forgotten and no release PR appears. Mitigation: the first fix or feature commit opens one at `0.0.1` or
`0.1.0`; the owner can then close it and use the footer.
