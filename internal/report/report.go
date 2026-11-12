// Package report renders an analyze.Report as JSON or as a text table.
package report

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/PaulPerigault/rsa-ct-analyzer/internal/analyze"
)

// Format names an output format.
type Format string

// Supported formats.
const (
	Table Format = "table"
	JSON  Format = "json"
)

// ParseFormat validates a --format value.
func ParseFormat(s string) (Format, error) {
	switch f := Format(strings.ToLower(s)); f {
	case Table, JSON:
		return f, nil
	}
	return "", fmt.Errorf("unknown format %q (want table or json)", s)
}

// Write renders rep, hiding findings below minSeverity.
func Write(w io.Writer, rep *analyze.Report, f Format, minSeverity analyze.Severity) error {
	filtered := filter(rep, minSeverity)
	if f == JSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(filtered)
	}
	return writeTable(w, filtered)
}

// filter returns a copy of rep without findings under floor. Keys left with no
// finding are dropped only when floor is above Info, so a plain run still lists
// every key it analyzed.
func filter(rep *analyze.Report, floor analyze.Severity) *analyze.Report {
	out := *rep
	out.Keys = nil
	for _, k := range rep.Keys {
		kept := k
		kept.Findings = nil
		for _, f := range k.Findings {
			if f.Severity >= floor {
				kept.Findings = append(kept.Findings, f)
			}
		}
		if len(kept.Findings) == 0 && floor > analyze.Info {
			continue
		}
		out.Keys = append(out.Keys, kept)
	}
	return &out
}

func writeTable(w io.Writer, rep *analyze.Report) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "SEVERITY\tFINDING\tKEY\tBITS\tSOURCE\tDETAIL")
	for _, k := range rep.Keys {
		if len(k.Findings) == 0 {
			fmt.Fprintf(tw, "ok\t-\t%s\t%d\t%s\t-\n", k.Fingerprint[:12], k.Bits, sourceList(k))
			continue
		}
		for _, f := range k.Findings {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%s\t%s\n",
				f.Severity, f.ID, k.Fingerprint[:12], k.Bits, sourceList(k), f.Title)
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	_, err := fmt.Fprint(w, "\n", summaryLine(rep))
	return err
}

func sourceList(k analyze.KeyReport) string {
	const maxShown = 2
	paths := make([]string, 0, len(k.Sources))
	for _, s := range k.Sources {
		paths = append(paths, s.Path)
	}
	if len(paths) > maxShown {
		return fmt.Sprintf("%s (+%d more)", strings.Join(paths[:maxShown], ", "), len(paths)-maxShown)
	}
	return strings.Join(paths, ", ")
}

func summaryLine(rep *analyze.Report) string {
	s := rep.Summary
	names := make([]string, 0, len(s.BySeverity))
	for name := range s.BySeverity {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		a, _ := analyze.ParseSeverity(names[i])
		b, _ := analyze.ParseSeverity(names[j])
		return a > b
	})
	parts := make([]string, len(names))
	for i, n := range names {
		parts[i] = fmt.Sprintf("%d %s", s.BySeverity[n], n)
	}
	findings := "no finding"
	if len(parts) > 0 {
		findings = strings.Join(parts, ", ")
	}
	return fmt.Sprintf("%d key(s) analyzed (%d distinct, %d non-RSA skipped): %s\n",
		s.KeysAnalyzed, s.DistinctKeys, s.NonRSASkipped, findings)
}
