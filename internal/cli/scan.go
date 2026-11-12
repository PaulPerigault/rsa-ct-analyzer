package cli

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/PaulPerigault/rsa-ct-analyzer/internal/analyze"
	"github.com/PaulPerigault/rsa-ct-analyzer/internal/keys"
	"github.com/PaulPerigault/rsa-ct-analyzer/internal/report"
	"github.com/PaulPerigault/rsa-ct-analyzer/internal/version"
)

type scanFlags struct {
	format        string
	output        string
	minSeverity   string
	failOn        string
	minBits       int
	fermatRounds  int
	noBatchGCD    bool
	revealFactors bool
	workers       int
}

func newScanCmd(streams IO) *cobra.Command {
	var f scanFlags
	cmd := &cobra.Command{
		Use:   "scan [path|-]...",
		Short: "Audit the RSA keys of local certificates and key files",
		Long: `Reads PEM or DER files (X.509 certificates, PKIX/PKCS#1 public keys, private
keys). Directories are walked recursively (.pem .crt .cer .der .key .pub .txt).
Use "-" to read standard input, e.g. after "crawl".

Exit status: 0 clean, 1 error, 2 a finding reached --fail-on.`,
		Example: `  rsa-ct-analyzer scan testdata/certs
  rsa-ct-analyzer scan --format json --fail-on high server.pem
  rsa-ct-analyzer crawl --count 5000 | rsa-ct-analyzer scan -`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScan(cmd, streams, args, f)
		},
	}
	fl := cmd.Flags()
	fl.StringVarP(&f.format, "format", "f", "table", "output format: table or json")
	fl.StringVarP(&f.output, "output", "o", "", "write the report to this file instead of stdout")
	fl.StringVar(&f.minSeverity, "min-severity", "info", "hide findings below: info, low, medium, high, critical")
	fl.StringVar(&f.failOn, "fail-on", "none", "exit 2 if a finding is at least: low, medium, high, critical, none")
	fl.IntVar(&f.minBits, "min-bits", 2048, "minimum acceptable modulus size in bits")
	fl.IntVar(&f.fermatRounds, "fermat-rounds", analyze.Defaults().FermatRounds, "Fermat close-primes iterations per key (0 disables)")
	fl.BoolVar(&f.noBatchGCD, "no-batch-gcd", false, "skip the shared-prime search across keys")
	fl.BoolVar(&f.revealFactors, "reveal-factors", false, "include recovered primes in the JSON evidence (sensitive)")
	fl.IntVar(&f.workers, "workers", runtime.NumCPU(), "parallel workers for per-key checks")
	return cmd
}

func runScan(cmd *cobra.Command, streams IO, paths []string, f scanFlags) error {
	format, err := report.ParseFormat(f.format)
	if err != nil {
		return err
	}
	minSev, err := analyze.ParseSeverity(f.minSeverity)
	if err != nil {
		return fmt.Errorf("--min-severity: %w", err)
	}
	failOn, enabled, err := parseFailOn(f.failOn)
	if err != nil {
		return err
	}

	loaded := keys.Load(paths, streams.In)
	for _, e := range loaded.Errors {
		fmt.Fprintln(streams.Err, "warning:", e)
	}
	if len(loaded.Keys) == 0 {
		return fmt.Errorf("no RSA key found in the given inputs (%d non-RSA skipped, %d unreadable)",
			loaded.Skipped, len(loaded.Errors))
	}

	opt := analyze.Defaults()
	opt.MinBits, opt.FermatRounds, opt.Workers = f.minBits, f.fermatRounds, f.workers
	opt.BatchGCD, opt.RevealFactors = !f.noBatchGCD, f.revealFactors
	opt.Tool, opt.Version = "rsa-ct-analyzer", version.Version
	rep, err := analyze.Analyze(cmd.Context(), loaded.Keys, loaded.Skipped, opt)
	if err != nil {
		return err
	}

	if err := writeReport(streams, rep, reportOutput{f.output, format, minSev}); err != nil {
		return err
	}
	if top, found := rep.MaxSeverity(); enabled && found && top >= failOn {
		return &ExitCodeError{Code: ExitFindings,
			Err: fmt.Errorf("findings at or above %q (highest: %s)", failOn, top)}
	}
	return nil
}

func parseFailOn(s string) (analyze.Severity, bool, error) {
	if s == "none" {
		return analyze.Info, false, nil
	}
	sev, err := analyze.ParseSeverity(s)
	if err != nil {
		return 0, false, fmt.Errorf("--fail-on: %w", err)
	}
	return sev, true, nil
}

// reportOutput says where and how to render a report.
type reportOutput struct {
	path   string // empty means stdout
	format report.Format
	floor  analyze.Severity
}

func writeReport(streams IO, rep *analyze.Report, out reportOutput) (err error) {
	w := streams.Out
	if out.path != "" {
		file, openErr := os.OpenFile(out.path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600) //nolint:gosec // user-chosen output path
		if openErr != nil {
			return openErr
		}
		defer func() { err = errors.Join(err, file.Close()) }()
		w = file
	}
	return report.Write(w, rep, out.format, out.floor)
}
