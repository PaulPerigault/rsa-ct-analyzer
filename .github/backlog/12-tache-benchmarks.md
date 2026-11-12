---
title: "perf(batchgcd): benchmarks and a scaling table"
labels: [type:tache, pkg:batchgcd, taille:page, prio:normale]
milestone: "v0.3.0 Real CT logs & hardening"
assignee: PaulPerigault
---

## Context

The README claims quasi-linear scaling. A claim needs a measure with the machine and the date.

## Goal

`go test -bench` benchmarks for batch GCD and for a full scan, and a results table in `docs/perf.md`.

## Acceptance criteria

- [ ] Benchmarks for 1 000, 10 000 and 100 000 moduli of 2048 bits.
- [ ] The table records CPU, Go version, date and the command.
- [ ] Memory use (`-benchmem`) is reported for the largest size.
- [ ] The README "Limits" section is updated with the measured numbers.

## Package

batchgcd

## Size

page

## Priority

normale

## Milestone

v0.3.0 Real CT logs & hardening

## Out of scope

Optimising the algorithm (follow-up if the numbers call for it).

## Definition of done

- [ ] Tests added, `task verify` green.
- [ ] Documentation up to date (README, ADR or journal).
- [ ] PR linked with `Closes #n`.
