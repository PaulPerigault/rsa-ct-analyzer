package batchgcd_test

import (
	"context"
	"errors"
	"math/big"
	"math/rand/v2"
	"testing"

	"github.com/PaulPerigault/rsa-ct-analyzer/pkg/batchgcd"
)

func bi(v int64) *big.Int { return big.NewInt(v) }

func TestComputeKnownValues(t *testing.T) {
	// 15 = 3·5, 21 = 3·7, 35 = 5·7, 11·13 = 143 isolated.
	in := []*big.Int{bi(15), bi(21), bi(35), bi(143)}
	got, err := batchgcd.Compute(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	// gcd(15, 21·35·143) = 15 (3 and 5 both shared): result equals n.
	want := []int64{15, 21, 35, 1}
	for i, w := range want {
		if got[i].Cmp(bi(w)) != 0 {
			t.Errorf("index %d: got %s, want %d", i, got[i], w)
		}
	}
}

func TestComputeSingleSharedPrime(t *testing.T) {
	p := bi(1000003)
	moduli := []*big.Int{
		new(big.Int).Mul(p, bi(1000033)),
		new(big.Int).Mul(p, bi(1000037)),
		new(big.Int).Mul(bi(1000039), bi(1000081)),
	}
	got, err := batchgcd.Compute(context.Background(), moduli)
	if err != nil {
		t.Fatal(err)
	}
	if got[0].Cmp(p) != 0 || got[1].Cmp(p) != 0 {
		t.Errorf("shared prime not recovered: %s, %s", got[0], got[1])
	}
	if got[2].Cmp(bi(1)) != 0 {
		t.Errorf("isolated modulus flagged: %s", got[2])
	}
}

func TestComputeEdgeCases(t *testing.T) {
	ctx := context.Background()
	if out, err := batchgcd.Compute(ctx, nil); err != nil || out != nil {
		t.Errorf("empty input: got %v, %v", out, err)
	}
	out, err := batchgcd.Compute(ctx, []*big.Int{bi(77)})
	if err != nil || len(out) != 1 || out[0].Cmp(bi(1)) != 0 {
		t.Errorf("single input: got %v, %v", out, err)
	}
	for _, bad := range []*big.Int{nil, bi(0), bi(-5)} {
		if _, err := batchgcd.Compute(ctx, []*big.Int{bi(15), bad}); !errors.Is(err, batchgcd.ErrNonPositive) {
			t.Errorf("input %v: want ErrNonPositive, got %v", bad, err)
		}
	}
}

func TestComputeHonoursCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := batchgcd.Compute(ctx, []*big.Int{bi(15), bi(21), bi(35)}); !errors.Is(err, context.Canceled) {
		t.Fatalf("want context.Canceled, got %v", err)
	}
}

// TestComputeMatchesPairwise cross-checks the tree algorithm against the naive
// O(n²) definition on random sets with an odd and an even size.
func TestComputeMatchesPairwise(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	for _, size := range []int{2, 5, 8, 13} {
		primes := make([]*big.Int, 0, size*2)
		for len(primes) < size+3 { // fewer primes than needed: forces sharing
			p := new(big.Int).SetUint64(rng.Uint64()>>34 | 1<<30)
			if p.ProbablyPrime(10) {
				primes = append(primes, p)
			}
		}
		moduli := make([]*big.Int, size)
		for i := range moduli {
			a, b := primes[rng.IntN(len(primes))], primes[rng.IntN(len(primes))]
			for a.Cmp(b) == 0 {
				b = primes[rng.IntN(len(primes))]
			}
			moduli[i] = new(big.Int).Mul(a, b)
		}
		got, err := batchgcd.Compute(context.Background(), moduli)
		if err != nil {
			t.Fatal(err)
		}
		for i, n := range moduli {
			rest := big.NewInt(1)
			for j, m := range moduli {
				if j != i {
					rest.Mul(rest, m)
				}
			}
			want := new(big.Int).GCD(nil, nil, n, rest)
			if got[i].Cmp(want) != 0 {
				t.Fatalf("size %d index %d: got %s, want %s", size, i, got[i], want)
			}
		}
	}
}
