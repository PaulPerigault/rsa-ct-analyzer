package analyze

import (
	"context"
	"fmt"
	"math/big"
	"runtime"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/PaulPerigault/rsa-ct-analyzer/internal/keys"
	"github.com/PaulPerigault/rsa-ct-analyzer/pkg/batchgcd"
)

// Options tune an analysis. The zero value is not usable: start from Defaults.
type Options struct {
	MinBits       int  // keys below are High (below 1024: Critical)
	FermatRounds  int  // 0 disables Fermat's close-primes check
	BatchGCD      bool // look for primes shared between keys
	RevealFactors bool // print recovered primes in the evidence
	Workers       int
	Now           func() time.Time
	Tool, Version string
}

// Defaults returns the recommended options.
func Defaults() Options {
	return Options{
		MinBits:      defaultMinBit,
		FermatRounds: 10_000,
		BatchGCD:     true,
		Workers:      runtime.NumCPU(),
		Now:          time.Now,
	}
}

// group is a distinct modulus and every place it was seen.
type group struct {
	n        *big.Int
	fp       string
	keys     []*keys.Key
	findings []Finding
}

// Analyze audits ks and returns a deterministic report. nonRSA is the number
// of inputs skipped by the loader, only used for the summary.
func Analyze(ctx context.Context, ks []*keys.Key, nonRSA int, opt Options) (*Report, error) {
	opt = normalise(opt)
	groups := groupByModulus(ks)

	if err := runModulusChecks(ctx, groups, opt); err != nil {
		return nil, err
	}
	if opt.BatchGCD {
		if err := sharedPrimeChecks(ctx, groups, opt); err != nil {
			return nil, err
		}
	}
	return buildReport(groups, len(ks), nonRSA, opt), nil
}

func normalise(opt Options) Options {
	if opt.Workers < 1 {
		opt.Workers = 1
	}
	if opt.Now == nil {
		opt.Now = time.Now
	}
	if opt.MinBits == 0 {
		opt.MinBits = defaultMinBit
	}
	return opt
}

func groupByModulus(ks []*keys.Key) []*group {
	byFP := map[string]*group{}
	var order []*group
	for _, k := range ks {
		fp := k.Fingerprint()
		g, ok := byFP[fp]
		if !ok {
			g = &group{n: k.N, fp: fp}
			byFP[fp] = g
			order = append(order, g)
		}
		g.keys = append(g.keys, k)
	}
	return order
}

