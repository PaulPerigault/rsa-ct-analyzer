# Security policy

## Reporting a vulnerability

Do not open a public issue for a vulnerability. Use GitHub's
[private vulnerability reporting](https://github.com/PaulPerigault/rsa-ct-analyzer/security/advisories/new)
on this repository. Expect an answer within a few days.

## Scope and design choices

- The tool reads public key material. It never needs a private key; a private key
  passed as input is reported as a finding.
- Recovered prime factors are not printed unless `--reveal-factors` is set.
- `crawl` only sends `GET` requests to the log URL you give it, bounded in time,
  rate and response size. `mock-log` listens on loopback by default.
- Test keys under `testdata/` are weak on purpose, generated from a fixed seed by
  `dev/gen-testdata`. They protect nothing.

Only the latest release receives fixes.
