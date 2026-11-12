---
title: "docs(portfolio): project page on paulperigault.fr"
labels: [type:doc, pkg:docs, taille:trait, prio:basse]
milestone: "v0.2.0 Web playground (Rust/WASM + Astro)"
assignee: PaulPerigault
---

## Context

The portfolio lists featured repositories from GitHub at build time and links project sites through `SITE.projectSites`.

## Goal

rsa-ct-analyzer appears in the portfolio with a French and an English description and a link to the playground.

## Acceptance criteria

- [ ] The repository is added to `projects-config.json` with `descriptions` in both languages.
- [ ] `SITE.projectSites` maps the repository to the playground URL.
- [ ] The portfolio `verify` and e2e pass on its side (a PR in `PaulPerigault/portfolio`).

## Package

docs

## Size

trait

## Priority

basse

## Milestone

v0.2.0 Web playground (Rust/WASM + Astro)

## Out of scope

Any change to the portfolio design.

## Notes

Cross-repository: the PR is opened in the portfolio, this issue tracks it.

## Definition of done

- [ ] Tests added, `task verify` green.
- [ ] Documentation up to date (README, ADR or journal).
- [ ] PR linked with `Closes #n`.
