---
title: Settings & MCP Pages
layout: doc
summary: How the JSON configuration files render — server cards for .mcp.json, curated sections for settings, and graceful degradation for malformed input.
nav_section: "Guides"
nav_weight: 2
---

The JSON files in a `.claude` directory get rich, purpose-built pages rather than syntax-highlighted dumps.

## `.mcp.json` → `/mcp/`

One **card per server**, showing the transport (`stdio` command, `sse`/`http` URL), arguments, and declared environment variable *names*. Server names are [cross-reference targets](docs/concepts/cross-references/), so an agent that mentions a server by name links here — and the server's card is where its backlinks accumulate. A collapsible raw-JSON block sits at the bottom for exact contents.

A plugin's own `.mcp.json` renders the same way at `/plugins/<name>/mcp/`.

## `settings.json` → `/settings/`

Claude Code's settings file has a long tail of keys, so the page is layered:

1. **Curated sections** for the high-signal groups — model, permissions (default mode, allow/deny rules), hooks, environment, status line, MCP server allow/deny, plugin controls.
2. **A generic key/value table** for everything else, so an unrecognized or newly introduced key is still visible.
3. **A collapsible raw JSON block** as ground truth.

`settings.local.json` renders identically at `/settings/local/`.

## Malformed input degrades, it doesn't fail

A JSON file that doesn't parse still gets a page — degraded to the raw view, with the problem reported. Unknown keys are preserved, not dropped. The same tolerance applies to Markdown frontmatter across the site: an invalid YAML block (unquoted colons are endemic in real agent files) renders the page with the raw frontmatter preserved and a warning banner, instead of failing the build.

To turn these soft warnings into a hard gate, use [`validate --strict`](docs/guides/validate-in-ci/).

## A note on secrets

These pages render what the files contain. `.mcp.json` commonly carries environment values and `settings.local.json` is machine-local by convention — review both before [publishing a built site](docs/guides/publish-your-workspace/) anywhere public.
