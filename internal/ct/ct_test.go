package ct_test

import (
	"context"
	"encoding/base64"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/PaulPerigault/rsa-ct-analyzer/internal/ct"
	"github.com/PaulPerigault/rsa-ct-analyzer/internal/ct/mocklog"
	"github.com/PaulPerigault/rsa-ct-analyzer/internal/keys"
)

func sampleEntries(t *testing.T) []mocklog.Entry {
	t.Helper()
	res := keys.Load([]string{filepath.Join("..", "..", "testdata", "certs")}, nil)
	if len(res.Keys) == 0 || len(res.Errors) != 0 {
		t.Fatalf("cannot load samples: %v", res.Errors)
	}
	var out []mocklog.Entry
	for i, k := range res.Keys {
		out = append(out, mocklog.Entry{DER: k.Raw, Precert: i%3 == 2})
	}
	return out
}

func newClient(t *testing.T, url string) *ct.Client {
	t.Helper()
	c, err := ct.NewClient(ct.ClientConfig{BaseURL: url, Retries: 3, Backoff: time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestCrawlWholeLogWithSmallPages(t *testing.T) {
	entries := sampleEntries(t)
	srv := httptest.NewServer(mocklog.Handler(entries, 4)) // log caps answers below our batch size
	defer srv.Close()

	var got, precerts int
	stats, err := ct.Crawl(context.Background(), newClient(t, srv.URL),
		ct.CrawlOptions{Start: 0, Count: 1000, BatchSize: 10},
		func(_ int64, _ []byte, precert bool) error {
			got++
			if precert {
				precerts++
			}
			return nil
		})
	if err != nil {
		t.Fatal(err)
	}
	if stats.Fetched != int64(len(entries)) || got != len(entries) || stats.RSA != int64(len(entries)) {
		t.Errorf("fetched %d, emitted %d, want %d", stats.Fetched, got, len(entries))
	}
	if precerts == 0 {
		t.Error("precertificate entries must be decoded from extra_data")
	}
}

func TestCrawlLatestEntries(t *testing.T) {
	entries := sampleEntries(t)
	srv := httptest.NewServer(mocklog.Handler(entries, 0))
	defer srv.Close()

	var first int64 = -1
	stats, err := ct.Crawl(context.Background(), newClient(t, srv.URL),
		ct.CrawlOptions{Start: ct.LatestEntries, Count: 3, BatchSize: 256},
		func(i int64, _ []byte, _ bool) error {
			if first < 0 {
				first = i
			}
			return nil
		})
	if err != nil {
		t.Fatal(err)
	}
	if want := int64(len(entries) - 3); first != want || stats.Start != want || stats.Fetched != 3 {
		t.Errorf("first index %d (start %d, fetched %d), want %d", first, stats.Start, stats.Fetched, want)
	}
}

func TestCrawlCountsMalformedEntries(t *testing.T) {
	ecPEM, err := os.ReadFile(filepath.Join("..", "..", "testdata", "certs", "ecdsa-p256.pem"))
	if err != nil {
		t.Fatal(err)
	}
	block, _ := pem.Decode(ecPEM)
	entries := append(sampleEntries(t),
		mocklog.Entry{DER: block.Bytes},
		mocklog.Entry{DER: []byte("not a certificate")})
	srv := httptest.NewServer(mocklog.Handler(entries, 0))
	defer srv.Close()

	stats, err := ct.Crawl(context.Background(), newClient(t, srv.URL),
		ct.CrawlOptions{Start: 0, Count: 1000, BatchSize: 256},
		func(int64, []byte, bool) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	if stats.ParseErrors != 1 || stats.NonRSA != 1 { // 1 garbage entry, 1 ECDSA sample
		t.Errorf("stats: %+v", stats)
	}
}

func TestCrawlValidatesRange(t *testing.T) {
	srv := httptest.NewServer(mocklog.Handler(sampleEntries(t), 0))
	defer srv.Close()
	c := newClient(t, srv.URL)
	noop := func(int64, []byte, bool) error { return nil }

	if _, err := ct.Crawl(context.Background(), c, ct.CrawlOptions{Start: 0, Count: 0, BatchSize: 10}, noop); err == nil {
		t.Error("zero count must be rejected")
	}
	if _, err := ct.Crawl(context.Background(), c, ct.CrawlOptions{Start: 9999, Count: 10, BatchSize: 10}, noop); err == nil {
		t.Error("a start beyond the tree size must be rejected")
	}
}

func TestClientRetriesTransientErrors(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		switch calls.Add(1) {
		case 1:
			w.WriteHeader(http.StatusServiceUnavailable)
		case 2:
			w.WriteHeader(http.StatusTooManyRequests)
		default:
			_, _ = w.Write([]byte(`{"tree_size": 42, "timestamp": 1}`))
		}
	}))
	defer srv.Close()

	sth, err := newClient(t, srv.URL).GetSTH(context.Background())
	if err != nil || sth.TreeSize != 42 || calls.Load() != 3 {
		t.Fatalf("sth=%v err=%v calls=%d", sth, err, calls.Load())
	}
}

func TestClientDoesNotRetryClientErrors(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		http.NotFound(w, nil)
	}))
	defer srv.Close()

	if _, err := newClient(t, srv.URL).GetSTH(context.Background()); err == nil || calls.Load() != 1 {
		t.Errorf("a 404 must fail at once, got err=%v after %d calls", err, calls.Load())
	}
}

