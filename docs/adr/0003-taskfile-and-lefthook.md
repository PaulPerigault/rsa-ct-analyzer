# ADR 0003: Taskfile, Lefthook and tools pinned in `tools/go.mod`

Status: accepted (2026-11-12)

## Context

A `Makefile` needs GNU Make, which Windows does not have, and its recipes depend on the shell. Developer
tools (`golangci-lint`, `govulncheck`, `gitleaks`) were installed with `@latest` or from a download, so two
machines could run two versions with nothing verifying what was fetched.

## Decision

- **Task** (`Taskfile.yml`, v3) is the single entry point. Targets are normalised across the projects:
  `default setup dev verify test lint fmt build build:all demo security check:docs data:gen data:check
  docker:build docker:scan clean`, each with a `desc`. `verify` aggregates everything and is both the
  `pre-push` hook and the CI job.
- **Lefthook** installs the hooks: `pre-commit` (gitleaks on the staged files, gofmt, golangci-lint),
  `commit-msg` (Conventional Commits, no tool signature), `pre-push` (`task verify`).
- **Tools are Go `tool` directives in `tools/go.mod`**, with hashes in `tools/go.sum`: `golangci-lint`,
  `govulncheck`, `gitleaks`, `lefthook` and `task`. They run with `go tool -modfile=tools/go.mod <name>`.
  Nothing is downloaded at run time without a checksum.
- Developer programs that are not dependencies live in `dev/` (`gen-testdata`, `readme-gen`,
  `checkdocs`, `commitlint`) in the main module, so they use the same code the CLI does.

## Consequences

- Same tool versions locally and in CI; an update is a reviewed `go.mod`/`go.sum` diff (Dependabot, 7-day
  cooldown).
- The first `go tool` run compiles the tool, which takes minutes once; the build cache keeps it afterwards.
- Task itself is installed in CI with its official action pinned by SHA; locally the pinned copy
  (`go tool -modfile=tools/go.mod task`) needs no install.
- The shell commands inside a few targets (`test`, `cp`, `diff`, `rm`) need POSIX tools. On Windows, Git for
  Windows provides them.

## Alternatives rejected

- Keep Make: not portable, no `desc`, no variables from the platform.
- npm scripts and Husky (used by the web projects): this is a Go repository with no Node toolchain.
