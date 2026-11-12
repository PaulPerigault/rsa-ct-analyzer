# ADR 0005: Agent skills from a master catalogue, profiles core + go + docs

Status: accepted (2026-11-12)

## Context

Coding agents load skills (method files). Third-party skills are code-adjacent instructions: they can
contradict the repository rules, call the network or widen tool permissions. Owner projects need one coherent
set instead of ad hoc copies per repository.

## Decision

- Skills come from a master catalogue kept in the owner's workspace, audited before vendoring (full read of
  every file, grep for `curl | sh`, `eval`, network calls, secret reads, wide `allowed-tools`, invisible
  Unicode, contradicting instructions, redistribution licence). A doubtful skill is refused.
- This repository installs the profiles `core`, `go` and `docs`: 13 skills, 8 auto-triggered (cap: 12), the
  others manual. They are copied unmodified into `.claude/skills/` with `.claude/skills.lock.json` (source,
  SHA, licence, hashes). Third-party licences are kept in `.claude/skills/_licenses/`.
- `task skills:verify` (part of `task verify` and CI) fails on any divergence from the lock or if more than 12
  skills are auto-triggered. The catalogue is not in this repository, so the check is against the lock.
- `ux-ui` is not installed now. It will be added with the web playground (v0.2.0), when there is UI to review.
- Order of authority: this repository's rules, then house skills, then `encre-*`, then third-party skills.

## Consequences

- A skill is never edited here; a fix goes to the catalogue, then `skills:sync`.
- Some skills advise things this repository forbids (`testify`, `go install @latest`): the rules win.

## Alternatives rejected

- Installing skills per developer machine: unverifiable, drifts.
- Installing every available profile: exceeds the 12 auto-triggered cap and dilutes the context.

## Main risk and mitigation

A vendored skill hides an instruction the audit missed. Mitigation: grep-based audit recorded in the catalogue,
manual trigger for anything not needed every day, and the rule that a skill is method, never authority.

## Update 2026-11-12: full read

Every file of the 13 installed skills was read completely (`audit_level: full-read` in the lock). Two skills were
removed: one instructs the agent to write into `CLAUDE.md`/`AGENTS.md` without confirmation, the other pre-approves
network, write and git tools on every Go file and cannot be narrowed because skills are vendored unmodified.
Details in the workspace `skills/AUDIT.md`, decisions D-019 and D-020.
