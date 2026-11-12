// Command checkdocs fails when the documentation mentions something that no
// longer exists: a `task` target, a CLI command or flag, a repository path or
// a relative Markdown link. Run from the repository root (`task check:docs`).
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/PaulPerigault/rsa-ct-analyzer/internal/cli"
)

var (
	codeSpan   = regexp.MustCompile("`([^`\n]+)`")
	fenceStart = regexp.MustCompile("^\\s*```")
	linkRe     = regexp.MustCompile(`\]\(([^)\s#]+)(?:#[^)]*)?\)`)
	taskRe     = regexp.MustCompile(`(?:^|\s)(?:go tool -modfile=tools/go\.mod )?task ([a-z][a-z0-9:-]*)`)
	cliRe      = regexp.MustCompile(`rsa-ct-analyzer ([a-z][a-z-]*)((?: [^|;&\n]*)?)`)
	taskDefRe  = regexp.MustCompile(`^  ([a-z][a-z0-9:-]*):\s*$`)
	pathRe     = regexp.MustCompile(`^\.?[A-Za-z0-9_.-]+(/[A-Za-z0-9_.-]+)+/?$`)
	flagRe     = regexp.MustCompile(`(?:^|\s)--([a-z][a-z-]*)`)
)

// ignored top-level paths are build output or local files, absent from a checkout.
var ignored = map[string]bool{"bin": true, "dist": true, ".env": true, ".tools": true, "work": true}

func main() {
	problems, err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "checkdocs:", err)
		os.Exit(2)
	}
	if len(problems) > 0 {
		sort.Strings(problems)
		for _, p := range problems {
			fmt.Fprintln(os.Stderr, "checkdocs:", p)
		}
		os.Exit(1)
	}
	fmt.Println("checkdocs: documentation matches the repository")
}

func run() ([]string, error) {
	tasks, err := taskNames("Taskfile.yml")
	if err != nil {
		return nil, err
	}
	root := cli.NewRootCmd(cli.StdIO())
	var problems []string
	err = filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if filepath.ToSlash(path) == ".claude/skills" {
				return filepath.SkipDir // vendored third-party text, checked by skillscheck
			}
			switch d.Name() {
			case ".git", "node_modules", "bin", "dist", "testdata", "backlog":
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".md") {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			problems = append(problems, checkFile(path, string(data), tasks, root)...)
		}
		return nil
	})
	if err != nil {
		return problems, err
	}
	readme, err := os.ReadFile("README.md")
	if err != nil {
		return problems, err
	}
	return append(problems, checkContext(string(readme))...), nil
}

func taskNames(file string) (map[string]bool, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	names := map[string]bool{}
	inTasks := false
	for _, l := range strings.Split(string(data), "\n") {
		l = strings.TrimRight(l, "\r")
		if strings.HasPrefix(l, "tasks:") {
			inTasks = true
			continue
		}
		if m := taskDefRe.FindStringSubmatch(l); inTasks && m != nil {
			names[m[1]] = true
		}
	}
	return names, nil
}

// checkFile inspects code spans and fenced blocks of one Markdown file.
func checkFile(path, text string, tasks map[string]bool, root *cobra.Command) []string {
	var problems []string
	dir := filepath.Dir(path)
	inFence := false
	for n, line := range strings.Split(text, "\n") {
		where := fmt.Sprintf("%s:%d", filepath.ToSlash(path), n+1)
		if fenceStart.MatchString(line) {
			inFence = !inFence
			continue
		}
		var snippets []string
		if inFence {
			snippets = []string{strings.TrimPrefix(strings.TrimSpace(line), "$ ")}
		} else {
			for _, m := range codeSpan.FindAllStringSubmatch(line, -1) {
				snippets = append(snippets, m[1])
			}
			problems = append(problems, checkLinks(where, dir, line)...)
		}
		for _, s := range snippets {
			problems = append(problems, checkSnippet(where, s, tasks, root)...)
		}
	}
	return problems
}

func checkLinks(where, dir, line string) []string {
	var out []string
	for _, m := range linkRe.FindAllStringSubmatch(line, -1) {
		target := m[1]
		if strings.Contains(target, "://") || strings.HasPrefix(target, "mailto:") {
			continue
		}
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(target))); err != nil {
			out = append(out, fmt.Sprintf("%s: link to missing file %q", where, target))
		}
	}
	return out
}

func checkSnippet(where, s string, tasks map[string]bool, root *cobra.Command) []string {
	var out []string
	if m := taskRe.FindStringSubmatch(s); m != nil && !tasks[m[1]] {
		out = append(out, fmt.Sprintf("%s: unknown task %q", where, m[1]))
	}
	if m := cliRe.FindStringSubmatch(s); m != nil {
		out = append(out, checkCLI(where, m[1], m[2], root)...)
	}
	if pathRe.MatchString(s) && !strings.Contains(s, "*") {
		first := strings.SplitN(strings.TrimPrefix(s, "./"), "/", 2)[0]
		if _, err := os.Stat(first); err == nil && !ignored[first] {
			if _, err := os.Stat(strings.TrimSuffix(filepath.FromSlash(s), "/")); err != nil {
				out = append(out, fmt.Sprintf("%s: path %q does not exist", where, s))
			}
		}
	}
	return out
}

func checkCLI(where, sub, rest string, root *cobra.Command) []string {
	cmd, _, err := root.Find([]string{sub})
	if err != nil || cmd == root {
		return []string{fmt.Sprintf("%s: unknown command `rsa-ct-analyzer %s`", where, sub)}
	}
	var out []string
	for _, m := range flagRe.FindAllStringSubmatch(rest, -1) {
		if name := m[1]; cmd.Flags().Lookup(name) == nil && name != "help" && name != "version" {
			out = append(out, fmt.Sprintf("%s: `%s` has no flag --%s", where, sub, name))
		}
	}
	return out
}
