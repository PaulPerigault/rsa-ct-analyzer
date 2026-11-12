---
title: "chore(infra): Cloudflare Pages in OpenTofu, as the portfolio"
labels: [type:tache, pkg:infra, taille:page, prio:normale]
milestone: "v0.2.0 Web playground (Rust/WASM + Astro)"
assignee: PaulPerigault
---

## Context

The portfolio is hosted on Cloudflare Pages with OpenTofu as the only write path. The playground follows the same method.

## Goal

An `infra/` directory with a Pages project, a custom domain and response headers, written, formatted and validated by Claude Code; applied by the owner only.

## Acceptance criteria

- [ ] `tofu fmt -check`, `tofu validate` and `tofu test` (mock provider) pass in CI.
- [ ] Provider and OpenTofu versions pinned, lock file committed, state outside the repository.
- [ ] Headers in `web/public/_headers`, checked against production by a script.
- [ ] `allowed-destroys.txt` exists and is empty; no `apply` from CI.
- [ ] An `owner` issue lists the manual steps (token, DNS, first apply).

## Package

infra

## Size

page

## Priority

normale

## Milestone

v0.2.0 Web playground (Rust/WASM + Astro)

## Out of scope

Applying the plan: owner only.

## Notes

Label `owner` for the apply step.

## Definition of done

- [ ] Tests added, `task verify` green.
- [ ] Documentation up to date (README, ADR or journal).
- [ ] PR linked with `Closes #n`.
