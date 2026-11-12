// Command skillscheck verifies the project skills against .claude/skills.lock.json:
// every copy matches its hash, nothing unlocked is present, at most 12 skills are auto-triggered.
// The master catalogue lives outside this repository; `task skills:sync` in the workspace copies from it.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	maxAuto  = 12
	lockPath = ".claude/skills.lock.json"
	skillDir = ".claude/skills"
)

type entry struct {
	Trigger string            `json:"trigger"`
	Files   map[string]string `json:"files"`
	Tree    string            `json:"treeSha256"`
}

type lockFile struct {
	Skills map[string]entry `json:"skills"`
}

func main() {
	problems, err := run(".")
	if err != nil {
		fmt.Fprintln(os.Stderr, "skillscheck:", err)
		os.Exit(2)
	}
	for _, p := range problems {
		fmt.Fprintln(os.Stderr, "skillscheck:", p)
	}
	if len(problems) > 0 {
		os.Exit(1)
	}
	fmt.Println("skillscheck: project skills match the lock")
}

func run(root string) ([]string, error) {
	b, err := os.ReadFile(filepath.Join(root, lockPath))
	if err != nil {
		return nil, fmt.Errorf("read lock: %w", err)
	}
	var lock lockFile
	if err := json.Unmarshal(b, &lock); err != nil {
		return nil, fmt.Errorf("parse lock: %w", err)
	}
	var problems []string
	auto := 0
	for _, name := range sortedNames(lock.Skills) {
		e := lock.Skills[name]
		if e.Trigger == "auto" {
			auto++
		}
		got, err := hashDir(filepath.Join(root, skillDir, name))
		if err != nil {
			return nil, err
		}
		problems = append(problems, compare(name, got, e)...)
	}
	if auto > maxAuto {
		problems = append(problems, fmt.Sprintf("%d auto-triggered skills, the cap is %d", auto, maxAuto))
	}
	return append(problems, unlocked(root, lock)...), nil
}

func compare(name string, got map[string]string, e entry) []string {
	var problems []string
	for p, h := range e.Files {
		if got[p] != h {
			problems = append(problems, fmt.Sprintf("%s: %s differs from the lock or is missing", name, p))
		}
	}
	for p := range got {
		if _, ok := e.Files[p]; !ok {
			problems = append(problems, fmt.Sprintf("%s: %s is not in the lock", name, p))
		}
	}
	if len(problems) == 0 && treeHash(got) != e.Tree {
		problems = append(problems, name+": tree hash differs from the lock")
	}
	return problems
}

func unlocked(root string, lock lockFile) []string {
	entries, err := os.ReadDir(filepath.Join(root, skillDir))
	if err != nil {
		return []string{"cannot read " + skillDir}
	}
	var problems []string
	for _, d := range entries {
		if _, ok := lock.Skills[d.Name()]; d.IsDir() && !ok && !strings.HasPrefix(d.Name(), "_") {
			problems = append(problems, "skill not in the lock: "+d.Name())
		}
	}
	return problems
}

func hashDir(dir string) (map[string]string, error) {
	out := map[string]string{}
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(b)
		rel, _ := filepath.Rel(dir, p)
		out[filepath.ToSlash(rel)] = hex.EncodeToString(sum[:])
		return nil
	})
	return out, err
}

func treeHash(files map[string]string) string {
	h := sha256.New()
	for _, k := range sortedNames(files) {
		h.Write([]byte(k + "\x00" + files[k] + "\n"))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func sortedNames[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
