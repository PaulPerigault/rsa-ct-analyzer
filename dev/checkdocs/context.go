package main

import (
	"regexp"
	"strings"
)

// dateTok is a year (YYYY) or a month (YYYY-MM). A project whose original month is unknown only states the year.
const dateTok = `(?:19|20)\d{2}(?:-(?:0[1-9]|1[0-2]))?`

var (
	periodRe = regexp.MustCompile(`\b(` + dateTok + `) to (` + dateTok + `)\b`)
	looseRe  = regexp.MustCompile(`\b(?:19|20)\d{2}(?:-\d{2})?\b`)
	strictRe = regexp.MustCompile(`^` + dateTok + `$`)
)

// checkContext requires the README to carry a "Context" section right after "Overview", holding a
// date: a year (YYYY), a month (YYYY-MM) or a period "A to B" where A and B are each a year or a month
// and A does not come after B.
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
	if p := dateProblem(body["Context"]); p != "" {
		problems = append(problems, "README.md: "+p)
	}
	return problems
}

// dateProblem says what is wrong with the dates of the Context section ("" when it is fine).
func dateProblem(ctx string) string {
	if m := periodRe.FindStringSubmatch(ctx); m != nil {
		if first(m[1]) > last(m[2]) {
			return "\"Context\" period " + m[1] + " to " + m[2] + " ends before it starts"
		}
		return ""
	}
	for _, d := range looseRe.FindAllString(ctx, -1) {
		if strictRe.MatchString(d) {
			return ""
		}
	}
	return "\"Context\" must state a year (YYYY), a month (YYYY-MM) or a period (YYYY[-MM] to YYYY[-MM])"
}

// first and last widen a year to its first and last month so that a year and a month compare.
func first(d string) string { return pad(d, "01") }
func last(d string) string  { return pad(d, "12") }

func pad(d, month string) string {
	if len(d) == 4 {
		return d + "-" + month
	}
	return d
}

func followsOverview(headings []string) bool {
	for i, h := range headings {
		if h == "Overview" {
			return i+1 < len(headings) && headings[i+1] == "Context"
		}
	}
	return false
}
