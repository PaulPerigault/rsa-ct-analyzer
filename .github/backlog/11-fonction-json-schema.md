---
title: "feat(report): versioned JSON Schema for the report"
labels: [type:fonction, pkg:report, taille:page, prio:normale]
milestone: "v0.3.0 Real CT logs & hardening"
assignee: PaulPerigault
---

## Context

CI users parse the JSON report. Its shape must be a contract, not an accident.

## Goal

A `schema_version` field and a published JSON Schema, with a test that fails when the output stops matching the schema.

## Acceptance criteria

- [ ] `rsa-ct-analyzer schema` prints the JSON Schema of the report.
- [ ] The report carries `schema_version`; a breaking change bumps it.
- [ ] A test validates real reports (from `testdata/`) against the schema.
- [ ] The README documents the compatibility rule.

## Package

report

## Size

page

## Priority

normale

## Milestone

v0.3.0 Real CT logs & hardening

## Out of scope

A schema for the table output.

## Notes

Finding IDs are already stable (`analyze.ID*`).

## Definition of done

- [ ] Tests added, `task verify` green.
- [ ] Documentation up to date (README, ADR or journal).
- [ ] PR linked with `Closes #n`.
