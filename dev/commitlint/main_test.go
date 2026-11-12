package main

import (
	"strings"
	"testing"
)

func TestLint(t *testing.T) {
	long := "feat: " + strings.Repeat("a", 73)
	cases := []struct {
		name string
		msg  string
		ok   bool
	}{
		{"plain", "feat: add resumable crawl\n", true},
		{"scope", "fix(ct): bound the response size\n\nbody\n", true},
		{"breaking", "feat(report)!: rename a field\n", true},
		{"merge", "Merge branch 'x'\n", true},
		{"no type", "add something\n", false},
		{"unknown type", "feature: add something\n", false},
		{"upper case", "feat: Add something\n", false},
		{"period", "feat: add something.\n", false},
		{"too long", long + "\n", false},
		{"co-author", "feat: x\n\nCo-authored-by: Someone <a@b.c>\n", false},
		{"generated with", "feat: x\n\nGenerated with a tool\n", false},
		{"tool name", "feat: x\n\nmade by Claude Code\n", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := len(lint(tc.msg)) == 0; got != tc.ok {
				t.Errorf("lint(%q) ok=%v, want %v (%v)", tc.msg, got, tc.ok, lint(tc.msg))
			}
		})
	}
}
