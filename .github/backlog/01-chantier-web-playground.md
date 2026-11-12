---
title: "epic: web playground in Rust/WASM and Astro"
labels: [type:chantier, pkg:web, taille:cahier, prio:haute]
milestone: "v0.2.0 Web playground (Rust/WASM + Astro)"
assignee: PaulPerigault
---

## Context

Today the analysis runs only in the Go CLI. A browser playground lets someone paste a certificate or a list of moduli and see the findings without installing anything, and shows the project on paulperigault.fr.

## Goal

A static site, built with Astro on the @paulperigault/design-system tokens, that runs the core checks in the browser through a Rust crate compiled to WebAssembly, with results identical to the Go CLI.

## Sub-issues

- [ ] Rust crate with the base checks
- [ ] WebAssembly build and size budget
- [ ] Go vs Rust conformance on golden files
- [ ] Astro site on design-system tokens
- [ ] WASM island loaded on demand
- [ ] Accessibility: zero axe violations
- [ ] Cloudflare Pages in OpenTofu
- [ ] Project page on the portfolio

## Acceptance criteria

- [ ] The sub-issues below are closed.
- [ ] `web/` builds statically and serves with zero JavaScript by default.
- [ ] The WebAssembly island loads only on demand.
- [ ] Go and Rust agree on every golden file (see the conformance issue).
- [ ] The playground is deployed to Cloudflare Pages by OpenTofu and linked from the portfolio.

## Package

web

## Size

cahier

## Priority

haute

## Milestone

v0.2.0 Web playground (Rust/WASM + Astro)

## Out of scope

Server-side analysis, uploads, accounts, any storage of the pasted data.

## Notes

ADR 0004 (proposed) is accepted or rejected before the first sub-issue starts.

## Definition of done

- [ ] Tests added, `task verify` green.
- [ ] Documentation up to date (README, ADR or journal).
- [ ] PR linked with `Closes #n`.