// runModulusChecks fans the per-modulus checks out on a worker pool.
func runModulusChecks(ctx context.Context, groups []*group, opt Options) error {
	jobs := make(chan *group)
	var wg sync.WaitGroup
	for range opt.Workers {
		wg.Go(func() {
			for g := range jobs {
				g.findings = append(g.findings, modulusChecks(g.n, distinctExponents(g.keys), opt)...)
			}
		})
	}
	defer wg.Wait()
	defer close(jobs)
	for _, g := range groups {
		select {
		case jobs <- g:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

// sharedPrimeChecks runs batch GCD on the distinct moduli. Duplicated moduli
// were merged beforehand: left in, they would all come out as gcd = n.
func sharedPrimeChecks(ctx context.Context, groups []*group, opt Options) error {
	if len(groups) < 2 {
		return nil
	}
	moduli := make([]*big.Int, len(groups))
	for i, g := range groups {
		moduli[i] = g.n
	}
	gcds, err := batchgcd.Compute(ctx, moduli)
	if err != nil {
		return fmt.Errorf("batch gcd: %w", err)
	}
	one := big.NewInt(1)
	for i, g := range gcds {
		if g.Cmp(one) == 0 {
			continue
		}
		factor := g
		if g.Cmp(moduli[i]) == 0 { // both primes shared: recover one pairwise
			factor = pairwiseFactor(i, moduli)
		}
		groups[i].findings = append(groups[i].findings, sharedFinding(i, factor, groups, opt))
	}
	return nil
}

// pairwiseFactor finds a proper factor of moduli[i] by gcd with each other
// modulus. It only runs for the rare keys whose two primes are both reused.
func pairwiseFactor(i int, moduli []*big.Int) *big.Int {
	one := big.NewInt(1)
	for j, m := range moduli {
		if j == i {
			continue
		}
		d := new(big.Int).GCD(nil, nil, moduli[i], m)
		if d.Cmp(one) > 0 && d.Cmp(moduli[i]) < 0 {
			return d
		}
	}
	return moduli[i]
}

func sharedFinding(i int, factor *big.Int, groups []*group, opt Options) Finding {
	var others []string
	var rem big.Int
	for j, o := range groups {
		if j != i && rem.Mod(o.n, factor).Sign() == 0 {
			others = append(others, o.fp[:12])
		}
	}
	ev := factorEvidence(groups[i].n, factor, opt.RevealFactors)
	ev["shared_with"] = strings.Join(others, ",")
	return Finding{
		ID: IDSharedFactor, Severity: Critical,
		Title: fmt.Sprintf("shares a %d-bit prime with %d other key(s)", factor.BitLen(), len(others)),
		Detail: "Found by batch GCD: gcd of the moduli reveals the prime, so every key involved " +
			"is factored. Typical of low-entropy key generation.",
		Evidence: ev,
	}
}

func buildReport(groups []*group, total, nonRSA int, opt Options) *Report {
	now := opt.Now()
	rep := &Report{
		Tool: opt.Tool, Version: opt.Version, GeneratedAt: now.UTC(),
		Summary: Summary{
			KeysAnalyzed: total, DistinctKeys: len(groups), NonRSASkipped: nonRSA,
			BySeverity: map[string]int{}, BatchGCDRan: opt.BatchGCD && len(groups) > 1,
		},
	}
	for _, g := range groups {
		kr := keyReport(g, now)
		for _, f := range kr.Findings {
			rep.Summary.BySeverity[f.Severity.String()]++
		}
		if len(kr.Findings) > 0 {
			rep.Summary.KeysWithIssues++
		}
		rep.Keys = append(rep.Keys, kr)
	}
	sort.SliceStable(rep.Keys, func(a, b int) bool {
		sa, sb := rep.Keys[a].MaxSeverity(), rep.Keys[b].MaxSeverity()
		if sa != sb {
			return sa > sb
		}
		return rep.Keys[a].Sources[0].Path < rep.Keys[b].Sources[0].Path
	})
	return rep
}

func keyReport(g *group, now time.Time) KeyReport {
	kr := KeyReport{
		Fingerprint: g.fp, Bits: g.n.BitLen(), Exponent: g.keys[0].E,
		Findings: slices.Clone(g.findings),
	}
	if len(g.keys) > 1 {
		kr.Findings = append(kr.Findings, duplicateFinding(g))
	}
	for _, k := range g.keys {
		kr.Sources = append(kr.Sources, sourceOf(k))
		if k.Cert != nil {
			kr.Findings = append(kr.Findings, certificateChecks(k.Cert, now)...)
		}
		if k.Kind == keys.KindPrivateKey {
			kr.Findings = append(kr.Findings, privateKeyFinding())
		}
	}
	slices.SortStableFunc(kr.Findings, func(a, b Finding) int { return int(b.Severity) - int(a.Severity) })
	return kr
}

func duplicateFinding(g *group) Finding {
	paths := make([]string, len(g.keys))
	for i, k := range g.keys {
		paths[i] = k.Source
	}
	return Finding{
		ID: IDDuplicateKey, Severity: High,
		Title:  fmt.Sprintf("the same RSA key is used by %d inputs", len(g.keys)),
		Detail: "Key reuse widens the blast radius of a compromise and breaks per-service key rotation.",
		Evidence: map[string]string{
			"count": fmt.Sprint(len(g.keys)),
			"seen":  strings.Join(paths, ","),
		},
	}
}

func sourceOf(k *keys.Key) Source {
	s := Source{Path: k.Source, Kind: string(k.Kind)}
	if c := k.Cert; c != nil {
		na := c.NotAfter.UTC()
		s.Subject, s.Issuer, s.NotAfter = c.Subject, c.Issuer, &na
		s.CertSHA256, s.Precert, s.SigAlgoName = c.SHA256, c.Precert, c.SignatureAlgorithm
	}
	return s
}
