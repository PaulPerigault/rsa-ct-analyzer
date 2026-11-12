// Package cli wires the command line: scan, crawl, mock-log, version.
package cli

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/PaulPerigault/rsa-ct-analyzer/internal/version"
)

// Exit codes.
const (
	ExitOK       = 0
	ExitError    = 1 // runtime failure
	ExitFindings = 2 // findings at or above --fail-on
)

// ExitCodeError carries a specific process exit code.
type ExitCodeError struct {
	Code int
	Err  error
}

func (e *ExitCodeError) Error() string { return e.Err.Error() }
func (e *ExitCodeError) Unwrap() error { return e.Err }

// IO groups the streams a command uses, so tests can inject buffers.
type IO struct {
	In       io.Reader
	Out, Err io.Writer
}

// StdIO is the real process streams.
func StdIO() IO { return IO{In: os.Stdin, Out: os.Stdout, Err: os.Stderr} }

// NewRootCmd builds the command tree.
func NewRootCmd(streams IO) *cobra.Command {
	root := &cobra.Command{
		Use:   "rsa-ct-analyzer",
		Short: "Audit RSA keys from certificates and Certificate Transparency logs",
		Long: `rsa-ct-analyzer finds weak RSA keys: undersized moduli, bad exponents,
trivially factorable keys, duplicated keys, and primes shared between keys
(batch GCD). It reads local PEM/DER files or crawls a CT log.`,
		Version:       version.String(),
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetIn(streams.In)
	root.SetOut(streams.Out)
	root.SetErr(streams.Err)
	root.SetVersionTemplate("rsa-ct-analyzer {{.Version}}\n")
	root.AddCommand(newScanCmd(streams), newCrawlCmd(streams), newMockLogCmd(streams))
	return root
}

// envOr returns the environment variable key, or def when unset or empty.
func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envFloat(key string, def float64) float64 {
	if v, err := strconv.ParseFloat(os.Getenv(key), 64); err == nil {
		return v
	}
	return def
}

func envDuration(key string, def time.Duration) time.Duration {
	if v, err := time.ParseDuration(os.Getenv(key)); err == nil {
		return v
	}
	return def
}

func userAgent() string {
	return fmt.Sprintf("rsa-ct-analyzer/%s (+https://github.com/PaulPerigault/rsa-ct-analyzer)", version.Version)
}
