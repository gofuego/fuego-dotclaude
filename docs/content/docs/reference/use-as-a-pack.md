---
title: Use as a Pack
layout: doc
summary: Consume dotclaude.Pack() from your own Fuego project — one line for the full rendering, with your config and theme winning over the pack's.
nav_section: "Reference"
nav_weight: 3
---

The pack is the product; the CLI is a zero-config wrapper over it. If you're building a [Fuego](https://gofuego.github.io/fuego-docs/) site of your own — or want `.claude` rendering as one section of a larger site — consume the pack directly:

```go
import (
    "github.com/gofuego/fuego-dotclaude/dotclaude"
    "github.com/gofuego/fuego/engine"
)

eng := engine.New()
eng.Use(dotclaude.Pack())
err := eng.Build(ctx, engine.BuildOptions{
    ContentDir: ".claude",
    OutputDir:  "build",
    SiteName:   "My Workspace",
})
```

`Pack()` bundles everything: the Markdown and JSON parsers, the classification and enrichment hooks, the cross-reference linker, the embedded theme with its static assets, and the route/taxonomy config defaults.

## Sibling injection is the CLI's job

Folding in a project root's `CLAUDE.md`/`CLAUDE.local.md`/`.mcp.json` is *not* part of the pack — it's an extra hook the CLI registers when [scope detection](docs/concepts/scope-detection/) calls for it. To get it programmatically:

```go
eng.AfterParse(dotclaude.SiblingHook(projectRoot))
```

## Routes the pack claims

The pack's config defaults route each artifact kind under its own section — `/agents/{slug}`, `/skills/{slug}`, `/commands/{slug}`, `/output-styles/{slug}`, `/memory/{slug}`, `/mcp`, `/settings`, `/settings/local`, `/plugins/{slug}` (+ `/plugins/{slug}/mcp`), `/marketplaces/{slug}` — plus the `/tools`, `/models`, and `/sources` taxonomies and the generated `/references/` graph. Composing into a larger site, make sure nothing else claims those prefixes (Fuego's collision detection will fail the build loudly if something does).

## Overriding

Standard Fuego pack precedence applies: your project's `config.yaml` deep-merges **over** the pack's defaults (change a route by restating it), and any file in your project's `theme/` directory wins over the pack's embedded theme — drop in your own `theme/layouts/agent.html` to restyle agent pages without forking anything.

## Version pinning

The pack pins a tagged Fuego release, so `go install`/`go get` work from the module proxy with no local checkouts. Requires Go 1.25+.
