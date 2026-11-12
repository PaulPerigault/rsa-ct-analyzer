@AGENTS.md

# Claude Code notes

- Start with `task test` to know the baseline, finish with `task verify`.
- Prefer `go test -race ./internal/<pkg>/...` while iterating on one package.
- `testdata/` is generated: edit `dev/gen-testdata/main.go`, run `task data:gen`, commit both. The README sample
  output is generated too: `task docs:readme`.
- Publishing is the owner's decision, announced in the current message: Claude Code prepares and Paul runs the
  publication script.
- `.claude/settings.json` denies the risky commands (push to `main`, force push, tags, releases, `gh repo edit`,
  `.env` access): a denial is a stop, not an obstacle to route around.
- No tool signature in commits or pull requests (the settings leave `attribution` empty).
