package main

import (
	"regexp"
	"strings"
)

var periodRe = regexp.MustCompile(`\b\d{4}-(0[1-9]|1[0-2]) to \d{4}-(0[1-9]|1[0-2])\b`)

// checkContext requires the README to carry a "Context" section right after "Overview", holding a
// period written YYYY-MM to YYYY-MM.
func checkContext(readme string) []string {
	var headings []string
	body := map[string]string{}
	cur := ""
	for _, l := range strings.Split(readme, "\n") {
		if strings.HasPrefix(l, "## ") {
			cur = strings.TrimSpace(strings.TrimPrefix(l, "## "))
			headings = append(headings, cur)
			continue
		}
		if cur != "" {
			body[cur] += l + "\n"
		}
	}
	var problems []string
	if _, ok := body["Context"]; !ok {
		return []string{"README.md: missing the \"Context\" section"}
	}
	if !followsOverview(headings) {
		problems = append(problems, "README.md: \"Context\" must come right after \"Overview\"")
	}
	if !periodRe.MatchString(body["Context"]) {
		problems = append(problems, "README.md: \"Context\" must state a period written YYYY-MM to YYYY-MM")
	}
	return problems
}

func followsOverview(headings []string) bool {
	for i, h := range headings {
		if h == "Overview" {
			return i+1 < len(headings) && headings[i+1] == "Context"
		}
	}
	return false
}
