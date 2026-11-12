---
title: "feat(web): WebAssembly island loaded on demand"
labels: [type:fonction, pkg:web, taille:page, prio:normale]
milestone: "v0.2.0 Web playground (Rust/WASM + Astro)"
assignee: PaulPerigault
---

## Context

Loading a `.wasm` on every visit would break the zero-JavaScript rule and cost bandwidth for readers who only want to read.

## Goal

An analysis form whose script and `.wasm` load only after an explicit action (button or first focus).

## Acceptance criteria

- [ ] No script and no `.wasm` request before the action (checked by a Playwright test on the network log).
- [ ] Pasted data never leaves the browser (no request carries it; checked by a test).
- [ ] Findings shown match the Go goldens for the same input.
- [ ] Errors (bad PEM, size limit) are announced in a `role="status"` region.

## Package

web

## Size

page

## Priority

normale

## Milestone

v0.2.0 Web playground (Rust/WASM + Astro)

## Out of scope

Uploading files larger than the documented limit.

## Notes

Depends on the WebAssembly build and the Astro site.

## Definition of done

- [ ] Tests added, `task verify` green.
- [ ] Documentation up to date (README, ADR or journal).
- [ ] PR linked with `Closes #n`.
