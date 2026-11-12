package analyze_test

import (
	"context"
	"math/big"
	"path/filepath"
	"testing"
	"time"

	"github.com/PaulPerigault/rsa-ct-analyzer/internal/analyze"
	"github.com/PaulPerigault/rsa-ct-analyzer/internal/keys"
)

const certs = "../../testdata/certs"

var fixedNow = time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)

func options() analyze.Options {
	o := analyze.Defaults()
	o.Now = func() time.Time { return fixedNow }
	return o
}

func run(t *testing.T, opt analyze.Options, files ...string) *analyze.Report {
	t.Helper()
	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = filepath.Join(certs, f)
	}
	loaded := keys.Load(paths, nil)
	if len(loaded.Errors) != 0 {
		t.Fatal(loaded.Errors)
	}
	rep, err := analyze.Analyze(context.Background(), loaded.Keys, loaded.Skipped, opt)
	if err != nil {
		t.Fatal(err)
	}
	return rep
}

func ids(rep *analyze.Report) map[string]analyze.Severity {
	out := map[string]analyze.Severity{}
	for _, k := range rep.Keys {
		for _, f := range k.Findings {
			out[f.ID] = max(out[f.ID], f.Severity)
		}
	}
	return out
}

// Each sample file must trigger exactly the finding it was built for.
func TestSampleCertificates(t *testing.T) {
	cases := []struct {
		file string
		want map[string]analyze.Severity
	}{
		{"ok-2048.pem", nil},
		{"ok-4096.pem", nil},
		{"weak-512.pem", map[string]analyze.Severity{analyze.IDKeySize: analyze.Critical}},
		{"weak-1024.pem", map[string]analyze.Severity{analyze.IDKeySize: analyze.High}},
		{"weak-exponent-3.pem", map[string]analyze.Severity{analyze.IDExponent: analyze.Medium}},
		{"small-factor.pem", map[string]analyze.Severity{analyze.IDSmallFactor: analyze.Critical}},
		{"close-primes.pem", map[string]analyze.Severity{analyze.IDCloseFactors: analyze.Critical}},
		{"weak-sha1-signature.pem", map[string]analyze.Severity{analyze.IDWeakSignature: analyze.Medium}},
		{"expired.pem", map[string]analyze.Severity{analyze.IDExpired: analyze.Low}},
	}
	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			got := ids(run(t, options(), tc.file))
			if len(got) != len(tc.want) {
				t.Fatalf("findings %v, want %v", got, tc.want)
			}
			for id, sev := range tc.want {
				if got[id] != sev {
					t.Errorf("%s: severity %s, want %s", id, got[id], sev)
				}
			}
		})
	}
}

func TestSharedPrimeNeedsBothKeys(t *testing.T) {
	alone := run(t, options(), "shared-prime-a.pem", "ok-2048.pem")
	if _, hit := ids(alone)[analyze.IDSharedFactor]; hit {
		t.Fatal("a key must not be flagged without its partner in the set")
	}

	both := run(t, options(), "shared-prime-a.pem", "shared-prime-b.pem", "ok-2048.pem")
	flagged := 0
	for _, k := range both.Keys {
		for _, f := range k.Findings {
			if f.ID == analyze.IDSharedFactor {
				flagged++
				if f.Severity != analyze.Critical || f.Evidence["factor_bits_small"] != "1024" {
					t.Errorf("unexpected finding: %+v", f)
				}
				if _, leaked := f.Evidence["p"]; leaked {
					t.Error("primes must stay hidden unless RevealFactors is set")
				}
			}
		}
	}
	if flagged != 2 {
		t.Errorf("want both keys flagged, got %d", flagged)
	}
}

func TestSharedPrimeDisabled(t *testing.T) {
	opt := options()
	opt.BatchGCD = false
	rep := run(t, opt, "shared-prime-a.pem", "shared-prime-b.pem")
	if _, hit := ids(rep)[analyze.IDSharedFactor]; hit || rep.Summary.BatchGCDRan {
		t.Error("batch GCD must not run when disabled")
	}
}

func TestRevealFactorsDivides(t *testing.T) {
	opt := options()
	opt.RevealFactors = true
	rep := run(t, opt, "small-factor.pem")
	f := rep.Keys[0].Findings[0]
	p, q := new(big.Int), new(big.Int)
	p.SetString(f.Evidence["p"], 16)
	q.SetString(f.Evidence["q"], 16)
	loaded := keys.Load([]string{filepath.Join(certs, "small-factor.pem")}, nil)
	if new(big.Int).Mul(p, q).Cmp(loaded.Keys[0].N) != 0 {
		t.Error("revealed p·q must equal the modulus")
	}
}

