# PRD — fuego-dotclaude

A static-site generator that renders a `.claude/` folder as a navigable,
cross-referenced documentation site. Built on the [Fuego](https://github.com/gofuego/fuego)
meta-engine as a **format pack** (`dotclaude.Pack()`) plus a thin zero-config
CLI (`fuego-dotclaude`), following the same architecture as
[fuego-adr](https://github.com/gofuego/fuego-adr) — no engine fork.

> Status: design complete (grilled 2026-06-29), pre-implementation.

---

## Problem Statement

A Claude Code setup is a sprawl of heterogeneous artifacts — `CLAUDE.md` memory
files, subagents, skills, slash commands, output styles, `settings.json`,
`.mcp.json`, hook scripts, and whole plugins that bundle their own copies of all
of the above. They live across `~/.claude/` (user) and a project's `.claude/`,
in different formats (Markdown-with-frontmatter, JSON, shell scripts), and they
reference each other implicitly: an agent is invoked by `name`, a command by
`/slug`, an MCP tool by `mcp__server__tool`. Today the only way to understand
"what's in my `.claude/`, and how does it all relate?" is to open files one by
one in an editor. There is no map.

This hurts two audiences:

- The **maintainer of a `~/.claude/` environment** who has accumulated dozens of
  agents/skills/commands and can no longer see the whole, spot duplicates, or
  trace which artifacts mention which.
- The **author of a dedicated `.claude` repository** (a repo meant to be cloned
  into `~/.claude`, or a project's checked-in `.claude/`) who wants to publish a
  browsable, shareable reference for that configuration.

## Solution

`fuego-dotclaude` turns any `.claude/` directory into a static website with one
command. Point it at a folder (default `~/.claude`) and it discovers every
Claude Code artifact, renders each as a purpose-built page, uses the top-level
`CLAUDE.md` as the home page (with an auto-generated catalog of everything
present), and — the signature feature — **builds a navigable reference graph**:
every mention of an artifact's name anywhere in the site becomes a link to that
artifact's page, and every page shows what references it.

From the user's perspective:

- `fuego-dotclaude` with no arguments renders `~/.claude` into a site.
- `fuego-dotclaude .` inside a project renders that project's `.claude/`, and
  automatically folds in the root-level `CLAUDE.md` and `.mcp.json` that live
  beside it.
- `fuego-dotclaude serve` gives a live-reloading dev preview.
- `fuego-dotclaude validate` reports incoherent configuration (missing required
  frontmatter, dangling references, malformed JSON) and can fail CI with
  `--strict`.

The output is a complete, self-contained site (bespoke embedded theme; zero
files required in the consumer's project) suitable for local browsing or
deployment via the standard gofuego site-deploy workflows.

## User Stories

### Discovery & home

1. As a `.claude` maintainer, I want to run one command with no configuration and
   get a browsable site of my `~/.claude`, so that I can see my whole setup at a
   glance.
2. As a project developer, I want to point the tool at my repo root and have it
   find the project `.claude/` automatically, so that I don't have to know the
   exact content path.
3. As a project developer, I want the root-level `CLAUDE.md` and `.mcp.json` that
   sit beside `.claude/` to be included automatically, so that the site reflects
   the artifacts that straddle the repo root.
4. As an author of a dedicated `.claude` repo, I want to render it in isolation
   (no sibling lookup), so that the site shows exactly the environment the repo
   provides when cloned into `~/.claude`.
5. As a reader, I want the top-level `CLAUDE.md` to be the home page, so that the
   site opens with the human-written overview rather than a generated index.
6. As a reader, I want the home page to also show an auto-generated catalog of
   all artifacts grouped by type with counts, so that I can jump to any artifact
   from the landing page.
7. As a reader of a `.claude` with no `CLAUDE.md`, I want a generated dashboard
   home instead, so that the site still has a coherent entry point.
8. As a maintainer, I want nested and `*.local.md` memory files to appear as
   their own pages, so that no memory content is hidden.

### Per-artifact pages

9. As a reader, I want each subagent rendered as a page showing its name,
   description, allowed tools, pinned model, and system prompt, so that I
   understand what it does and how it's constrained.
10. As a reader, I want each skill rendered as a page showing its frontmatter and
    body, so that I understand when and how it triggers.
11. As a reader, I want a skill's bundled files (scripts, reference docs) listed
    and linked from the skill page, so that I can inspect the whole skill, not
    just `SKILL.md`.
12. As a reader, I want bundled Markdown files rendered as sub-pages and other
    bundled files available as raw/downloadable assets, so that every file is
    reachable.
13. As a reader, I want each slash command rendered as a page, with its namespace
    (from subdirectories) reflected in its name, so that `/ns:command` is shown
    correctly.
14. As a reader, I want output styles rendered as pages, so that they're part of
    the catalog.
15. As a reader, I want `CLAUDE.md` memory files rendered with their Markdown
    formatting intact, so that they're readable.

### JSON config pages

16. As a reader, I want `.mcp.json` rendered as a page with one card per MCP
    server (transport, command/url, env), so that I can see the configured
    servers without reading JSON.
17. As a reader, I want each MCP server name to be a linkable target, so that
    `mcp__server__tool` mentions elsewhere link to it.
18. As a reader, I want `settings.json` rendered as structured sections —
    permissions (allow/deny/ask), hooks, env, model, statusLine, MCP allow/deny,
    plugin/marketplace controls — so that I can read the important configuration
    at a glance.
19. As a reader, I want the long tail of `settings.json` keys shown in a generic
    key/value table, so that nothing is dropped even if it isn't bespoke-rendered.
20. As a reader, I want a collapsible raw-JSON block on every JSON page, so that I
    can always see the exact source.
21. As a reader, I want `settings.local.json` rendered the same way as
    `settings.json`, so that local overrides are visible.

### Cross-reference graph (signature)

22. As a reader, I want every whole-word mention of an artifact's name in prose
    and inline code to become a link to that artifact's page, so that I can
    navigate related content by following names.
23. As a reader, I want mentions inside fenced code blocks left un-linked, so that
    example code stays readable.
24. As a reader, I want a page to never link its own name, so that self-references
    aren't noise.
25. As a reader, I want text that is already a link left untouched, so that
    authored links win.
26. As a reader, I want name matching to be case-insensitive, so that "Docs-Writer"
    and "docs-writer" both link.
27. As a reader, I want common English words excluded via a stopword guard, so
    that short/common artifact names don't over-link.
28. As a reader, I want each artifact page to show a "Referenced by" section
    listing everything that mentions it, so that I can navigate the graph in
    reverse.
29. As a reader, I want a global reference-graph index page showing who links
    whom, so that I can see the structure of the whole configuration.
30. As a reader, when a bare name matches several artifacts (a collision across
    types, scopes, or plugins), I want it to link to a disambiguation page listing
    all artifacts that share the name, so that no link silently points at the
    wrong thing.
31. As a maintainer, I want a build warning whenever a name is ambiguous, so that
    I can rename to remove the collision.

### Taxonomies / faceted navigation

32. As a reader, I want to browse artifacts by the tools they declare
    (`tools` / `allowed-tools`), so that I can answer "what can touch the
    filesystem / run bash?".
33. As a reader, I want to browse artifacts by the model they pin, so that I can
    see which use opus/sonnet/haiku/inherit.
34. As a reader, I want to browse artifacts by source/provenance (project/user
    core vs each plugin), so that I can see where each artifact came from and
    review a single plugin's contribution.

### Plugins (recursive)

35. As a reader, I want each installed plugin's internal agents, commands, skills,
    hooks, and MCP config rendered as first-class pages, so that plugin contents
    are as browsable as top-level artifacts.
36. As a reader, I want plugin-provided artifacts namespaced by plugin, so that
    their names don't silently collide with my own.
37. As a reader, I want plugin marketplace metadata rendered, so that I can see
    where plugins were sourced from.

### Validation

38. As a maintainer, I want a warning when an agent/skill/command is missing
    required frontmatter (name/description), so that I can fix incomplete artifacts.
39. As a maintainer, I want a warning when a skill's `name` doesn't match its
    directory, so that invocation surprises are caught.
40. As a maintainer, I want a warning for dangling explicit references (a
    `subagent_type`, hook script, or MCP server that doesn't resolve), so that I
    catch broken wiring.
41. As a maintainer, I want a warning for malformed JSON, so that a bad config
    file is reported rather than silently skipped.
42. As a maintainer, I want a `validate` command that exits non-zero under
    `--strict`, so that I can gate CI on a coherent `.claude`.
43. As a maintainer, I want the site to still render despite warnings, so that
    validation never blocks visualization.

### CLI / dev experience

44. As a user, I want `build` (default), `serve`, `validate`, and `list`
    subcommands, so that the tool matches the familiar fuego-adr shape.
45. As a user, I want `serve` to live-reload when I edit files inside the
    `.claude` directory, so that I can iterate on my configuration and see changes.
46. As a user, I want `--siblings` / `--no-siblings` to override the automatic
    sibling detection, so that I control whether root-level files are folded in.
47. As a user, I want `list` to print all discovered artifacts, so that I can
    script against or quickly audit the inventory.
48. As an integrator, I want to consume `dotclaude.Pack()` in my own Fuego site
    with one line, so that a `.claude` view can be one section of a larger site.

## Implementation Decisions

### Architecture

- **Format pack + thin CLI**, identical in spirit to fuego-adr: all domain logic
  lives in the importable `dotclaude` package assembled into a single
  `core.Pack`; the CLI calls `eng.Use(dotclaude.Pack())` and the programmatic
  `engine.Build/Serve/Validate` API. No fork or modification of the engine; any
  gap is fixed in Fuego's pack API, not worked around here.
- **MIT-licensed**, new public gofuego repo `fuego-dotclaude` (open adoption
  funnel). Pack package `dotclaude`, binary `fuego-dotclaude`.
- **Bespoke embedded theme** (`//go:embed theme`, compiled Tailwind in
  `static/`), tuned to artifact cards, the reference graph, and JSON/permission
  tables. Renders a complete site with zero files in the consumer project.

### Dispatch & classification (constrained by the engine)

- Fuego dispatches parsers by file **extension** or **basename glob** only, and
  `Parser.Parse(raw []byte)` receives no path. Therefore `.md` artifacts
  (agents, commands, skills, output styles, memory) are **indistinguishable at
  parse time** and must be classified after the fact.
- A single **generic Markdown parser** handles all `.md` (frontmatter split +
  goldmark body, sections as `Raw: true` HTML nodes). An **AfterParse hook
  classifies each page by its `RelPath`** into an artifact kind and sets
  type/layout — the fuego-adr enrich pattern, extended to a path-convention
  table that also understands plugin-nested paths.
- JSON artifacts have **distinctive basenames**, so they use separate
  `FilenameParser`s: `.mcp.json`, `settings.json`, `settings.local.json` — each
  classified correctly at dispatch with no hook gymnastics.

### Content scope & siblings

- The content directory is **always a `.claude` directory**, default `~/.claude`.
  The CLI resolves the path with **smart detection**: no arg → `~/.claude`
  isolated; arg that *is* a `.claude` → isolated; arg that *contains* `.claude/`
  → use `<arg>/.claude` and auto-enable siblings; `--siblings`/`--no-siblings`
  override.
- **Root-level siblings** (`CLAUDE.md`, `CLAUDE.local.md`, `.mcp.json` one level
  above the content dir) are **injected via a hook** that reads them off disk and
  appends them as pages. The content dir stays the real `.claude` so `serve`
  live-reload works for everything inside it (sibling edits need a manual
  rebuild — accepted, rare).
- User-scope MCP config (`~/.claude.json`) is **out of scope** by default: it is
  secret-bearing and lives outside `~/.claude`. The rich MCP page targets project
  `.mcp.json`.

### Home & catalog

- Home = the top `CLAUDE.md` (root sibling wins over `.claude/CLAUDE.md` when
  both exist; the loser becomes a normal page), rendered, followed by an
  auto-generated catalog (artifacts by type with counts). If no `CLAUDE.md`
  exists, home is a pure generated dashboard.

### Cross-reference graph

- **Slug targets:** all named artifact types — agents/skills/output-styles by
  frontmatter `name` (fallback filename stem), commands by namespaced path, MCP
  servers by their `mcpServers` key.
- **Linker:** matches **all whole-word occurrences**, case-insensitive, in prose
  and inline code; **skips fenced code blocks**, text already inside a link, and
  the page's own slug; guarded by a stopword list. Implemented by walking the
  rendered HTML's text nodes (not naive regex replacement) to avoid corrupting
  tags, attributes, or existing links.
- **Collisions** resolve to a generated **disambiguation page** (name → all
  sharing artifacts) and emit a build warning.
- **Reverse navigation:** per-page "Referenced by" sections plus a global
  reference-graph index, generated in an Index hook from the same scan.

### Taxonomies

- The AfterParse hook normalizes per-type frontmatter into unified envelope keys
  (`tools` from `tools`/`allowed-tools`; `model`; `source`/provenance). Three
  taxonomies are declared in the pack's config defaults: **tool**, **model**,
  **source**. Provenance is especially load-bearing once plugin internals are
  rendered.

### Plugins

- The classifier recurses into `plugins/<plugin>/…`, rendering each plugin's
  internal agents/commands/skills/hooks/MCP as first-class, cross-referenced
  pages; plugin artifacts are namespaced by plugin and tagged with their source.
  Marketplace metadata is rendered. This is the largest scope multiplier and is
  sequenced last.

### Candidate deep modules (testable in isolation, stable interfaces)

These are the modules to extract and test independently — each hides a lot of
behavior behind a small surface:

1. **Classifier** — `RelPath → ArtifactKind` (incl. plugin nesting, command
   namespacing). Pure function; exhaustively table-testable.
2. **Slug linker** — `(html, registry, selfSlug) → html` with all matching rules
   (whole-word, case-insensitive, fenced/existing-link/self skips, stopwords).
   The hardest correctness surface; pure and table-testable.
3. **Reference-graph builder** — `linked pages → {backlinks, graph, disambiguation
   sets}`. Pure over an in-memory page set.
4. **JSON config models** — `raw JSON → structured MCP/settings model` for rich
   rendering. Pure decode; tolerant of unknown keys.
5. **Scope resolver** — `CLI path + flags → {contentDir, siblings, mode}`. Pure
   over a filesystem abstraction.
6. **Validator** — `artifact set → []diagnostic`. Pure over the page set.

## Testing Decisions

- **Test external behavior, not implementation details.** Tests assert on
  rendered output and module return values, not private helpers.
- **Vanilla-pack contract test** (the spine, modeled on fuego-adr's
  `pack_test.go`): build a fixture `.claude` tree with **only**
  `eng.Use(dotclaude.Pack())` and assert that every artifact type's page, the
  home/catalog, the cross-reference links and backlinks, the disambiguation and
  graph pages, the taxonomy pages, and the pack's static assets all render. Keep
  it green — it is the regression spine.
- **Deep-module unit tests** for the six modules above, especially the **slug
  linker** (extensive table tests: prose vs inline vs fenced, collisions,
  self-skip, stopwords, existing links, case) and the **classifier** (every path
  convention incl. plugins).
- **Golden-file integration tests** following the engine's `testdata/input` +
  `testdata/golden` pattern, with a fixture covering every artifact type and at
  least one plugin; regenerated with the engine's `-update` flow.
- **Validation tests** asserting each diagnostic fires on a crafted-bad fixture
  and that `--strict` changes exit code while the site still renders.
- **Prior art:** fuego-adr's `adr/pack_test.go`, `adr/parser_test.go`, and the
  engine's `integration_test.go` / `incremental_test.go` golden-file suites.

## Out of Scope

- **Editing** any `.claude` artifact through the site — this is a read-only
  visualizer (no fuego-studio integration in v1).
- **Merging multiple scopes** into one site (user + project layered, or
  multi-root sectioned). v1 renders exactly one `.claude` directory plus its
  immediate siblings.
- **Rendering `~/.claude.json`** (user-scope MCP / OAuth / caches) — secret-bearing
  and outside `~/.claude`.
- **Session transcripts, plans, backups, and other runtime/state files** under
  `~/.claude` (`sessions/`, `plans/`, timestamped backups) — operational data,
  not configuration artifacts.
- **Executing** hooks, skills, agents, or MCP servers — the site documents them,
  it does not run them.
- **Live runtime data** (which tools an MCP server actually exposes, whether a
  hook is currently firing) — only declared configuration is rendered.

## Further Notes

- The authoritative Claude Code artifact schema/locations are at
  `code.claude.com/docs` (index: `code.claude.com/docs/llms.txt`). Exact
  frontmatter fields should be re-verified against the docs at implementation
  time; `settings.json` has ~80 top-level keys, which is why settings rendering
  is "curated sections + generic table + raw" rather than fully bespoke.
- Implementation is sequenced as tracer-bullet vertical slices: (1) skeleton +
  vanilla-pack test, (2) Markdown artifacts + classifier, (3) home + catalog,
  (4) slug linker, (5) backlinks + disambiguation + graph, (6) JSON pages,
  (7) sibling injection + CLI smart detection, (8) taxonomies, (9) plugins
  (recursive), (10) validation + CLI polish + CI. Slices 1–8 deliver a complete
  non-plugin site before the plugin-recursion multiplier lands.
- `go.mod` will carry a `replace` directive to a local Fuego checkout while
  consuming any unreleased engine features; pin a tagged release and remove the
  replace before publishing.
- CI reuses the gofuego `.github` `go-ci.yml` reusable workflow.
