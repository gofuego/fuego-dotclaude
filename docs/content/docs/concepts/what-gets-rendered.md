---
title: What Gets Rendered
layout: doc
summary: The coverage map — every artifact kind in a .claude directory and the page route it becomes.
nav_section: "Concepts"
nav_weight: 1
---

Coverage is exhaustive and recursive: everything in the `.claude` directory that means something to Claude Code becomes a page, including the internals of installed plugins.

## The coverage map

| In your `.claude` | Becomes |
|---|---|
| `agents/<name>.md` | `/agents/<name>/` |
| `skills/<name>/SKILL.md` | `/skills/<name>/` |
| `skills/<name>/<doc>.md` | `/skills/<name>/<doc>/` — a routed sub-page, linked from the skill |
| `skills/<name>/<asset>` (scripts, data) | copied verbatim, linked from the skill |
| `commands/<ns>/<name>.md` | `/commands/<ns>/<name>/` — shown as `/<ns>:<name>` |
| `output-styles/<name>.md` | `/output-styles/<name>/` |
| `CLAUDE.md`, `CLAUDE.local.md`, other `.md` | `/memory/<name>/` (unless promoted to home) |
| `.mcp.json` | `/mcp/` — one card per server |
| `settings.json` | `/settings/` |
| `settings.local.json` | `/settings/local/` |
| `plugins/<plugin>/plugin.json` | `/plugins/<plugin>/` |
| `plugins/<plugin>/agents/…`, `skills/…`, `commands/…` | first-class pages under their own sections, namespaced by plugin |
| `plugins/<plugin>/.mcp.json` | `/plugins/<plugin>/mcp/` |
| `marketplace.json` | `/marketplaces/<name>/` |

Intermediate directories that no page claims (e.g. `/agents/`, `/commands/git/`) get a generated **section page** listing their contents, so nothing in the sidebar is a dead end.

## The home page

The site's home (`/`) is your top-level `CLAUDE.md` — the **project-root sibling wins** over `.claude/CLAUDE.md` — followed by a generated catalog grouped by kind (Agents, Commands, Skills, Output Styles, Memory, MCP, Settings, Plugins, Marketplaces). If no `CLAUDE.md` exists anywhere, a generated dashboard takes its place. Memory files that aren't promoted to home keep their normal `/memory/` routes.

## Sibling files

When you build a project (rather than an isolated `.claude`), three root-level files are folded in from one directory above: `CLAUDE.md`, `CLAUDE.local.md`, and `.mcp.json`. They render as pages like everything else, labeled as coming from the project root. [Scope Detection](docs/concepts/scope-detection/) covers when this happens automatically.

## Plugins are rendered recursively

An installed plugin isn't a single page — its internal agents, skills, and commands are rendered with the **same layouts and routes** as your project's own artifacts, namespaced by plugin (`/agents/<plugin>/<name>/`-style slugs, `source` set to the plugin's name). The plugin page itself renders the manifest; a plugin's `.mcp.json` gets its own server-card page. This means the [taxonomies](docs/guides/browse-by-taxonomy/) and [cross-references](docs/concepts/cross-references/) treat plugin artifacts as first-class citizens.

## Tolerant by design — one bad file never fails the build

Claude Code frontmatter in the wild routinely has unquoted descriptions with colons that aren't valid YAML. An unparseable frontmatter block degrades to an empty envelope with the raw block preserved and a per-file warning — the page still renders. Malformed JSON degrades to a raw view of the file. Strictness is opt-in through the [validate command](docs/guides/validate-in-ci/).

## What's ignored

Version control and dependency directories never become content: `**/.git`, `**/node_modules`, `**/.DS_Store`, `.fuego`, `.github`, and `build` are ignored by default — which is what makes it safe to render a whole repo as the content root.