func TestDuplicateKeysAreMerged(t *testing.T) {
	rep := run(t, options(), "ok-2048.pem", "duplicate-of-ok-2048.pem", "ok-2048.der")
	if rep.Summary.KeysAnalyzed != 3 || rep.Summary.DistinctKeys != 1 || len(rep.Keys[0].Sources) != 3 {
		t.Errorf("summary: %+v", rep.Summary)
	}
	if ids(rep)[analyze.IDDuplicateKey] != analyze.High {
		t.Error("duplicate key must be reported as high")
	}
	// Duplicates must not be mistaken for a shared prime.
	if _, hit := ids(rep)[analyze.IDSharedFactor]; hit {
		t.Error("duplicates must be deduplicated before batch GCD")
	}
}

func TestMinBitsOption(t *testing.T) {
	opt := options()
	opt.MinBits = 4096
	if ids(run(t, opt, "ok-2048.pem"))[analyze.IDKeySize] != analyze.High {
		t.Error("2048-bit key must be High under --min-bits 4096")
	}
}

func TestFermatCanBeDisabled(t *testing.T) {
	opt := options()
	opt.FermatRounds = 0
	if _, hit := ids(run(t, opt, "close-primes.pem"))[analyze.IDCloseFactors]; hit {
		t.Error("Fermat must not run with 0 rounds")
	}
}

func TestStructuralWeaknesses(t *testing.T) {
	prime := mustPrime(t, 1021)
	cases := []struct {
		name string
		n    *big.Int
		e    int
		want string
	}{
		{"perfect square", new(big.Int).Mul(prime, prime), 65537, analyze.IDPerfectSquare},
		{"prime modulus", mustPrime(t, 1024), 65537, analyze.IDPrimeModulus},
		{"even modulus", new(big.Int).Lsh(mustPrime(t, 1022), 1), 65537, analyze.IDSmallFactor},
		{"exponent 1", new(big.Int).Mul(mustPrime(t, 1024), mustPrime(t, 1027)), 1, analyze.IDExponent},
		{"even exponent", new(big.Int).Mul(mustPrime(t, 1024), mustPrime(t, 1027)), 65536, analyze.IDExponent},
		{"exponent 17", new(big.Int).Mul(mustPrime(t, 1024), mustPrime(t, 1027)), 17, analyze.IDExponent},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			k := &keys.Key{N: tc.n, E: tc.e, Source: "mem", Kind: keys.KindPublicKey}
			rep, err := analyze.Analyze(context.Background(), []*keys.Key{k}, 0, options())
			if err != nil {
				t.Fatal(err)
			}
			if _, hit := ids(rep)[tc.want]; !hit {
				t.Errorf("want finding %q, got %v", tc.want, ids(rep))
			}
		})
	}
}

func TestPrivateKeyInputIsFlagged(t *testing.T) {
	n := new(big.Int).Mul(mustPrime(t, 1024), mustPrime(t, 1027))
	k := &keys.Key{N: n, E: 65537, Source: "id_rsa", Kind: keys.KindPrivateKey}
	rep, err := analyze.Analyze(context.Background(), []*keys.Key{k}, 0, options())
	if err != nil {
		t.Fatal(err)
	}
	if ids(rep)[analyze.IDPrivateKeyFile] != analyze.Medium {
		t.Errorf("findings: %v", ids(rep))
	}
}

func TestReportIsDeterministicAndSorted(t *testing.T) {
	files := []string{"ok-2048.pem", "weak-512.pem", "expired.pem", "shared-prime-a.pem", "shared-prime-b.pem"}
	a, b := run(t, options(), files...), run(t, options(), files...)
	for i := range a.Keys {
		if a.Keys[i].Fingerprint != b.Keys[i].Fingerprint {
			t.Fatal("two runs on the same input must give the same order")
		}
		if i > 0 && a.Keys[i].MaxSeverity() > a.Keys[i-1].MaxSeverity() {
			t.Error("keys must be sorted by decreasing severity")
		}
	}
	top, found := a.MaxSeverity()
	if !found || top != analyze.Critical {
		t.Errorf("max severity %s (found=%v), want critical", top, found)
	}
}

func TestCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	loaded := keys.Load([]string{certs}, nil)
	if _, err := analyze.Analyze(ctx, loaded.Keys, 0, options()); err == nil {
		t.Error("a cancelled context must abort the analysis")
	}
}

func TestSeverityParsing(t *testing.T) {
	for _, name := range []string{"info", "LOW", "Medium", "high", "critical"} {
		if _, err := analyze.ParseSeverity(name); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
	if _, err := analyze.ParseSeverity("urgent"); err == nil {
		t.Error("unknown severity must be rejected")
	}
	if analyze.Critical.String() != "critical" || analyze.Severity(42).String() == "" {
		t.Error("String() misbehaves")
	}
}

func mustPrime(t *testing.T, bits int) *big.Int {
	t.Helper()
	// Deterministic enough for tests: next prime above 2^(bits-1) + a fixed offset.
	p := new(big.Int).Lsh(big.NewInt(1), uint(bits-1))
	p.Add(p, big.NewInt(int64(bits)*7919))
	p.SetBit(p, 0, 1)
	for !p.ProbablyPrime(10) {
		p.Add(p, big.NewInt(2))
	}
	return p
}
