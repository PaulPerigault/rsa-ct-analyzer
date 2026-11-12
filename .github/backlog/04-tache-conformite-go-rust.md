---
title: "test(wasm): Go and Rust conformance on golden files"
labels: [type:tache, pkg:wasm, taille:page, prio:haute]
milestone: "v0.2.0 Web playground (Rust/WASM + Astro)"
assignee: PaulPerigault
---

## Context

Two implementations of the same checks can drift apart silently. ADR 0004 names this as the main risk of the Rust choice.

## Goal

Golden files derived from `testdata/` that both implementations must reproduce byte for byte.

## Acceptance criteria

- [ ] `task data:golden` writes one JSON golden per sample file from the Go analysis, with fixed options.
- [ ] A Go test and a Rust test each read the goldens and fail on any difference in findings, severities or evidence.
- [ ] CI fails if the goldens are stale, as `task data:check` does for `testdata/`.
- [ ] Adding a finding in Go without the Rust side fails the build.

## Package

wasm

## Size

page

## Priority

haute

## Milestone

v0.2.0 Web playground (Rust/WASM + Astro)

## Out of scope

Fuzzing both implementations against each other (follow-up).

## Notes

Depends on the Rust crate issue.

## Definition of done

- [ ] Tests added, `task verify` green.
- [ ] Documentation up to date (README, ADR or journal).
- [ ] PR linked with `Closes #n`.
