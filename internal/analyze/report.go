// Package analyze audits RSA public keys: size, exponent, trivial factoring,
// duplicated keys and primes shared between keys (batch GCD).
package analyze

import (
	"fmt"
	"strings"
	"time"
)

// Severity ranks a finding.
type Severity int

// Severities, lowest first. Critical means the private key can be recovered.
const (
	Info Severity = iota
	Low
	Medium
	High
	Critical
)

var severityNames = [...]string{"info", "low", "medium", "high", "critical"}

func (s Severity) String() string {
	if s < Info || s > Critical {
		return fmt.Sprintf("severity(%d)", int(s))
	}
	return severityNames[s]
}

// MarshalText makes severities readable in JSON.
func (s Severity) MarshalText() ([]byte, error) { return []byte(s.String()), nil }

// ParseSeverity converts a name ("high") to a Severity.
func ParseSeverity(name string) (Severity, error) {
	for i, n := range severityNames {
		if strings.EqualFold(name, n) {
			return Severity(i), nil
		}
	}
	return Info, fmt.Errorf("unknown severity %q (want one of %s)", name, strings.Join(severityNames[:], ", "))
}

// Finding IDs, stable so they can be matched in CI.
const (
	IDKeySize        = "key-size"
	IDExponent       = "public-exponent"
	IDSmallFactor    = "small-factor"
	IDPerfectSquare  = "perfect-square"
	IDPrimeModulus   = "prime-modulus"
	IDCloseFactors   = "close-primes"
	IDSharedFactor   = "shared-prime"
	IDDuplicateKey   = "duplicate-key"
	IDWeakSignature  = "weak-signature"
	IDExpired        = "expired"
	IDPrivateKeyFile = "private-key-material"
)

// Finding is one problem found on one key.
type Finding struct {
	ID       string            `json:"id"`
	Severity Severity          `json:"severity"`
	Title    string            `json:"title"`
	Detail   string            `json:"detail,omitempty"`
	Evidence map[string]string `json:"evidence,omitempty"`
}

// KeyReport is everything known about one distinct RSA modulus.
type KeyReport struct {
	Fingerprint string    `json:"fingerprint_sha256"`
	Bits        int       `json:"bits"`
	Exponent    int       `json:"exponent"`
	Sources     []Source  `json:"sources"`
	Findings    []Finding `json:"findings"`
}

// Source is one place a key was seen.
type Source struct {
	Path        string     `json:"path"`
	Kind        string     `json:"kind"`
	Subject     string     `json:"subject,omitempty"`
	Issuer      string     `json:"issuer,omitempty"`
	NotAfter    *time.Time `json:"not_after,omitempty"`
	CertSHA256  string     `json:"cert_sha256,omitempty"`
	Precert     bool       `json:"precertificate,omitempty"`
	SigAlgoName string     `json:"signature_algorithm,omitempty"`
}

// MaxSeverity returns the highest severity among the findings (Info if none).
func (k *KeyReport) MaxSeverity() Severity {
	top := Info
	for _, f := range k.Findings {
		top = max(top, f.Severity)
	}
	return top
}

// Summary aggregates a run.
type Summary struct {
	KeysAnalyzed   int            `json:"keys_analyzed"`
	DistinctKeys   int            `json:"distinct_keys"`
	NonRSASkipped  int            `json:"non_rsa_skipped"`
	KeysWithIssues int            `json:"keys_with_findings"`
	BySeverity     map[string]int `json:"findings_by_severity"`
	BatchGCDRan    bool           `json:"batch_gcd"`
}

// Report is the full, serialisable result of an analysis.
type Report struct {
	Tool        string      `json:"tool"`
	Version     string      `json:"version"`
	GeneratedAt time.Time   `json:"generated_at"`
	Summary     Summary     `json:"summary"`
	Keys        []KeyReport `json:"keys"`
}

// MaxSeverity returns the highest severity found anywhere, and whether any
// finding exists at all.
func (r *Report) MaxSeverity() (Severity, bool) {
	top, found := Info, false
	for i := range r.Keys {
		for _, f := range r.Keys[i].Findings {
			top, found = max(top, f.Severity), true
		}
	}
	return top, found
}
