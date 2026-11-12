---
title: "test(web): zero axe violations on the playground"
labels: [type:accessibilite, pkg:web, taille:page, prio:normale]
milestone: "v0.2.0 Web playground (Rust/WASM + Astro)"
assignee: PaulPerigault
---

## Context

The portfolio and the design system enforce zero axe violations. The playground must keep that bar.

## Goal

An axe run on every page of `web/`, light and dark, desktop and mobile, with a canary proving that detection works.

## Acceptance criteria

- [ ] Zero violations on WCAG 2.0, 2.1 and 2.2 A and AA.
- [ ] Keyboard-only use of the form, visible focus, accessible names, results announced.
- [ ] Contrast read from the published tokens, 4.5:1 for text and 3:1 for controls.
- [ ] The run is part of `ci`.

## Package

web

## Size

page

## Priority

normale

## Milestone

v0.2.0 Web playground (Rust/WASM + Astro)

## Out of scope

Manual screen-reader audit (tracked separately).

## Notes

Depends on the Astro site.

## Definition of done

- [ ] Tests added, `task verify` green.
- [ ] Documentation up to date (README, ADR or journal).
- [ ] PR linked with `Closes #n`.
