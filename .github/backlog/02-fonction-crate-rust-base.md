---
title: "feat(wasm): Rust crate with the base checks"
labels: [type:fonction, pkg:wasm, taille:page, prio:haute]
milestone: "v0.2.0 Web playground (Rust/WASM + Astro)"
assignee: PaulPerigault
---

## Context

The playground needs the same checks as the CLI, in a language that compiles to WebAssembly. The Go code cannot be reused in the browser without a large runtime.

## Goal

A `crates/rsa-core` Rust crate with key-size, small-factor, perfect-square, Fermat close-primes and batch GCD, with the same finding IDs and severities as `internal/analyze`.

## Acceptance criteria

- [ ] `cargo test` passes, including property tests (`proptest`) on batch GCD against the pairwise definition.
- [ ] Finding IDs, severities and evidence keys equal the Go ones (`analyze.ID*`).
- [ ] No `unsafe`, no network, no filesystem; `cargo clippy -- -D warnings` and `cargo fmt --check` pass.
- [ ] `rust-toolchain.toml` pins the toolchain and the `wasm32-unknown-unknown` target.
- [ ] `cargo deny` and `cargo audit` run in CI on the Rust paths.

## Package

wasm

## Size

page

## Priority

haute

## Milestone

v0.2.0 Web playground (Rust/WASM + Astro)

## Out of scope

The WebAssembly packaging (next issue) and the web UI.

## Notes

Blocked by the decision in ADR 0004.

## Definition of done

- [ ] Tests added, `task verify` green.
- [ ] Documentation up to date (README, ADR or journal).
- [ ] PR linked with `Closes #n`.
