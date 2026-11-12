package analyze

import (
	"fmt"
	"math/big"
	"slices"
	"strings"
	"time"

	"github.com/PaulPerigault/rsa-ct-analyzer/internal/keys"
)

// Thresholds. Sources: NIST SP 800-131A r2 / SP 800-57 (RSA < 2048 bits is
// disallowed for new keys), CA/B Forum Baseline Requirements §6.1.5.
const (
	brokenBits    = 1024
	defaultMinBit = 2048
	stdExponent   = 65537
)

// modulusChecks runs the checks that depend on the modulus only (and the
// exponents it is used with).
func modulusChecks(n *big.Int, exps []int, opt Options) []Finding {
	var out []Finding
	out = append(out, sizeFinding(n.BitLen(), opt.MinBits)...)
	for _, e := range exps {
		out = append(out, exponentFinding(e)...)
	}
	out = append(out, factoringFindings(n, opt)...)
	return out
}

func sizeFinding(bits, minBits int) []Finding {
	switch {
	case bits < brokenBits:
		return []Finding{{
			ID: IDKeySize, Severity: Critical,
			Title:    fmt.Sprintf("%d-bit RSA modulus can be factored with commodity hardware", bits),
			Evidence: map[string]string{"bits": fmt.Sprint(bits)},
		}}
	case bits < minBits:
		return []Finding{{
			ID: IDKeySize, Severity: High,
			Title:    fmt.Sprintf("%d-bit RSA modulus is below the %d-bit minimum", bits, minBits),
			Detail:   "NIST SP 800-131A disallows RSA below 2048 bits for new keys.",
			Evidence: map[string]string{"bits": fmt.Sprint(bits), "minimum": fmt.Sprint(minBits)},
		}}
	}
	return nil
}

func exponentFinding(e int) []Finding {
	ev := map[string]string{"exponent": fmt.Sprint(e)}
	switch {
	case e < 3 || e%2 == 0:
		return []Finding{{
			ID: IDExponent, Severity: Critical, Evidence: ev,
			Title:  fmt.Sprintf("public exponent %d is invalid", e),
			Detail: "RSA needs an odd exponent ≥ 3 coprime with φ(n); e=1 does not encrypt at all.",
		}}
	case e == 3:
		return []Finding{{
			ID: IDExponent, Severity: Medium, Evidence: ev,
			Title:  "public exponent 3 is fragile",
			Detail: "Enables cube-root attacks on unpadded messages and Bleichenbacher e=3 signature forgery against lax verifiers.",
		}}
	case e < stdExponent:
		return []Finding{{
			ID: IDExponent, Severity: Low, Evidence: ev,
			Title:  fmt.Sprintf("public exponent %d is below the standard 65537", e),
			Detail: "Exponents below 2^16+1 are discouraged by NIST SP 800-56B (§6.2).",
		}}
	}
	return nil
}

// factoringFindings tries the cheap factoring methods and stops at the first
// hit: once one factorisation is known the key is already fully broken.
func factoringFindings(n *big.Int, opt Options) []Finding {
	if n.Sign() <= 0 {
		return nil
	}
	if q := smallFactor(n); q != 0 {
		p := new(big.Int).SetUint64(q)
		return []Finding{factoredFinding(IDSmallFactor,
			fmt.Sprintf("modulus is divisible by the small prime %d", q), factoring{n, p}, opt)}
	}
	if r, ok := isPerfectSquare(n); ok {
		return []Finding{factoredFinding(IDPerfectSquare, "modulus is a perfect square (n = p²)", factoring{n, r}, opt)}
	}
	if n.BitLen() > 2 && n.ProbablyPrime(0) {
		return []Finding{{
			ID: IDPrimeModulus, Severity: Critical,
			Title:  "modulus is prime, not a product of two primes",
			Detail: "φ(n) = n−1 is public, so the private exponent is computable.",
		}}
	}
	if p, _, ok := fermat(n, opt.FermatRounds); ok {
		return []Finding{factoredFinding(IDCloseFactors,
			"the two primes are close: factored with Fermat's method", factoring{n, p}, opt)}
	}
	return nil
}

// factoring is a modulus together with one of its proper factors.
type factoring struct{ n, p *big.Int }

func factoredFinding(id, title string, f factoring, opt Options) Finding {
	return Finding{
		ID: id, Severity: Critical, Title: title,
		Detail:   "The private key can be recovered from the public key alone.",
		Evidence: factorEvidence(f.n, f.p, opt.RevealFactors),
	}
}

// factorEvidence documents a factorisation without publishing the factors
// unless the operator asked for it.
func factorEvidence(n, p *big.Int, reveal bool) map[string]string {
	q := new(big.Int).Quo(n, p)
	small, large := p, q
	if small.Cmp(large) > 0 {
		small, large = large, small
	}
	ev := map[string]string{
		"factor_bits_small": fmt.Sprint(small.BitLen()),
		"factor_bits_large": fmt.Sprint(large.BitLen()),
	}
	if reveal {
		ev["p"] = small.Text(16)
		ev["q"] = large.Text(16)
	}
	return ev
}

var weakSignatures = []string{"MD2", "MD4", "MD5", "SHA1"}

// certificateChecks runs the checks specific to a certificate source.
func certificateChecks(c *keys.CertInfo, now time.Time) []Finding {
	var out []Finding
	for _, w := range weakSignatures {
		if strings.HasPrefix(c.SignatureAlgorithm, w) {
			out = append(out, Finding{
				ID: IDWeakSignature, Severity: Medium,
				Title:    "certificate signed with " + c.SignatureAlgorithm,
				Detail:   "Collision attacks make this hash unsuitable for signatures (RFC 9155 deprecates SHA-1 in TLS).",
				Evidence: map[string]string{"signature_algorithm": c.SignatureAlgorithm},
			})
			break
		}
	}
	if now.After(c.NotAfter) {
		out = append(out, Finding{
			ID: IDExpired, Severity: Low,
			Title:    "certificate expired on " + c.NotAfter.UTC().Format(time.DateOnly),
			Evidence: map[string]string{"not_after": c.NotAfter.UTC().Format(time.RFC3339)},
		})
	}
	return out
}

func privateKeyFinding() Finding {
	return Finding{
		ID: IDPrivateKeyFile, Severity: Medium,
		Title:  "a private key was supplied as input",
		Detail: "Private keys do not belong in audited material or in a repository; rotate it if it was ever shared.",
	}
}

func distinctExponents(ks []*keys.Key) []int {
	var out []int
	for _, k := range ks {
		if !slices.Contains(out, k.E) {
			out = append(out, k.E)
		}
	}
	slices.Sort(out)
	return out
}
