// Command commitlint checks a commit message file against the repository
// rules: Conventional Commits, lower-case subject of at most 72 characters,
// no tool signature. Used by the Lefthook commit-msg hook.
//
//	go run ./dev/commitlint <message-file>
package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

const maxSubject = 72

var (
	subjectRe = regexp.MustCompile(`^(feat|fix|docs|style|refactor|test|chore|ci|perf|revert)(\([a-z0-9._/-]+\))?!?: (.+)$`)
	signature = regexp.MustCompile(`(?i)^(co-authored-by:|signed-off-by:.*(claude|anthropic|copilot|bot)|.*generated (with|by) )|\b(claude code|noreply@anthropic\.com)\b`)
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: commitlint <message-file>")
		os.Exit(2)
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "commitlint:", err)
		os.Exit(2)
	}
	if problems := lint(string(data)); len(problems) > 0 {
		for _, p := range problems {
			fmt.Fprintln(os.Stderr, "commitlint:", p)
		}
		os.Exit(1)
	}
}

// lint returns every rule the message breaks.
func lint(msg string) []string {
	var out []string
	lines := strings.Split(strings.TrimRight(msg, "\n"), "\n")
	if strings.HasPrefix(lines[0], "Merge ") || strings.HasPrefix(lines[0], "fixup!") {
		return nil
	}
	m := subjectRe.FindStringSubmatch(lines[0])
	if m == nil {
		out = append(out, "first line must be `type(scope): subject` (feat fix docs style refactor test chore ci perf revert)")
	} else {
		subject := m[3]
		if len([]rune(subject)) > maxSubject {
			out = append(out, fmt.Sprintf("subject is %d characters, 72 at most", len([]rune(subject))))
		}
		if subject != strings.ToLower(subject) {
			out = append(out, "subject must be lower-case")
		}
		if strings.HasSuffix(subject, ".") {
			out = append(out, "subject must not end with a period")
		}
	}
	for _, l := range lines {
		if signature.MatchString(l) {
			out = append(out, "no tool signature allowed (Co-authored-by, Generated with...): "+l)
		}
	}
	return out
}
