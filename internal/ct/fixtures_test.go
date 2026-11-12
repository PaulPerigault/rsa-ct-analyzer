package ct_test

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/PaulPerigault/rsa-ct-analyzer/internal/ct"
)

// Responses recorded from a public log (see testdata/ct/README.md). Order in
// the file: precert, x509, x509, precert.
var wantPrecert = []bool{true, false, false, true}

func recorded(t *testing.T) (sth, entries []byte) {
	t.Helper()
	dir := filepath.Join("..", "..", "testdata", "ct")
	var err error
	if sth, err = os.ReadFile(filepath.Join(dir, "get-sth.json")); err != nil {
		t.Fatal(err)
	}
	if entries, err = os.ReadFile(filepath.Join(dir, "get-entries.json")); err != nil {
		t.Fatal(err)
	}
	return sth, entries
}

func recordedEntries(t *testing.T) []ct.RawEntry {
	t.Helper()
	_, raw := recorded(t)
	var doc struct {
		Entries []ct.RawEntry `json:"entries"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	return doc.Entries
}

func TestParseRecordedEntries(t *testing.T) {
	entries := recordedEntries(t)
	if len(entries) != len(wantPrecert) {
		t.Fatalf("fixture holds %d entries, want %d", len(entries), len(wantPrecert))
	}
	for i, e := range entries {
		der, precert, err := ct.ParseEntry(e)
		if err != nil {
			t.Fatalf("entry %d: %v", i, err)
		}
		if precert != wantPrecert[i] {
			t.Errorf("entry %d: precert=%v, want %v", i, precert, wantPrecert[i])
		}
		cert, err := x509.ParseCertificate(der)
		if err != nil {
			t.Fatalf("entry %d: certificate does not parse: %v", i, err)
		}
		pub, ok := cert.PublicKey.(*rsa.PublicKey)
		if !ok || pub.N.BitLen() != 2048 || pub.E != 65537 {
			t.Errorf("entry %d: want an RSA-2048 key with e=65537, got %T", i, cert.PublicKey)
		}
	}
}

func TestParseRecordedEntryMalformed(t *testing.T) {
	x509Entry := recordedEntries(t)[1]
	precertEntry := recordedEntries(t)[0]

	leaf := func(e ct.RawEntry) []byte {
		b, err := base64.StdEncoding.DecodeString(e.LeafInput)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	enc := base64.StdEncoding.EncodeToString
	with := func(e ct.RawEntry, mutate func([]byte) []byte) ct.RawEntry {
		e.LeafInput = enc(mutate(leaf(e)))
		return e
	}

	cases := []struct {
		name string
		e    ct.RawEntry
	}{
		{"leaf truncated in the header", with(x509Entry, func(b []byte) []byte { return b[:8] })},
		{"leaf truncated in the certificate", with(x509Entry, func(b []byte) []byte { return b[:200] })},
		{"leaf missing extensions", with(x509Entry, func(b []byte) []byte { return b[:len(b)-2] })},
		{"leaf with trailing bytes", with(x509Entry, func(b []byte) []byte { return append(b, 1, 2, 3) })},
		{"certificate length larger than the leaf", with(x509Entry, func(b []byte) []byte { b[12], b[13] = 0xff, 0xff; return b })},
		{"unknown entry type", with(x509Entry, func(b []byte) []byte { b[10], b[11] = 0, 7; return b })},
		{"unsupported version", with(x509Entry, func(b []byte) []byte { b[0] = 1; return b })},
		{"unsupported leaf type", with(x509Entry, func(b []byte) []byte { b[1] = 1; return b })},
		{"precert leaf truncated before the TBS", with(precertEntry, func(b []byte) []byte { return b[:20] })},
		{"precert leaf extensions inconsistent", with(precertEntry, func(b []byte) []byte { b[len(b)-1] = 9; return b })},
		{"precert extra_data empty", ct.RawEntry{LeafInput: precertEntry.LeafInput}},
		{"precert extra_data truncated", ct.RawEntry{LeafInput: precertEntry.LeafInput, ExtraData: precertEntry.ExtraData[:40]}},
		{"precert extra_data chain length inconsistent", ct.RawEntry{LeafInput: precertEntry.LeafInput, ExtraData: enc([]byte{0, 0, 1, 0xAA, 0, 0, 9})}},
		{"leaf_input beyond the size bound", ct.RawEntry{LeafInput: strings.Repeat("A", 2<<20)}},
		{"extra_data beyond the size bound", ct.RawEntry{LeafInput: x509Entry.LeafInput, ExtraData: strings.Repeat("A", 2<<20)}},
		{"leaf_input not base64", ct.RawEntry{LeafInput: "###"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if der, _, err := ct.ParseEntry(tc.e); err == nil {
				t.Errorf("expected an error, got a %d-byte certificate", len(der))
			}
		})
	}
}

// replayLog answers like the recorded log: any get-entries range returns the
// recorded entries, whatever the indices.
func replayLog(t *testing.T) *httptest.Server {
	t.Helper()
	sth, entries := recorded(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/ct/v1/get-sth", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(sth)
	})
	mux.HandleFunc("/ct/v1/get-entries", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(entries)
	})
	return httptest.NewServer(mux)
}

func TestCrawlReplaysRecordedLog(t *testing.T) {
	srv := replayLog(t)
	defer srv.Close()

	var precerts, total int
	stats, err := ct.Crawl(context.Background(), newClient(t, srv.URL),
		ct.CrawlOptions{Start: 0, Count: 4, BatchSize: 4},
		func(_ int64, _ []byte, precert bool) error {
			total++
			if precert {
				precerts++
			}
			return nil
		})
	if err != nil {
		t.Fatal(err)
	}
	if stats.TreeSize != 3501428149 || stats.RSA != 4 || total != 4 || precerts != 2 || stats.ParseErrors != 0 {
		t.Errorf("stats %+v, emitted %d (precerts %d)", stats, total, precerts)
	}
}
