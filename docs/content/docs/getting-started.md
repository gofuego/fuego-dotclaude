---
title: Getting Started
layout: doc
summary: Install the CLI, point it at a project or a .claude directory, and browse the generated site.
nav_section: "Getting Started"
nav_weight: 1
---

## Install

```bash
go install github.com/gofuego/fuego-dotclaude@latest
```

Requires Go 1.25+. The binary is self-contained — the theme is embedded, and nothing is ever written into the directory you point it at.

## Build your first site

The common case is a project repo that contains a `.claude/` directory. `cd` in and run:

```bash
cd my-project
fuego-dotclaude serve
```

That builds `./.claude` — folding in the root-level `CLAUDE.md`, `CLAUDE.local.md`, and `.mcp.json` — and serves it with live reload, rebuilding as you edit. Output and cache go to a scratch directory outside your project, so watching never loops and no artifacts land in your repo.

For a one-shot static build:

```bash
fuego-dotclaude build            # writes to ./build
fuego-dotclaude build -o /tmp/site
```

You can also point it at things explicitly:

```bash
fuego-dotclaude build path/to/project   # a repo containing .claude/
fuego-dotclaude build ~/.claude         # a .claude directly — isolated
fuego-dotclaude build path/to/ai        # a dedicated repo that IS the .claude
                                        # layout under another name — isolated
```

The detection rules are small and predictable — see [Scope Detection](docs/concepts/scope-detection/).

## What you get

- **Home** — your top-level `CLAUDE.md` (the project-root one wins over `.claude/CLAUDE.md`), followed by a generated catalog of every artifact. With no `CLAUDE.md` anywhere, a generated dashboard.
- **A page per artifact** — agents, skills (with their bundled files), slash commands (namespaced, `/git:commit`), output styles, memory files, `.mcp.json`, `settings.json`, and each installed plugin's internals. The full map is in [What Gets Rendered](docs/concepts/what-gets-rendered/).
- **Cross-references** — whole-word mentions of an artifact's name auto-link to its page, each page lists what references it, and `/references/` is the global graph. See [Cross-References](docs/concepts/cross-references/).
- **Taxonomy hubs** — browse by `tools`, `model`, and `source` (project vs. each plugin). See [Browse by Taxonomy](docs/guides/browse-by-taxonomy/).
- **An IDE-style sidebar** — a collapsible file tree of the whole artifact set, identical on every page.

## Check your workspace without building

```bash
fuego-dotclaude list             # every discovered artifact: TYPE, NAME, ROUTE
fuego-dotclaude validate         # advisory coherence diagnostics
```

`validate` reports missing frontmatter, name mismatches, dangling references, and malformed JSON — warnings by default, `--strict` for CI. See [Validate in CI](docs/guides/validate-in-ci/).

## Next steps

- Publishing the site somewhere? Read [Publish Your Workspace](docs/guides/publish-your-workspace/) — including what **not** to publish.
- Building a larger Fuego site and want `.claude` rendering as one section? See [Use as a Pack](docs/reference/use-as-a-pack/).
