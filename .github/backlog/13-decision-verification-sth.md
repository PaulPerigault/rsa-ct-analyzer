---
title: "docs(ct): decide whether to verify tree head signatures"
labels: [type:decision, pkg:ct, taille:trait, prio:basse]
milestone: "v0.3.0 Real CT logs & hardening"
assignee: PaulPerigault
---

## Context

`crawl` does not verify the Signed Tree Head or SCTs: it audits keys, not the log. Some users will want assurance that the entries come from the log they named.

## Goal

An ADR that states the options (do nothing, verify the STH with the log key, verify inclusion proofs), the cost of each, and the decision.

## Acceptance criteria

- [ ] The ADR is written in `docs/adr/` and linked from `docs/decisions-log.md`.
- [ ] If the decision is to implement, a `feat` issue is created with acceptance criteria.
- [ ] The README "Limits" section matches the decision.

## Package

ct

## Size

trait

## Priority

basse

## Milestone

v0.3.0 Real CT logs & hardening

## Out of scope

Implementation.

## Definition of done

- [ ] Tests added, `task verify` green.
- [ ] Documentation up to date (README, ADR or journal).
- [ ] PR linked with `Closes #n`.
