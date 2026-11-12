package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/PaulPerigault/rsa-ct-analyzer/internal/cli"
	"github.com/PaulPerigault/rsa-ct-analyzer/internal/ct/mocklog"
	"github.com/PaulPerigault/rsa-ct-analyzer/internal/keys"
)

var certsDir = filepath.Join("..", "..", "testdata", "certs")

// execute runs the CLI in-process and returns stdout, stderr and the exit code.
func execute(t *testing.T, stdin string, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	var out, errb bytes.Buffer
	root := cli.NewRootCmd(cli.IO{In: strings.NewReader(stdin), Out: &out, Err: &errb})
	root.SetArgs(args)
	err := root.ExecuteContext(context.Background())
	var exit *cli.ExitCodeError
	switch {
	case err == nil:
		code = cli.ExitOK
	case errors.As(err, &exit):
		code = exit.Code
	default:
		code = cli.ExitError
		errb.WriteString(err.Error())
	}
	return out.String(), errb.String(), code
}

func TestScanTableDefaultExitsZero(t *testing.T) {
	out, _, code := execute(t, "", "scan", certsDir)
	if code != cli.ExitOK || !strings.Contains(out, "shared-prime") {
		t.Errorf("code %d, output:\n%s", code, out)
	}
}

func TestScanFailOn(t *testing.T) {
	cases := []struct {
		failOn string
		paths  []string
		want   int
	}{
		{"critical", []string{certsDir}, cli.ExitFindings},
		{"high", []string{filepath.Join(certsDir, "ok-4096.pem")}, cli.ExitOK},
		{"low", []string{filepath.Join(certsDir, "expired.pem")}, cli.ExitFindings},
		{"critical", []string{filepath.Join(certsDir, "expired.pem")}, cli.ExitOK},
		{"none", []string{certsDir}, cli.ExitOK},
	}
	for _, tc := range cases {
		args := append([]string{"scan", "--fail-on", tc.failOn}, tc.paths...)
		if _, _, code := execute(t, "", args...); code != tc.want {
			t.Errorf("--fail-on %s on %v: exit %d, want %d", tc.failOn, tc.paths, code, tc.want)
		}
	}
}

func TestScanJSONToFile(t *testing.T) {
	file := filepath.Join(t.TempDir(), "report.json")
	out, _, code := execute(t, "", "scan", "-f", "json", "-o", file, certsDir)
	if code != cli.ExitOK || out != "" {
		t.Fatalf("code %d, stdout %q (report must go to the file)", code, out)
	}
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var rep struct {
		Summary struct {
			Keys int `json:"keys_analyzed"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(data, &rep); err != nil || rep.Summary.Keys != 13 {
		t.Errorf("report: keys=%d err=%v", rep.Summary.Keys, err)
	}
}

func TestScanFromStdin(t *testing.T) {
	pem, _ := os.ReadFile(filepath.Join(certsDir, "weak-512.pem"))
	out, _, code := execute(t, string(pem), "scan", "-")
	if code != cli.ExitOK || !strings.Contains(out, "key-size") || !strings.Contains(out, "stdin") {
		t.Errorf("code %d, output:\n%s", code, out)
	}
}

func TestScanErrors(t *testing.T) {
	cases := map[string][]string{
		"no argument":   {"scan"},
		"missing file":  {"scan", "nope.pem"},
		"bad format":    {"scan", "-f", "xml", certsDir},
		"bad severity":  {"scan", "--min-severity", "urgent", certsDir},
		"bad fail-on":   {"scan", "--fail-on", "urgent", certsDir},
		"no key inside": {"scan", filepath.Join(certsDir, "ecdsa-p256.pem")},
	}
	for name, args := range cases {
		if _, _, code := execute(t, "", args...); code != cli.ExitError {
			t.Errorf("%s: exit %d, want %d", name, code, cli.ExitError)
		}
	}
}

func TestVersionFlag(t *testing.T) {
	out, _, code := execute(t, "", "--version")
	if code != cli.ExitOK || !strings.HasPrefix(out, "rsa-ct-analyzer ") {
		t.Errorf("code %d, output %q", code, out)
	}
}

// TestCrawlThenScan runs the whole pipeline offline: simulated CT log, crawl
// into a PEM bundle, scan the bundle.
func TestCrawlThenScan(t *testing.T) {
	loaded := keys.Load([]string{certsDir}, nil)
	var entries []mocklog.Entry
	for _, k := range loaded.Keys {
		entries = append(entries, mocklog.Entry{DER: k.Raw})
	}
	srv := httptest.NewServer(mocklog.Handler(entries, 5))
	defer srv.Close()

	bundle := filepath.Join(t.TempDir(), "crawled.pem")
	_, stderr, code := execute(t, "", "crawl", "--log-url", srv.URL, "--start", "0",
		"--count", "100", "--batch-size", "7", "--rate-limit", "0", "-o", bundle)
	if code != cli.ExitOK || !strings.Contains(stderr, "13 RSA kept") {
		t.Fatalf("crawl: code %d, stderr:\n%s", code, stderr)
	}

	out, _, code := execute(t, "", "scan", "--fail-on", "critical", bundle)
	if code != cli.ExitFindings || !strings.Contains(out, "shared-prime") {
		t.Errorf("scan of crawled bundle: code %d, output:\n%s", code, out)
	}
}

func TestCrawlRequiresLogURL(t *testing.T) {
	t.Setenv("RSA_CT_LOG_URL", "")
	if _, stderr, code := execute(t, "", "crawl"); code != cli.ExitError || !strings.Contains(stderr, "--log-url") {
		t.Errorf("code %d, stderr %q", code, stderr)
	}
}

func TestCrawlReadsLogURLFromEnv(t *testing.T) {
	srv := httptest.NewServer(mocklog.Handler([]mocklog.Entry{{DER: mustCert(t, "ok-2048.der")}}, 0))
	defer srv.Close()
	t.Setenv("RSA_CT_LOG_URL", srv.URL)
	out, _, code := execute(t, "", "crawl", "--count", "1", "--rate-limit", "0")
	if code != cli.ExitOK || !strings.Contains(out, "BEGIN CERTIFICATE") {
		t.Errorf("code %d, stdout %q", code, out)
	}
}

func TestMockLogServesUntilCancelled(t *testing.T) {
	var out, errb bytes.Buffer
	root := cli.NewRootCmd(cli.IO{In: strings.NewReader(""), Out: &out, Err: &errb})
	root.SetArgs([]string{"mock-log", "--addr", "127.0.0.1:0", certsDir})

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	if err := root.ExecuteContext(ctx); err != nil {
		t.Fatalf("mock-log: %v", err)
	}
	if !strings.Contains(errb.String(), "serving 13 certificates") {
		t.Errorf("stderr: %s", errb.String())
	}
}

func TestMockLogNeedsCertificates(t *testing.T) {
	if _, _, code := execute(t, "", "mock-log", filepath.Join("..", "..", "testdata", "keys")); code != cli.ExitError {
		t.Errorf("public keys are not certificates: exit %d, want %d", code, cli.ExitError)
	}
}

func mustCert(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(certsDir, name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}
