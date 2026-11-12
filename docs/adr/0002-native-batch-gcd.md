# ADR 0002: batch GCD implemented in Go on `math/big`

Status: accepted (2026-11-12)

## Context

Finding RSA moduli that share a prime is O(n²) with pairwise GCD. Bernstein's product-tree and
remainder-tree algorithm is quasi-linear. The first version called an external C++ tool
(`batch_gcd` from the Cisco mercury project) through a CSV-to-hex conversion and a subprocess.

## Decision

Implement the algorithm in `pkg/batchgcd` on `math/big`: product tree up, remainder tree down
(`P mod n²`), then `gcd(n, (P mod n²)/n)`. The package has no dependency, takes a `context.Context`, and
is public so that others can reuse it. Identical moduli are merged before the call (they would return
`n`); a result equal to `n` is resolved by pairwise GCD against the other moduli, which only happens for
keys whose two primes are both shared.

## Consequences

- No native dependency, no subprocess, works on every platform the binary targets.
- Correctness is checked against the O(n²) definition on random sets of odd and even sizes, and on known
  values, edge cases and cancellation.
- `math/big` is slower than GMP. The benchmark issue (v0.3.0) will measure it and state the limits in the
  README. If the numbers are not acceptable, the answer is an optimisation of this package, not a
  subprocess.

## Alternatives rejected

- Wrap the C++ tool: build chain, platform support and a trust problem for a binary fetched from a branch.
- Use cgo with GMP bindings: breaks `CGO_ENABLED=0` and cross-compilation.
