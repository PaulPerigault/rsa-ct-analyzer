// Command gen-testdata writes the sample certificates and keys under testdata/.
//
// Everything is derived from a fixed seed (a deterministic PRNG, not
// crypto/rand), so the output is byte-for-byte reproducible and reviewable.
// The keys are for tests only: they are weak on purpose and must never be used
// to protect anything.
//
//	go run ./dev/gen-testdata -out testdata
package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"math/big"
	mrand "math/rand/v2"
	"os"
	"path/filepath"
	"time"
)

const seed = 20261112

var (
	rng   = mrand.New(mrand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
	two   = big.NewInt(2)
	valid = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
)

func main() {
	out := flag.String("out", "testdata", "output directory")
	flag.Parse()
	if err := run(*out); err != nil {
		log.Fatal(err)
	}
}

type gen struct {
	dir    string
	ca     *x509.Certificate
	caKey  *rsa.PrivateKey
	serial int64
}

func run(out string) error {
	certs := filepath.Join(out, "certs")
	keysDir := filepath.Join(out, "keys")
	for _, d := range []string{certs, keysDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}

	caKey := mustKey(prime(1024, 65537), prime(1024, 65537), 65537)
	g := &gen{dir: certs, caKey: caKey}
	g.ca = g.selfSignedCA()

	ok2048 := pubOf(prime(1024, 65537), prime(1024, 65537), 65537)
	ok4096 := pubOf(prime(2048, 65537), prime(2048, 65537), 65537)
	shared := prime(1024, 65537) // the prime reused by two moduli
	sharedA := pubOf(shared, prime(1024, 65537), 65537)
	sharedB := pubOf(shared, prime(1024, 65537), 65537)
	close1 := prime(1024, 65537)
	closeKey := pubOf(close1, nextPrime(close1, 65537), 65537)

	items := []struct {
		name string
		pub  *rsa.PublicKey
		opts certOpts
	}{
		{"ok-2048", ok2048, certOpts{}},
		{"ok-4096", ok4096, certOpts{}},
		{"duplicate-of-ok-2048", ok2048, certOpts{}},
		{"weak-512", pubOf(prime(256, 65537), prime(256, 65537), 65537), certOpts{}},
		{"weak-1024", pubOf(prime(512, 65537), prime(512, 65537), 65537), certOpts{}},
		{"weak-exponent-3", pubOf(prime(1024, 3), prime(1024, 3), 3), certOpts{}},
		{"small-factor", pubOf(big.NewInt(101), prime(2041, 65537), 65537), certOpts{}},
		{"close-primes", closeKey, certOpts{}},
		{"shared-prime-a", sharedA, certOpts{}},
		{"shared-prime-b", sharedB, certOpts{}},
		{"weak-sha1-signature", pubOf(prime(1024, 65537), prime(1024, 65537), 65537), certOpts{sig: x509.SHA1WithRSA}},
		{"expired", pubOf(prime(1024, 65537), prime(1024, 65537), 65537), certOpts{expired: true}},
	}
	for _, it := range items {
		der, err := g.leaf(it.name, it.pub, it.opts)
		if err != nil {
			return fmt.Errorf("%s: %w", it.name, err)
		}
		if err := writePEM(filepath.Join(certs, it.name+".pem"), "CERTIFICATE", der); err != nil {
			return err
		}
		if it.name == "ok-2048" { // same certificate in DER, to cover binary input
			if err := os.WriteFile(filepath.Join(certs, "ok-2048.der"), der, 0o644); err != nil { //nolint:gosec
				return err
			}
		}
	}

	if err := g.ecdsaLeaf(); err != nil {
		return err
	}
	return writePublicKeys(keysDir, ok2048, ok4096)
}

// prime returns a deterministic prime of the given size such that gcd(e, p-1) = 1.
func prime(bits, e int) *big.Int {
	ee := big.NewInt(int64(e))
	for {
		buf := make([]byte, (bits+7)/8)
		for i := range buf {
			buf[i] = byte(rng.Uint32())
		}
		p := new(big.Int).SetBytes(buf)
		p.Rsh(p, uint(len(buf)*8-bits))
		p.SetBit(p, bits-1, 1)
		p.SetBit(p, bits-2, 1)
		p.SetBit(p, 0, 1)
		if p.ProbablyPrime(20) && coprime(ee, p) {
			return p
		}
	}
}

func nextPrime(p *big.Int, e int) *big.Int {
	q := new(big.Int).Add(p, two)
	for !(q.ProbablyPrime(20) && coprime(big.NewInt(int64(e)), q)) {
		q.Add(q, two)
	}
	return q
}

func coprime(e, p *big.Int) bool {
	pm1 := new(big.Int).Sub(p, big.NewInt(1))
	return new(big.Int).GCD(nil, nil, e, pm1).Cmp(big.NewInt(1)) == 0
}

func pubOf(p, q *big.Int, e int) *rsa.PublicKey {
	return &rsa.PublicKey{N: new(big.Int).Mul(p, q), E: e}
}

// mustKey builds a full private key from two primes. Only the CA signs; leaf
// keys are public keys, some of them deliberately impossible to validate.
func mustKey(p, q *big.Int, e int) *rsa.PrivateKey {
	n := new(big.Int).Mul(p, q)
	phi := new(big.Int).Mul(new(big.Int).Sub(p, big.NewInt(1)), new(big.Int).Sub(q, big.NewInt(1)))
	d := new(big.Int).ModInverse(big.NewInt(int64(e)), phi)
	k := &rsa.PrivateKey{PublicKey: rsa.PublicKey{N: n, E: e}, D: d, Primes: []*big.Int{p, q}}
	k.Precompute()
	if err := k.Validate(); err != nil {
		log.Fatalf("generated key invalid: %v", err)
	}
	return k
}

type certOpts struct {
	sig     x509.SignatureAlgorithm
	expired bool
}

func (g *gen) template(cn string) *x509.Certificate {
	g.serial++
	return &x509.Certificate{
		SerialNumber: big.NewInt(g.serial),
		Subject:      pkix.Name{CommonName: cn, Organization: []string{"rsa-ct-analyzer test data"}},
		NotBefore:    valid,
		NotAfter:     valid.AddDate(10, 0, 0),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{cn},
	}
}

func (g *gen) selfSignedCA() *x509.Certificate {
	t := g.template("Test CA (not trusted, tests only)")
	t.DNSNames, t.ExtKeyUsage = nil, nil
	t.IsCA, t.BasicConstraintsValid = true, true
	t.KeyUsage = x509.KeyUsageCertSign
	der, err := x509.CreateCertificate(rand.Reader, t, t, &g.caKey.PublicKey, g.caKey)
	if err != nil {
		log.Fatal(err)
	}
	c, err := x509.ParseCertificate(der)
	if err != nil {
		log.Fatal(err)
	}
	return c
}

func (g *gen) leaf(name string, pub *rsa.PublicKey, o certOpts) ([]byte, error) {
	t := g.template(name + ".example.test")
	t.SignatureAlgorithm = o.sig
	if o.expired {
		t.NotBefore = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		t.NotAfter = time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	return x509.CreateCertificate(rand.Reader, t, g.ca, pub, g.caKey)
}

// ecdsaLeaf adds a non-RSA certificate: the tool must skip it, not fail.
func (g *gen) ecdsaLeaf() error {
	raw := make([]byte, 32)
	for i := range raw {
		raw[i] = byte(rng.Uint32())
	}
	raw[0] &= 0x7f // keep the scalar below the curve order
	priv, err := ecdsa.ParseRawPrivateKey(elliptic.P256(), raw)
	if err != nil {
		return err
	}
	t := g.template("ecdsa.example.test")
	der, err := x509.CreateCertificate(rand.Reader, t, g.ca, &priv.PublicKey, g.caKey)
	if err != nil {
		return err
	}
	return writePEM(filepath.Join(g.dir, "ecdsa-p256.pem"), "CERTIFICATE", der)
}

func writePublicKeys(dir string, ok2048, ok4096 *rsa.PublicKey) error {
	pkix, err := x509.MarshalPKIXPublicKey(ok2048)
	if err != nil {
		return err
	}
	if err := writePEM(filepath.Join(dir, "public-pkix-2048.pem"), "PUBLIC KEY", pkix); err != nil {
		return err
	}
	return writePEM(filepath.Join(dir, "public-pkcs1-4096.pem"), "RSA PUBLIC KEY", x509.MarshalPKCS1PublicKey(ok4096))
}

func writePEM(path, typ string, der []byte) error {
	return os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: typ, Bytes: der}), 0o644) //nolint:gosec // public test data
}
