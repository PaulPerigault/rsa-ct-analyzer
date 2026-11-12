package analyze

import (
	"math/big"
	"sync"
)

// smallPrimeLimit bounds trial division: every prime below 2^16.
const smallPrimeLimit = 1 << 16

var (
	smallPrimesOnce sync.Once
	smallPrimes     []uint64
)

func primesBelowLimit() []uint64 {
	smallPrimesOnce.Do(func() {
		composite := make([]bool, smallPrimeLimit)
		for i := 2; i < smallPrimeLimit; i++ {
			if composite[i] {
				continue
			}
			smallPrimes = append(smallPrimes, uint64(i))
			for j := i * i; j < smallPrimeLimit; j += i {
				composite[j] = true
			}
		}
	})
	return smallPrimes
}

// smallFactor returns the smallest prime below 2^16 dividing n, or 0.
// n itself being that prime does not count: the modulus is then simply tiny
// and reported by the size check.
func smallFactor(n *big.Int) uint64 {
	var r, p big.Int
	for _, q := range primesBelowLimit() {
		p.SetUint64(q)
		if n.Cmp(&p) <= 0 {
			return 0
		}
		if r.Mod(n, &p).Sign() == 0 {
			return q
		}
	}
	return 0
}

// isPerfectSquare reports whether n = r² and returns r.
func isPerfectSquare(n *big.Int) (*big.Int, bool) {
	r := new(big.Int).Sqrt(n)
	return r, new(big.Int).Mul(r, r).Cmp(n) == 0
}

// fermat tries Fermat's factorisation: n = a² − b² = (a−b)(a+b). It succeeds
// quickly only when the two primes are close, which is the signature of a
// broken key generator. It gives up after rounds iterations.
func fermat(n *big.Int, rounds int) (p, q *big.Int, ok bool) {
	if rounds <= 0 || n.Bit(0) == 0 {
		return nil, nil, false
	}
	a, sq := isPerfectSquare(n)
	if !sq {
		a.Add(a, big.NewInt(1)) // ceil(sqrt(n))
	}
	var b2, b, chk big.Int
	one := big.NewInt(1)
	for range rounds {
		b2.Mul(a, a)
		b2.Sub(&b2, n)
		b.Sqrt(&b2)
		if chk.Mul(&b, &b).Cmp(&b2) == 0 {
			p = new(big.Int).Sub(a, &b)
			q = new(big.Int).Add(a, &b)
			if p.Cmp(one) > 0 {
				return p, q, true
			}
			return nil, nil, false
		}
		a.Add(a, one)
	}
	return nil, nil, false
}
