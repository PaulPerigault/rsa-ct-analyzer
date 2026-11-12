---
title: "feat(ct): resumable crawl of a real log with a bounded rate"
labels: [type:fonction, pkg:ct, taille:page, prio:haute]
milestone: "v0.3.0 Real CT logs & hardening"
assignee: PaulPerigault
---

## Context

Crawling millions of entries from a public log takes hours. Today an interruption loses everything and the rate is a single fixed limit.

## Goal

A checkpoint file recording the next index so that `crawl` resumes, plus a rate that adapts to HTTP 429 without exceeding the configured ceiling.

## Acceptance criteria

- [ ] `crawl --resume <file>` continues where it stopped; the output is appended without duplicates.
- [ ] A 429 or a `Retry-After` lowers the rate; it never exceeds `--rate-limit`.
- [ ] Tested against the replay server of `testdata/ct/` with injected failures; the public network is never used in tests.
- [ ] The README documents a bounded real-log run (log URL, count, expected duration).

## Package

ct

## Size

page

## Priority

haute

## Milestone

v0.3.0 Real CT logs & hardening

## Out of scope

Verifying signed tree heads (separate decision issue).

## Notes

Uses the CT fixtures added in v0.1.0.

## Definition of done

- [ ] Tests added, `task verify` green.
- [ ] Documentation up to date (README, ADR or journal).
- [ ] PR linked with `Closes #n`.
