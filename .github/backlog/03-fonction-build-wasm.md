---
title: "feat(wasm): WebAssembly build with a size budget"
labels: [type:fonction, pkg:wasm, taille:page, prio:normale]
milestone: "v0.2.0 Web playground (Rust/WASM + Astro)"
assignee: PaulPerigault
---

## Context

The crate must reach the browser as a small `.wasm` with a typed JavaScript binding, built reproducibly in CI.

## Goal

A `task wasm:build` target producing the `.wasm` and its binding with pinned tools, and a CI check on the compressed size.

## Acceptance criteria

- [ ] The build is reproducible: two builds give the same hash.
- [ ] The compressed `.wasm` stays under the budget written in `docs/` (to measure first, then fix).
- [ ] The release profile uses `opt-level = "s"`, LTO and `panic = "abort"`.
- [ ] No tool is downloaded without a pinned version and checksum.

## Package

wasm

## Size

page

## Priority

normale

## Milestone

v0.2.0 Web playground (Rust/WASM + Astro)

## Out of scope

Publishing to npm: the playground consumes the file from the repository.

## Notes

Depends on the Rust crate issue.

## Definition of done

- [ ] Tests added, `task verify` green.
- [ ] Documentation up to date (README, ADR or journal).
- [ ] PR linked with `Closes #n`.
