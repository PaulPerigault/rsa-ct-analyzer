// Package keys loads RSA public keys from X.509 certificates, PKIX/PKCS#1
// public keys and private keys (PEM or DER), local or from a stream.
package keys

import (
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"time"
)

// Kind tells what the key was extracted from.
type Kind string

// Supported input kinds.
const (
	KindCertificate Kind = "certificate"
	KindPublicKey   Kind = "public-key"
	KindPrivateKey  Kind = "private-key"
)

// CertInfo carries the certificate metadata the analysis needs.
type CertInfo struct {
	Subject            string
	Issuer             string
	Serial             string
	NotBefore          time.Time
	NotAfter           time.Time
	SignatureAlgorithm string
	DNSNames           []string
	SHA256             string // hex digest of the DER certificate
	Precert            bool
}

// Key is one RSA public key together with where it was found.
type Key struct {
	N      *big.Int
	E      int
	Source string // file path, "stdin" or "ct:<index>"
	Kind   Kind
	Cert   *CertInfo // nil unless Kind == KindCertificate
	Raw    []byte    // DER certificate when Kind == KindCertificate
}

// Bits returns the modulus size in bits.
func (k *Key) Bits() int { return k.N.BitLen() }

// Fingerprint returns the hex SHA-256 of the big-endian modulus. Two keys with
// the same fingerprint are the same RSA modulus whatever the exponent.
func (k *Key) Fingerprint() string {
	sum := sha256.Sum256(k.N.Bytes())
	return hex.EncodeToString(sum[:])
}

// Result is the outcome of loading one or several inputs.
type Result struct {
	Keys    []*Key
	Skipped int     // well-formed inputs that are not RSA (ECDSA, Ed25519…)
	Errors  []error // inputs that could not be parsed at all
}

func (r *Result) merge(o Result) {
	r.Keys = append(r.Keys, o.Keys...)
	r.Skipped += o.Skipped
	r.Errors = append(r.Errors, o.Errors...)
}
