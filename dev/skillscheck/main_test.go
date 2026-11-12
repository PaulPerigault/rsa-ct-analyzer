package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writeSkill(t *testing.T, root, name, content string) {
	t.Helper()
	dir := filepath.Join(root, skillDir, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func lockFor(t *testing.T, root string, triggers map[string]string) {
	t.Helper()
	lock := lockFile{Skills: map[string]entry{}}
	for name, trig := range triggers {
		files, err := hashDir(filepath.Join(root, skillDir, name))
		if err != nil {
			t.Fatal(err)
		}
		lock.Skills[name] = entry{Trigger: trig, Files: files, Tree: treeHash(files)}
	}
	b, _ := json.Marshal(lock)
	if err := os.WriteFile(filepath.Join(root, lockPath), b, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRun(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(root string)
		want   int
	}{
		{"consistent", func(string) {}, 0},
		{"edited copy", func(r string) {
			_ = os.WriteFile(filepath.Join(r, skillDir, "a", "SKILL.md"), []byte("changed"), 0o644)
		}, 1},
		{"extra file", func(r string) {
			_ = os.WriteFile(filepath.Join(r, skillDir, "a", "x.md"), []byte("x"), 0o644)
		}, 1},
		{"unlocked skill", func(r string) { writeSkill(t, r, "b", "b") }, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := t.TempDir()
			writeSkill(t, root, "a", "a")
			lockFor(t, root, map[string]string{"a": "auto"})
			c.mutate(root)
			got, err := run(root)
			if err != nil || len(got) != c.want {
				t.Fatalf("got %v err %v, want %d problem(s)", got, err, c.want)
			}
		})
	}
}

func TestCap(t *testing.T) {
	root := t.TempDir()
	triggers := map[string]string{}
	for i := 0; i < maxAuto+1; i++ {
		name := string(rune('a' + i))
		writeSkill(t, root, name, name)
		triggers[name] = "auto"
	}
	lockFor(t, root, triggers)
	got, err := run(root)
	if err != nil || len(got) != 1 {
		t.Fatalf("got %v err %v, want exactly the cap problem", got, err)
	}
}
