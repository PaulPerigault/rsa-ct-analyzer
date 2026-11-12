package ct

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"fmt"
)

// LatestEntries is the Start value meaning "the last Count entries of the log".
const LatestEntries int64 = -1

// CrawlOptions selects which part of the log to read.
type CrawlOptions struct {
	Start     int64 // first index, or LatestEntries
	Count     int64 // number of entries to read
	BatchSize int   // entries asked per request (logs usually cap at 256-1000)
	Progress  func(done, total int64)
}

// CrawlStats summarises a crawl.
type CrawlStats struct {
	TreeSize    int64
	Start       int64
	Fetched     int64
	RSA         int64
	NonRSA      int64
	ParseErrors int64
}

// Emit receives each RSA certificate found, as DER, with its log index.
type Emit func(index int64, der []byte, precert bool) error

// Crawl reads entries from the log and passes RSA certificates to emit.
// Malformed entries are counted, not fatal: logs contain odd data.
func Crawl(ctx context.Context, c *Client, opt CrawlOptions, emit Emit) (CrawlStats, error) {
	var st CrawlStats
	if opt.Count <= 0 || opt.BatchSize <= 0 {
		return st, errors.New("count and batch size must be positive")
	}
	sth, err := c.GetSTH(ctx)
	if err != nil {
		return st, err
	}
	st.TreeSize = sth.TreeSize
	st.Start = opt.Start
	if st.Start == LatestEntries {
		st.Start = max(sth.TreeSize-opt.Count, 0)
	}
	end := min(st.Start+opt.Count, sth.TreeSize) // exclusive
	if st.Start < 0 || st.Start >= end {
		return st, fmt.Errorf("nothing to read: start %d, tree size %d", st.Start, sth.TreeSize)
	}

	for next := st.Start; next < end; {
		last := min(next+int64(opt.BatchSize), end) - 1
		entries, err := c.GetEntries(ctx, next, last)
		if err != nil {
			return st, err
		}
		if len(entries) == 0 {
			return st, fmt.Errorf("log returned no entries for %d-%d", next, last)
		}
		for i, e := range entries {
			if err := handle(&st, next+int64(i), e, emit); err != nil {
				return st, err
			}
		}
		next += int64(len(entries))
		if opt.Progress != nil {
			opt.Progress(next-st.Start, end-st.Start)
		}
	}
	return st, nil
}

func handle(st *CrawlStats, index int64, e RawEntry, emit Emit) error {
	st.Fetched++
	der, precert, keep := decode(st, e)
	if !keep {
		return nil
	}
	return emit(index, der, precert)
}

// decode extracts the certificate of an entry and tallies it. keep is true
// only for well-formed RSA certificates.
func decode(st *CrawlStats, e RawEntry) (der []byte, precert, keep bool) {
	der, precert, err := ParseEntry(e)
	var cert *x509.Certificate
	if err == nil {
		cert, err = x509.ParseCertificate(der)
	}
	if err != nil {
		st.ParseErrors++
		return nil, false, false
	}
	if _, isRSA := cert.PublicKey.(*rsa.PublicKey); !isRSA {
		st.NonRSA++
		return nil, false, false
	}
	st.RSA++
	return der, precert, true
}
