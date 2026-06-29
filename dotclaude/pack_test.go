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
See the bundled checklist. Pairs well with the code-reviewer agent.
`)
	writeFile(t, content, "skills/code-review/checklist.md", `# Checklist
- Correctness first.
`)
	writeFile(t, content, "skills/code-review/lint.sh", "#!/usr/bin/env bash\necho lint\n")
	writeFile(t, content, "commands/git/commit.md", `---
description: Commit staged changes.
allowed-tools: Bash, Read
---
Create a conventional commit.
`)
	writeFile(t, content, "output-styles/concise.md", `---
name: Concise
description: Terse responses.
---
Lead with the answer.
`)
	// An ambiguous name "deploy": an agent and a command both claim it.
	writeFile(t, content, "agents/deployer.md", "---\nname: deploy\ndescription: Deploys releases.\n---\nDeploy it.\n")
	writeFile(t, content, "commands/deploy.md", "---\ndescription: Deploy the app.\n---\nDeploy it.\n")
	writeFile(t, content, "CLAUDE.md", "# Workspace memory\nProject conventions live here.\n")
	writeFile(t, content, "CLAUDE.local.md", "# Local memory\nMachine-specific notes.\n")
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

	// A plugin with its own manifest, agent, skill, and MCP config.
	writeFile(t, content, "plugins/acme/plugin.json", `{
  "name": "acme-tools",
  "version": "1.0.0",
  "description": "Acme developer tools.",
  "author": { "name": "Acme" }
}`)
	writeFile(t, content, "plugins/acme/agents/linter.md", "---\nname: linter\ndescription: Lints code.\n---\nLint everything.\n")
	writeFile(t, content, "plugins/acme/skills/format/SKILL.md", "---\nname: format\n---\nFormat code.\n")
	writeFile(t, content, "plugins/acme/.mcp.json", `{ "mcpServers": { "acme-api": { "url": "https://acme.dev/mcp" } } }`)

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
		{"memory/CLAUDE.local/index.html", []string{"Local memory"}},
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

// TestSiblingInjection builds a project-style layout (a repo containing .claude/
// plus root-level CLAUDE.md and .mcp.json) and verifies the siblings are folded
// in: the root CLAUDE.md becomes the home page and the root .mcp.json renders.
func TestSiblingInjection(t *testing.T) {
	repo := t.TempDir()
	content := filepath.Join(repo, ".claude")
	out := filepath.Join(repo, "out")

	writeFile(t, content, "agents/reviewer.md", "---\nname: reviewer\n---\nReviews.\n")
	// Root-level siblings, one level above the content dir.
	writeFile(t, repo, "CLAUDE.md", "# Project root memory\nTop-level conventions.\n")
	writeFile(t, repo, ".mcp.json", `{ "mcpServers": { "github": { "command": "npx" } } }`)

	eng := engine.New()
	eng.Use(dotclaude.Pack())
	eng.AfterParse(dotclaude.SiblingHook(repo)) // siblingDir = repo (parent of .claude)
	if err := eng.Build(context.Background(), engine.BuildOptions{ContentDir: content, OutputDir: out, SiteName: "Repo"}); err != nil {
		t.Fatalf("build with siblings failed: %v", err)
	}

	// The root CLAUDE.md is the home page.
	home := read(t, filepath.Join(out, "index.html"))
	if !strings.Contains(home, "Project root memory") {
		t.Error("root CLAUDE.md should be the home page")
	}
	// The root .mcp.json renders.
	mcp := read(t, filepath.Join(out, "mcp/index.html"))
	if !strings.Contains(mcp, "github") {
		t.Error("root .mcp.json should render its server")
	}
	// The in-.claude agent still renders and is cataloged.
	if !strings.Contains(home, `href="agents/reviewer/"`) {
		t.Error("home catalog should list the in-.claude agent")
	}
}

// TestValidationIsAdvisory verifies the site still renders despite coherence
// problems, and that Diagnostics reports them (the data behind `validate`).
func TestValidationIsAdvisory(t *testing.T) {
	dir := t.TempDir()
	content := filepath.Join(dir, ".claude")
	out := filepath.Join(dir, "out")
	writeFile(t, content, "agents/bad.md", "---\n---\nNo name, no description.\n")

	cap := dotclaude.NewCapture()
	eng := engine.New()
	eng.Use(dotclaude.Pack())
	eng.Index(cap.Index())
	if err := eng.Build(context.Background(), engine.BuildOptions{ContentDir: content, OutputDir: out, SiteName: "X"}); err != nil {
		t.Fatalf("build should succeed despite problems: %v", err)
	}
	// The site still renders.
	if _, err := os.Stat(filepath.Join(out, "agents/bad/index.html")); err != nil {
		t.Errorf("site should render despite problems: %v", err)
	}
	// Diagnostics flag the missing frontmatter.
	var found bool
	for _, d := range dotclaude.Diagnostics(cap.Pages()) {
		if strings.Contains(d.Message, "name") {
			found = true
		}
	}
	if !found {
		t.Error("expected a missing-name diagnostic for agents/bad.md")
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

// TestHomeIsClaudeMdPlusCatalog checks that the top-level CLAUDE.md is rendered
// at "/" followed by a catalog grouping artifacts by type with counts, and that
// the re-homed CLAUDE.md no longer renders at its memory route.
func TestHomeIsClaudeMdPlusCatalog(t *testing.T) {
	out := buildFixture(t)

	home := read(t, filepath.Join(out, "index.html"))
	for _, want := range []string{
		"Workspace memory",            // the CLAUDE.md body
		"Catalog",                     // generated catalog
		"Agents",                      // a section
		`href="agents/code-reviewer/"`, // a catalog link
	} {
		if !strings.Contains(home, want) {
			t.Errorf("home page missing %q", want)
		}
	}

	// The CLAUDE.md used as home is not also emitted at /memory/CLAUDE/.
	if _, err := os.Stat(filepath.Join(out, "memory/CLAUDE/index.html")); err == nil {
		t.Error("home CLAUDE.md should not also render at /memory/CLAUDE/")
	}
}

// TestHomeDashboardWithoutClaudeMd checks that, with no CLAUDE.md anywhere, "/"
// is a generated dashboard carrying the catalog.
func TestHomeDashboardWithoutClaudeMd(t *testing.T) {
	dir := t.TempDir()
	content := filepath.Join(dir, ".claude")
	out := filepath.Join(dir, "out")
	writeFile(t, content, "agents/reviewer.md", "---\nname: reviewer\n---\nReviews.\n")

	eng := engine.New()
	eng.Use(dotclaude.Pack())
	if err := eng.Build(context.Background(), engine.BuildOptions{ContentDir: content, OutputDir: out, SiteName: "X"}); err != nil {
		t.Fatal(err)
	}

	home := read(t, filepath.Join(out, "index.html"))
	if !strings.Contains(home, "Catalog") || !strings.Contains(home, `href="agents/reviewer/"`) {
		t.Errorf("dashboard should show the catalog with the agent; got:\n%s", home)
	}
}

// TestTaxonomies checks that tool, model, and source taxonomy term + index
// pages render, and that an artifact declaring multiple tools appears under each
// tool term.
func TestTaxonomies(t *testing.T) {
	out := buildFixture(t)

	// Index pages render for each taxonomy.
	for _, idx := range []string{"tools/index.html", "models/index.html", "sources/index.html"} {
		if _, err := os.Stat(filepath.Join(out, idx)); err != nil {
			t.Errorf("expected taxonomy index %s: %v", idx, err)
		}
	}

	// code-reviewer declares Read, Grep, Bash -> appears under each tool term.
	for _, term := range []string{"read", "grep", "bash"} {
		page := filepath.Join(out, "tools", term, "index.html")
		html, err := os.ReadFile(page)
		if err != nil {
			t.Errorf("expected tool term page %s: %v", term, err)
			continue
		}
		if !strings.Contains(string(html), "code-reviewer") {
			t.Errorf("tool term %q should list code-reviewer", term)
		}
	}

	// The Bash term also lists the git:commit command (allowed-tools -> tools).
	bash := read(t, filepath.Join(out, "tools/bash/index.html"))
	if !strings.Contains(bash, "/git:commit") {
		t.Error("Bash tool term should also list the git:commit command")
	}

	// Model term page for the agent's pinned model.
	model := read(t, filepath.Join(out, "models/sonnet/index.html"))
	if !strings.Contains(model, "code-reviewer") {
		t.Error("model term sonnet should list code-reviewer")
	}

	// Source term page groups project artifacts.
	if _, err := os.Stat(filepath.Join(out, "sources/project/index.html")); err != nil {
		t.Errorf("expected source term project: %v", err)
	}
}

// TestPluginRendersNamespaced checks that a plugin's internal artifacts render
// as first-class pages under a plugin-namespaced route, the plugin manifest page
// lists them, and the plugin's artifacts are grouped under a source term.
func TestPluginRendersNamespaced(t *testing.T) {
	out := buildFixture(t)

	// Internal artifacts render namespaced by plugin (no collision with core).
	for _, page := range []string{
		"agents/acme/linter/index.html",
		"skills/acme/format/index.html",
		"plugins/acme/mcp/index.html",
		"plugins/acme/index.html", // the manifest page
	} {
		if _, err := os.Stat(filepath.Join(out, page)); err != nil {
			t.Errorf("expected plugin page %s: %v", page, err)
		}
	}

	// The manifest page renders metadata and lists the plugin's components.
	manifest := read(t, filepath.Join(out, "plugins/acme/index.html"))
	for _, want := range []string{"acme-tools", "1.0.0", "Acme", "agents/acme/linter/"} {
		if !strings.Contains(manifest, want) {
			t.Errorf("plugin manifest page missing %q", want)
		}
	}

	// Source taxonomy groups the plugin's artifacts under its name.
	if _, err := os.Stat(filepath.Join(out, "sources/acme/index.html")); err != nil {
		t.Errorf("expected source term for plugin acme: %v", err)
	}
}

// TestCrossReferenceLinks checks that a mention of an artifact's name in another
// page's prose is rewritten into a link to that artifact's page.
func TestCrossReferenceLinks(t *testing.T) {
	out := buildFixture(t)

	// The code-review skill body mentions "code-reviewer", which should link to
	// the agent's page.
	skill := read(t, filepath.Join(out, "skills/code-review/index.html"))
	if !strings.Contains(skill, `href="agents/code-reviewer/"`) {
		t.Errorf("skill body should auto-link the code-reviewer agent; got:\n%s", skill)
	}
	if !strings.Contains(skill, "slug-link") {
		t.Error("auto-link should carry the slug-link class")
	}
}

// TestBacklinksDisambiguationGraph checks the reverse-navigation slice: a
// "Referenced by" list, a disambiguation page per ambiguous name, and a global
// reference-graph index.
func TestBacklinksDisambiguationGraph(t *testing.T) {
	out := buildFixture(t)

	// The code-review skill mentions code-reviewer, so the agent page lists it
	// under "Referenced by".
	agent := read(t, filepath.Join(out, "agents/code-reviewer/index.html"))
	if !strings.Contains(agent, "Referenced by") {
		t.Error("code-reviewer page should have a Referenced by section")
	}
	if !strings.Contains(agent, `href="skills/code-review/"`) {
		t.Error("code-reviewer should be backlinked from the code-review skill")
	}

	// "deploy" is claimed by an agent and a command -> a disambiguation page that
	// lists both.
	dis := read(t, filepath.Join(out, "disambiguation/deploy/index.html"))
	for _, want := range []string{`href="agents/deployer/"`, `href="commands/deploy/"`} {
		if !strings.Contains(dis, want) {
			t.Errorf("disambiguation page missing %q", want)
		}
	}

	// The global reference graph renders.
	graph := read(t, filepath.Join(out, "references/index.html"))
	if !strings.Contains(graph, "Reference graph") {
		t.Error("reference-graph index should render")
	}
	if !strings.Contains(graph, `href="agents/code-reviewer/"`) {
		t.Error("reference graph should show the skill -> code-reviewer edge")
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