func TestClientGivesUpAfterRetries(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()
	if _, err := newClient(t, srv.URL).GetSTH(context.Background()); err == nil {
		t.Error("persistent 502 must end in an error")
	}
}

func TestClientHonoursContext(t *testing.T) {
	srv := httptest.NewServer(mocklog.Handler(sampleEntries(t), 0))
	defer srv.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := newClient(t, srv.URL).GetSTH(ctx); err == nil {
		t.Error("cancelled context must abort the call")
	}
}

func TestNewClientRejectsBadURL(t *testing.T) {
	for _, u := range []string{"", "ftp://x/y", "not a url", "http://"} {
		if _, err := ct.NewClient(ct.ClientConfig{BaseURL: u}); err == nil {
			t.Errorf("%q must be rejected", u)
		}
	}
}

func TestRateLimitSpacesRequests(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tree_size":1}`))
	}))
	defer srv.Close()
	c, _ := ct.NewClient(ct.ClientConfig{BaseURL: srv.URL, RateLimit: 50}) // 20 ms apart

	start := time.Now()
	for range 4 {
		if _, err := c.GetSTH(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	if elapsed := time.Since(start); elapsed < 50*time.Millisecond {
		t.Errorf("4 calls at 50 req/s took %s, throttling is not applied", elapsed)
	}
}

func TestParseEntryErrors(t *testing.T) {
	b64 := base64.StdEncoding.EncodeToString
	cases := map[string]ct.RawEntry{
		"not base64":         {LeafInput: "!!!"},
		"too short":          {LeafInput: b64([]byte{0, 0, 1})},
		"unknown entry type": {LeafInput: b64(append(make([]byte, 10), 0, 9))},
		"x509 truncated":     {LeafInput: b64(append(make([]byte, 10), 0, 0, 0xff, 0xff, 0xff, 1))},
		"precert no extra":   {LeafInput: b64(append(make([]byte, 10), 0, 1)), ExtraData: ""},
	}
	for name, e := range cases {
		if _, _, err := ct.ParseEntry(e); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestParseEntryRoundTrip(t *testing.T) {
	der, err := os.ReadFile(filepath.Join("..", "..", "testdata", "certs", "ok-2048.der"))
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(mocklog.Handler([]mocklog.Entry{{DER: der}, {DER: der, Precert: true}}, 0))
	defer srv.Close()
	entries, err := newClient(t, srv.URL).GetEntries(context.Background(), 0, 1)
	if err != nil || len(entries) != 2 {
		t.Fatalf("entries=%d err=%v", len(entries), err)
	}
	for i, e := range entries {
		got, precert, err := ct.ParseEntry(e)
		if err != nil || string(got) != string(der) || precert != (i == 1) {
			t.Errorf("entry %d: err=%v precert=%v equal=%v", i, err, precert, string(got) == string(der))
		}
	}
}
