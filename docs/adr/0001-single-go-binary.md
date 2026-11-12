# ADR 0001: one Go binary, no cgo, no database

Status: accepted (2026-11-12)

## Context

The first version was three binaries (`crawler`, `extractor`, `analyzer`) chained through a SQLite
database (cgo), a YAML configuration with machine-specific paths, and an external C++ program for
batch GCD. Running it needed GCC, GMP, SQLite headers and a compiled third-party tool, or Docker.
The code that did the cryptography was a small part of the total.

## Decision

One static binary, `rsa-ct-analyzer`, with sub-commands: `scan` (local files and stdin), `crawl`
(CT log to a PEM bundle) and `mock-log` (simulated log for offline use). Pure Go, `CGO_ENABLED=0`,
standard library plus `cobra`. Data flows as PEM on stdout/stdin, not through a database. Configuration
is flags first, then `RSA_CT_*` environment variables documented in `.env.example`.

## Consequences

- Cross-compiles to five platforms from one machine; the container image is a single file on distroless.
- Pipelines compose with shell (`crawl | scan -`) and with CI (`--fail-on`, JSON output).
- No persistence: a crawl that is interrupted restarts. Resumable crawl is planned (issue in v0.3.0) with
  a checkpoint file, not a database.
- Memory grows with the number of moduli (batch GCD works on the whole set). Documented in the README, to
  be measured by the benchmark issue.

## Alternatives rejected

- Keep SQLite through a pure-Go driver: adds a dependency and a schema for data that is read once.
- Keep three binaries: no benefit once the database is gone.
