package keys_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/PaulPerigault/rsa-ct-analyzer/internal/keys"
)

const testdata = "../../testdata"

func pemOf(t *testing.T, typ string, der []byte) []byte {
	t.Helper()
	return pem.EncodeToMemory(&pem.Block{Type: typ, Bytes: der})
}

func TestParseBytesFormats(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	pkix, _ := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	pkcs8, _ := x509.MarshalPKCS8PrivateKey(priv)

	cases := []struct {
		name string
		data []byte
		kind keys.Kind
	}{
		{"pkix public", pemOf(t, "PUBLIC KEY", pkix), keys.KindPublicKey},
		{"pkcs1 public", pemOf(t, "RSA PUBLIC KEY", x509.MarshalPKCS1PublicKey(&priv.PublicKey)), keys.KindPublicKey},
		{"pkcs1 private", pemOf(t, "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(priv)), keys.KindPrivateKey},
		{"pkcs8 private", pemOf(t, "PRIVATE KEY", pkcs8), keys.KindPrivateKey},
		{"raw DER pkix", pkix, keys.KindPublicKey},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := keys.ParseBytes(tc.data, "x")
			if len(res.Errors) != 0 || len(res.Keys) != 1 {
				t.Fatalf("got %d keys, errors %v", len(res.Keys), res.Errors)
			}
			k := res.Keys[0]
			if k.Kind != tc.kind || k.Bits() != 2048 || k.E != 65537 || k.N.Cmp(priv.N) != 0 {
				t.Errorf("unexpected key: kind %s bits %d e %d", k.Kind, k.Bits(), k.E)
			}
		})
	}
}

func TestParseCertificateMetadata(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(testdata, "certs", "weak-sha1-signature.pem"))
	if err != nil {
		t.Fatal(err)
	}
	res := keys.ParseBytes(data, "weak.pem")
	if len(res.Keys) != 1 {
		t.Fatalf("want 1 key, got %d (%v)", len(res.Keys), res.Errors)
	}
	c := res.Keys[0].Cert
	if c == nil || !strings.Contains(c.Subject, "weak-sha1-signature.example.test") ||
		!strings.HasPrefix(c.SignatureAlgorithm, "SHA1") || len(c.SHA256) != 64 || c.Precert {
		t.Errorf("unexpected certificate info: %+v", c)
	}
	if res.Keys[0].Raw == nil || res.Keys[0].Kind != keys.KindCertificate {
		t.Error("certificate kind and raw DER must be set")
	}
}

func TestParseSkipsNonRSA(t *testing.T) {
	ec, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	der, _ := x509.MarshalPKIXPublicKey(&ec.PublicKey)
	res := keys.ParseBytes(pemOf(t, "PUBLIC KEY", der), "ec.pem")
	if len(res.Keys) != 0 || res.Skipped != 1 || len(res.Errors) != 0 {
		t.Errorf("want 1 skipped, got keys=%d skipped=%d errors=%v", len(res.Keys), res.Skipped, res.Errors)
	}
}

func TestParseRejectsGarbage(t *testing.T) {
	for name, data := range map[string][]byte{
		"not a key":      []byte("hello"),
		"empty":          nil,
		"truncated PEM":  []byte("-----BEGIN CERTIFICATE-----\nAAAA\n"),
		"bad cert bytes": pemOf(t, "CERTIFICATE", []byte{1, 2, 3}),
	} {
		t.Run(name, func(t *testing.T) {
			res := keys.ParseBytes(data, "bad")
			if len(res.Keys) != 0 || len(res.Errors) == 0 {
				t.Errorf("want an error and no key, got keys=%d errors=%v", len(res.Keys), res.Errors)
			}
		})
	}
}

func TestLoadDirectoryAndStdin(t *testing.T) {
	res := keys.Load([]string{filepath.Join(testdata, "certs")}, nil)
	if len(res.Errors) != 0 {
		t.Fatalf("unexpected errors: %v", res.Errors)
	}
	// 13 RSA certificates (12 PEM + 1 DER), plus 1 ECDSA skipped.
	if len(res.Keys) != 13 || res.Skipped != 1 {
		t.Errorf("want 13 keys and 1 skipped, got %d and %d", len(res.Keys), res.Skipped)
	}

	data, _ := os.ReadFile(filepath.Join(testdata, "certs", "ok-2048.pem"))
	res = keys.Load([]string{keys.StdinName}, strings.NewReader(string(data)))
	if len(res.Keys) != 1 || res.Keys[0].Source != "stdin" {
		t.Errorf("stdin load failed: %+v", res)
	}
}

func TestLoadReportsMissingPathWithoutStopping(t *testing.T) {
	res := keys.Load([]string{"does-not-exist.pem", filepath.Join(testdata, "certs", "ok-2048.pem")}, nil)
	if len(res.Errors) != 1 || len(res.Keys) != 1 {
		t.Errorf("want 1 error and 1 key, got %d and %d", len(res.Errors), len(res.Keys))
	}
}

func TestFingerprintIgnoresExponent(t *testing.T) {
	a := keys.ParseBytes(mustRead(t, "certs/ok-2048.pem"), "a").Keys[0]
	b := keys.ParseBytes(mustRead(t, "certs/duplicate-of-ok-2048.pem"), "b").Keys[0]
	c := keys.ParseBytes(mustRead(t, "certs/ok-4096.pem"), "c").Keys[0]
	if a.Fingerprint() != b.Fingerprint() || a.Fingerprint() == c.Fingerprint() {
		t.Error("fingerprint must follow the modulus only")
	}
}

func mustRead(t *testing.T, rel string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(testdata, rel))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestBundleEntriesAreNumbered(t *testing.T) {
	bundle := append(mustRead(t, "certs/ok-2048.pem"), mustRead(t, "certs/ok-4096.pem")...)
	res := keys.ParseBytes(bundle, "bundle.pem")
	if len(res.Keys) != 2 || res.Keys[0].Source != "bundle.pem#1" || res.Keys[1].Source != "bundle.pem#2" {
		t.Errorf("sources: %+v", res.Keys)
	}
	single := keys.ParseBytes(mustRead(t, "certs/ok-2048.pem"), "one.pem")
	if single.Keys[0].Source != "one.pem" {
		t.Errorf("a single-block file keeps its plain name, got %q", single.Keys[0].Source)
	}
}
