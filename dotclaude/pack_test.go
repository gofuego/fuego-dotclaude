package dotclaude_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gofuego/fuego-dotclaude/dotclaude"
	"github.com/gofuego/fuego/engine"
)

// writeFile writes body to name under dir, creating parent directories.
func writeFile(t *testing.T, dir, name, body string) {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// buildFixture writes a representative .claude tree and builds it with only the
// pack on a vanilla engine, returning the output directory.
func buildFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	content := filepath.Join(dir, ".claude")
	out := filepath.Join(dir, "out")

	writeFile(t, content, "agents/code-reviewer.md", `---
name: code-reviewer
description: Reviews code for correctness and style.
model: sonnet
tools: Read, Grep, Bash
---

# Code Reviewer
Reviews diffs and flags **risky** changes.
`)
	writeFile(t, content, "skills/code-review/SKILL.md", `---
name: code-review
description: Structured review workflow.
---

# Code Review
See the bundled checklist.
`)
	writeFile(t, content, "skills/code-review/checklist.md", `# Checklist
- Correctness first.
`)
	writeFile(t, content, "skills/code-review/lint.sh", "#!/usr/bin/env bash\necho lint\n")
	writeFile(t, content, "commands/git/commit.md", `---
description: Commit staged changes.
---
Create a conventional commit.
`)
	writeFile(t, content, "output-styles/concise.md", `---
name: Concise
description: Terse responses.
---
Lead with the answer.
`)
	writeFile(t, content, "CLAUDE.md", "# Workspace memory\nProject conventions live here.\n")
	writeFile(t, content, ".mcp.json", `{
  "mcpServers": {
    "github": { "command": "npx", "args": ["-y", "server-github"], "env": {"TOKEN": "x"} },
    "docs": { "url": "https://docs.example/mcp", "type": "sse" }
  }
}`)
	writeFile(t, content, "settings.json", `{
  "model": "opus",
  "permissions": { "allow": ["Bash(ls)"], "deny": ["Read(./.env)"], "defaultMode": "acceptEdits" },
  "env": { "FOO": "bar" },
  "cleanupPeriodDays": 30
}`)
	writeFile(t, content, "settings.local.json", `{ "model": "sonnet" }`)

	eng := engine.New()
	eng.Use(dotclaude.Pack())
	if err := eng.Build(context.Background(), engine.BuildOptions{
		ContentDir: content,
		OutputDir:  out,
		SiteName:   "Workspace",
	}); err != nil {
		t.Fatalf("vanilla pack build failed: %v", err)
	}
	return out
}

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return string(b)
}

// TestPackRendersEveryArtifactType is the vanilla-pack contract: a plain Fuego
// engine with only eng.Use(dotclaude.Pack()) renders one page per Markdown
// artifact kind, plus the pack's static asset.
func TestPackRendersEveryArtifactType(t *testing.T) {
	out := buildFixture(t)

	cases := []struct {
		page     string   // output path relative to out
		contains []string // substrings the rendered page must contain
	}{
		{"agents/code-reviewer/index.html", []string{"code-reviewer", "sonnet", "Grep", "<strong>risky</strong>"}},
		{"skills/code-review/index.html", []string{"code-review", "Structured review workflow"}},
		{"skills/code-review/checklist/index.html", []string{"Checklist"}},
		{"commands/git/commit/index.html", []string{"/git:commit", "conventional commit"}},
		{"output-styles/concise/index.html", []string{"Concise", "Lead with the answer"}},
		{"memory/CLAUDE/index.html", []string{"Workspace memory"}},
	}
	for _, c := range cases {
		t.Run(c.page, func(t *testing.T) {
			html := read(t, filepath.Join(out, c.page))
			for _, want := range c.contains {
				if !strings.Contains(html, want) {
					t.Errorf("page %s missing %q", c.page, want)
				}
			}
		})
	}

	// Pack static asset travels with the pack.
	if _, err := os.Stat(filepath.Join(out, "style.css")); err != nil {
		t.Errorf("expected pack static asset style.css: %v", err)
	}
}

// TestJSONPages checks the .mcp.json and settings pages render with structured
// content (server cards, curated sections) and a raw-JSON block.
func TestJSONPages(t *testing.T) {
	out := buildFixture(t)

	mcp := read(t, filepath.Join(out, "mcp/index.html"))
	for _, want := range []string{
		`id="server-github"`, // one card per server
		`id="server-docs"`,
		"sse",        // transport inferred/explicit
		"TOKEN",      // env surfaced
		"Raw",        // collapsible raw block
	} {
		if !strings.Contains(mcp, want) {
			t.Errorf("mcp page missing %q", want)
		}
	}

	settings := read(t, filepath.Join(out, "settings/index.html"))
	for _, want := range []string{
		"opus",               // model
		"acceptEdits",        // permissions default mode
		"Bash(ls)",           // allow rule
		"cleanupPeriodDays",  // long-tail key in generic table
		"30",                 // its value
		"Raw JSON",           // collapsible raw block
	} {
		if !strings.Contains(settings, want) {
			t.Errorf("settings page missing %q", want)
		}
	}

	local := read(t, filepath.Join(out, "settings/local/index.html"))
	if !strings.Contains(local, "sonnet") {
		t.Error("settings.local page should render its model")
	}
}

// TestMalformedJSONDegrades verifies a broken JSON file does not fail the build:
// the page still renders, reporting the error and showing the raw source.
func TestMalformedJSONDegrades(t *testing.T) {
	dir := t.TempDir()
	content := filepath.Join(dir, ".claude")
	out := filepath.Join(dir, "out")
	writeFile(t, content, ".mcp.json", `{ "mcpServers": { broken`)

	eng := engine.New()
	eng.Use(dotclaude.Pack())
	if err := eng.Build(context.Background(), engine.BuildOptions{ContentDir: content, OutputDir: out, SiteName: "X"}); err != nil {
		t.Fatalf("build should not fail on malformed JSON: %v", err)
	}

	mcp := read(t, filepath.Join(out, "mcp/index.html"))
	if !strings.Contains(mcp, "Could not parse") {
		t.Error("malformed mcp page should report the parse error")
	}
	if !strings.Contains(mcp, "broken") {
		t.Error("malformed mcp page should still show the raw source")
	}
}

// TestSkillBundledFiles checks that a skill page links its bundled Markdown doc
// (as a routed sub-page) and its non-Markdown asset (copied verbatim), and that
// the asset is actually present in the output.
func TestSkillBundledFiles(t *testing.T) {
	out := buildFixture(t)

	skill := read(t, filepath.Join(out, "skills/code-review/index.html"))
	// Links are base-relative (no leading slash) so they resolve under <base href>.
	if !strings.Contains(skill, `href="skills/code-review/checklist/"`) {
		t.Error("skill page should link its bundled doc sub-page (base-relative)")
	}
	if !strings.Contains(skill, `href="skills/code-review/lint.sh"`) {
		t.Error("skill page should link its bundled asset (base-relative)")
	}

	if _, err := os.Stat(filepath.Join(out, "skills/code-review/lint.sh")); err != nil {
		t.Errorf("bundled asset should be copied to output: %v", err)
	}

	// The doc sub-page back-links to its skill.
	doc := read(t, filepath.Join(out, "skills/code-review/checklist/index.html"))
	if !strings.Contains(doc, "skills/code-review/") {
		t.Error("skill doc should back-link to its skill")
	}
}
