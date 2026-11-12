package cli

import (
	"bufio"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/PaulPerigault/rsa-ct-analyzer/internal/ct"
)

type crawlFlags struct {
	logURL    string
	start     int64
	count     int64
	batchSize int
	rate      float64
	timeout   time.Duration
	retries   int
	out       string
}

func newCrawlCmd(streams IO) *cobra.Command {
	var f crawlFlags
	cmd := &cobra.Command{
		Use:   "crawl",
		Short: "Download RSA certificates from a Certificate Transparency log (RFC 6962)",
		Long: `Reads entries from a CT log and writes the RSA certificates as one PEM bundle,
ready for "scan". Non-RSA and malformed entries are counted and skipped.
Progress and statistics go to stderr so stdout can be piped.

Point --log-url at a public log, or at a local "mock-log" for an offline run.
Be polite with public logs: keep --rate-limit low.`,
		Example: `  rsa-ct-analyzer mock-log testdata/certs &
  rsa-ct-analyzer crawl --log-url http://127.0.0.1:8080 --count 100 -o crawled.pem
  rsa-ct-analyzer scan crawled.pem`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error { return runCrawl(cmd, streams, f) },
	}
	fl := cmd.Flags()
	fl.StringVar(&f.logURL, "log-url", envOr("RSA_CT_LOG_URL", ""), "CT log base URL (env RSA_CT_LOG_URL)")
	fl.Int64Var(&f.start, "start", ct.LatestEntries, "first entry index; -1 reads the latest --count entries")
	fl.Int64Var(&f.count, "count", 1000, "number of entries to read")
	fl.IntVar(&f.batchSize, "batch-size", 256, "entries requested per call")
	fl.Float64Var(&f.rate, "rate-limit", envFloat("RSA_CT_RATE_LIMIT", 5), "max requests per second (0 = unlimited)")
	fl.DurationVar(&f.timeout, "timeout", envDuration("RSA_CT_TIMEOUT", 30*time.Second), "per-request timeout")
	fl.IntVar(&f.retries, "retries", 4, "retries on HTTP 429, 5xx and network errors")
	fl.StringVarP(&f.out, "out", "o", "-", `output PEM file ("-" for stdout)`)
	return cmd
}

func runCrawl(cmd *cobra.Command, streams IO, f crawlFlags) (err error) {
	if f.logURL == "" {
		return errors.New("--log-url (or RSA_CT_LOG_URL) is required, e.g. https://oak.ct.letsencrypt.org/2026h2")
	}
	client, err := ct.NewClient(ct.ClientConfig{
		BaseURL: f.logURL, Timeout: f.timeout, RateLimit: f.rate,
		Retries: f.retries, UserAgent: userAgent(),
	})
	if err != nil {
		return err
	}

	sink := streams.Out
	if f.out != "-" {
		file, openErr := os.OpenFile(f.out, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600) //nolint:gosec // user-chosen output path
		if openErr != nil {
			return openErr
		}
		defer func() { err = errors.Join(err, file.Close()) }()
		sink = file
	}
	w := bufio.NewWriter(sink)

	opt := ct.CrawlOptions{
		Start: f.start, Count: f.count, BatchSize: f.batchSize,
		Progress: func(done, total int64) { fmt.Fprintf(streams.Err, "\rcrawl: %d/%d entries", done, total) },
	}
	stats, err := ct.Crawl(cmd.Context(), client, opt, func(_ int64, der []byte, _ bool) error {
		return pem.Encode(w, &pem.Block{Type: "CERTIFICATE", Bytes: der})
	})
	fmt.Fprintln(streams.Err)
	if err != nil {
		return err
	}
	if err := w.Flush(); err != nil {
		return err
	}
	fmt.Fprintf(streams.Err,
		"crawl: tree size %d, read %d from index %d: %d RSA kept, %d non-RSA, %d unparseable\n",
		stats.TreeSize, stats.Fetched, stats.Start, stats.RSA, stats.NonRSA, stats.ParseErrors)
	return nil
}
