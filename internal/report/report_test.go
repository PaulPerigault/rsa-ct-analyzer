package report_test

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/PaulPerigault/rsa-ct-analyzer/internal/analyze"
	"github.com/PaulPerigault/rsa-ct-analyzer/internal/keys"
	"github.com/PaulPerigault/rsa-ct-analyzer/internal/report"
)

func sampleReport(t *testing.T) *analyze.Report {
	t.Helper()
	loaded := keys.Load([]string{filepath.Join("..", "..", "testdata", "certs")}, nil)
	opt := analyze.Defaults()
	opt.Tool, opt.Version = "rsa-ct-analyzer", "test"
	rep, err := analyze.Analyze(context.Background(), loaded.Keys, loaded.Skipped, opt)
	if err != nil {
		t.Fatal(err)
	}
	return rep
}

func TestJSONIsValidAndComplete(t *testing.T) {
	rep := sampleReport(t)
	var buf bytes.Buffer
	if err := report.Write(&buf, rep, report.JSON, analyze.Info); err != nil {
		t.Fatal(err)
	}
	var back struct {
		Tool    string `json:"tool"`
		Summary struct {
			Distinct int            `json:"distinct_keys"`
			Severity map[string]int `json:"findings_by_severity"`
		} `json:"summary"`
		Keys []struct {
			Fingerprint string `json:"fingerprint_sha256"`
			Findings    []struct {
				ID       string `json:"id"`
				Severity string `json:"severity"`
			} `json:"findings"`
		} `json:"keys"`
	}
	if err := json.Unmarshal(buf.Bytes(), &back); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if back.Tool != "rsa-ct-analyzer" || back.Summary.Distinct != len(back.Keys) || back.Summary.Severity["critical"] == 0 {
		t.Errorf("unexpected content: %+v", back.Summary)
	}
	if len(back.Keys[0].Fingerprint) != 64 || back.Keys[0].Findings[0].Severity != "critical" {
		t.Error("keys must be sorted most severe first and carry a full fingerprint")
	}
}

func TestTableListsFindingsAndSummary(t *testing.T) {
	var buf bytes.Buffer
	if err := report.Write(&buf, sampleReport(t), report.Table, analyze.Info); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"SEVERITY", "shared-prime", "close-primes", "ok-2048.pem", "key(s) analyzed", "critical"} {
		if !strings.Contains(out, want) {
			t.Errorf("table lacks %q:\n%s", want, out)
		}
	}
}

func TestMinSeverityHidesNoise(t *testing.T) {
	var buf bytes.Buffer
	if err := report.Write(&buf, sampleReport(t), report.Table, analyze.High); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if strings.Contains(out, "expired") || strings.Contains(out, "public-exponent") {
		t.Errorf("findings below high must be hidden:\n%s", out)
	}
	if !strings.Contains(out, "key-size") {
		t.Errorf("high findings must remain:\n%s", out)
	}
}

func TestParseFormat(t *testing.T) {
	for _, s := range []string{"table", "JSON"} {
		if _, err := report.ParseFormat(s); err != nil {
			t.Errorf("%s: %v", s, err)
		}
	}
	if _, err := report.ParseFormat("xml"); err == nil {
		t.Error("xml must be rejected")
	}
}
