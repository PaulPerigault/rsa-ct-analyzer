// Package batchgcd implements Bernstein's batch GCD algorithm (product tree,
// remainder tree) on top of math/big. It finds RSA moduli that share a prime
// factor with at least one other modulus of the set in quasi-linear time,
// instead of the O(n²) pairwise approach.
//
// Reference: D. J. Bernstein, "How to find smooth parts of integers" (2004);
// Heninger, Durumeric, Wustrow, Halderman, "Mining your Ps and Qs" (USENIX
// Security 2012).
package batchgcd

import (
	"context"
	"errors"
	"fmt"
	"math/big"
)

// ErrNonPositive is returned when an input is not a strictly positive integer.
var ErrNonPositive = errors.New("batchgcd: modulus must be positive")

// Compute returns, for every input n_i, gcd(n_i, ∏_{j≠i} n_j).
//
// A result of 1 means n_i shares no factor with the rest of the set. A result
// equal to n_i means every prime of n_i appears elsewhere in the set (typical
// of duplicated inputs: deduplicate before calling). Any other value is a
// non-trivial factor of n_i.
func Compute(ctx context.Context, moduli []*big.Int) ([]*big.Int, error) {
	for i, n := range moduli {
		if n == nil || n.Sign() <= 0 {
			return nil, fmt.Errorf("%w (index %d)", ErrNonPositive, i)
		}
	}
	switch len(moduli) {
	case 0:
		return nil, nil
	case 1:
		return []*big.Int{big.NewInt(1)}, nil
	}

	tree, err := productTree(ctx, moduli)
	if err != nil {
		return nil, err
	}
	return remainders(ctx, tree)
}

// productTree builds the levels bottom-up: level 0 holds the inputs, the last
// level holds the single product of everything. An odd node is carried up.
func productTree(ctx context.Context, leaves []*big.Int) ([][]*big.Int, error) {
	levels := [][]*big.Int{leaves}
	for cur := leaves; len(cur) > 1; {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		next := make([]*big.Int, 0, (len(cur)+1)/2)
		for i := 0; i < len(cur); i += 2 {
			if i+1 == len(cur) {
				next = append(next, cur[i])
				continue
			}
			next = append(next, new(big.Int).Mul(cur[i], cur[i+1]))
		}
		levels = append(levels, next)
		cur = next
	}
	return levels, nil
}

// remainders walks the tree top-down, computing P mod n² for every leaf, then
// derives gcd(n, (P mod n²)/n), which equals gcd(n, P/n).
func remainders(ctx context.Context, levels [][]*big.Int) ([]*big.Int, error) {
	top := len(levels) - 1
	rem := []*big.Int{new(big.Int).Set(levels[top][0])} // P mod P² = P

	sq := new(big.Int)
	for lvl := top - 1; lvl >= 0; lvl-- {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		nodes := levels[lvl]
		next := make([]*big.Int, len(nodes))
		for i, node := range nodes {
			sq.Mul(node, node)
			next[i] = new(big.Int).Mod(rem[i/2], sq)
		}
		rem = next
	}

	out := make([]*big.Int, len(rem))
	for i, z := range rem {
		n := levels[0][i]
		z.Quo(z, n)
		out[i] = z.GCD(nil, nil, z, n)
	}
	return out, nil
}
