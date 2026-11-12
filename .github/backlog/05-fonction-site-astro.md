---
title: "feat(web): Astro site on the design-system tokens"
labels: [type:fonction, pkg:web, taille:page, prio:haute]
milestone: "v0.2.0 Web playground (Rust/WASM + Astro)"
assignee: PaulPerigault
---

## Context

The project needs a public page, consistent with the rest of paulperigault.fr, that does not reinvent a visual identity.

## Goal

A static Astro site in `web/` that consumes `@paulperigault/css` and `@paulperigault/tokens` as published, with zero JavaScript by default.

## Acceptance criteria

- [ ] `web/` builds with Node 24 and `npm ci`; the output is static.
- [ ] No raw colour, size or radius: only `--pp-*` tokens; no inline `style` attribute.
- [ ] The page works with JavaScript disabled (it explains the tool and shows a recorded example).
- [ ] Lighthouse budgets are written and checked in CI.
- [ ] The strict CSP uses build-generated hashes, never `unsafe-inline`.

## Package

web

## Size

page

## Priority

haute

## Milestone

v0.2.0 Web playground (Rust/WASM + Astro)

## Out of scope

The interactive island (next issue), deployment.

## Notes

Depends on the design-system packages being published.

## Definition of done

- [ ] Tests added, `task verify` green.
- [ ] Documentation up to date (README, ADR or journal).
- [ ] PR linked with `Closes #n`.
