# CLAUDE.md — fuego-dotclaude Contributor Guide

## What is fuego-dotclaude?

fuego-dotclaude is a **domain-specific static site generator for `.claude/`
directories**, built on the [Fuego](https://github.com/gofuego/fuego)
meta-engine. It turns the artifacts of a Claude Code workspace — agents, skills,
slash commands, output styles, memory, `.mcp.json`/`settings.json`, and installed
plugins — into a navigable site with whole-word cross-links, backlinks,
disambiguation, faceted taxonomies, and a catalog.

It is structured as a **Fuego format pack** (`dotclaude.Pack()`) plus a thin CLI.
The pack is the product; the CLI is a zero-config convenience wrapper.

## How it uses Fuego

fuego-dotclaude does not fork Fuego — everything works through public extension
points:

| Fuego extension point | fuego-dotclaude usage |
|---|---|
| `core.Parser` (extension) | `MarkdownParser` handles every `.md` artifact |
| `core.FilenameParser` | `.mcp.json`, `settings.json`, `settings.local.json`, `plugin.json`, `marketplace.json` |
| `core.Pack` (`eng.Use`) | `dotclaude.Pack()` bundles parsers + theme + config + hooks |
| `Pack.ConfigDefaults` | routes + the tools/model/source taxonomies |
| `core.AfterParseHook` | classify by `RelPath`; namespace plugin artifacts; normalize taxonomy fields |
| `core.IndexHook` | skill bundled files, plugin components, home + catalog, references |
| `core.BeforeRenderHook` | the slug cross-reference linker |
| `engine.AfterParse` (CLI) | sibling injection, registered by the CLI when applicable |

## The load-bearing constraint

Fuego dispatches parsers by **file extension or basename glob only**, and
`Parser.Parse(raw []byte)` receives **no path**. So every Markdown artifact
(agents, skills, commands, output styles, memory) is indistinguishable at parse
time. The fix: one generic `.md` parser, and **classification happens in an
AfterParse hook from each page's `RelPath`** (package `classify`). JSON artifacts
have distinctive basenames, so they get their own `FilenameParser`s.

## Architecture Decisions

### AD-1: A pack plus a thin CLI

All logic lives in the importable `dotclaude` package, assembled into one
`core.Pack`. The CLI calls `eng.Use(dotclaude.Pack())` and the programmatic build
API. A vanilla Fuego project needs one line.

### AD-2: Classification is a deep module over `RelPath`

`classify` is a pure function from a content-dir-relative path to an artifact
kind plus routing identity (slug, command namespace, skill root). It also strips
a `plugins/<name>/` prefix and classifies the inner artifact, so plugin
agents/skills/commands reuse the same kinds and layouts as project artifacts.

### AD-3: Classify in AfterParse, route by the assigned type

The AfterParse hook stamps `Type`, `Layout`, and a path-derived `slug` before
ROUTE, so routing and layout follow the artifact kind. Plugin artifacts are
namespaced (`slug` = `<plugin>/<slug>`, `source` = plugin name). Injected
siblings are flagged so the classifier skips them.

### AD-4: The slug linker walks the HTML token stream

The linker (package `sluglink`, run in BEFORE-RENDER) rewrites rendered HTML by
walking tokens (`golang.org/x/net/html`), not by regex over markup. It links in
prose and inline `<code>`, skips fenced `<pre>` and existing `<a>`, is whole-word
and case-insensitive, stopword-guarded, and skips the page's own name. Ambiguous
names link to a disambiguation route.

### AD-5: Reverse navigation is built in INDEX

`refgraph` computes backlinks and the edge list from forward links. The INDEX
reference hook scans bodies (without rewriting), attaches "Referenced by" to each
page, and appends collision-checked virtual pages (disambiguation pages, the
`/references/` graph). The body rewrite still happens in BEFORE-RENDER.

### AD-6: Parsing is tolerant — one bad file never fails the build

`jsonmodel` decodes `.mcp.json`, settings, and plugin/marketplace manifests,
preserving unknown keys (settings' long tail) and reporting malformed input so
pages degrade to a raw view. Markdown frontmatter is just as forgiving: Claude
Code agent frontmatter routinely has unquoted descriptions with colons that
aren't valid YAML, so an unparseable block degrades to an empty envelope with the
raw block preserved (`frontmatter_raw`) and a per-file warning, rather than
failing the build. Strictness is opt-in via the `validate` command.

### AD-7: Scope resolution is a deep module

`scope` maps a CLI path + flags to `{contentDir, siblingDir, siblings, mode}`
over a filesystem abstraction, so the two usage cases (isolated `.claude` vs a
project repo with siblings) are decided in one testable place.

### AD-8: Validation is advisory

`validate` emits diagnostics (missing frontmatter, skill name vs directory,
dangling references, malformed JSON) over an engine-independent artifact view.
The site always renders; `validate --strict` only changes the exit code.

## Project structure

```
fuego-dotclaude/
  main.go                  CLI entry point
  dotclaude/               the pack (importable)
    dotclaude.go           Pack() — parsers + theme + config + hooks
    parser.go              generic .md parser
    json.go                .mcp.json + settings parsers
    plugins.go             plugin.json + marketplace.json parsers, plugin hook
    hooks.go               AfterParse classify/enrich; skill bundled-file index hook
    home.go                home + catalog index hook
    linker.go              the BEFORE-RENDER slug linker hook
    refhooks.go            backlinks + disambiguation + graph index hook
    siblings.go            sibling injection hook
    inspect.go             capture + diagnostics + listing bridges
    config-defaults.yaml   routes + taxonomies
    classify/              deep module: RelPath -> kind + identity
    jsonmodel/             deep module: tolerant JSON decode
    sluglink/              deep module: HTML-aware cross-link engine
    refgraph/              deep module: forward links -> backlinks/edges
    scope/                 deep module: CLI path + flags -> what to build
    validate/              deep module: coherence diagnostics
    theme/                 embedded base + layouts + renderers + static
  internal/
    cli/                   build, serve, validate, list
    config/                optional fuego-dotclaude.yaml
  testdata/sample/.claude  a representative fixture for manual builds
```

## Testing

- `go test ./... -race` — each deep module is unit-tested in isolation, and
  `dotclaude/pack_test.go` is the **vanilla-pack contract**: it builds a
  representative `.claude` with only `eng.Use(dotclaude.Pack())` and asserts every
  artifact kind, the cross-links, taxonomies, plugins, and reverse navigation.
- `go run . build testdata/sample/.claude -o /tmp/out` for a manual build.

## Dependency note

Fuego is pinned to a tagged release (`github.com/gofuego/fuego v0.4.4`) and
resolved from the module proxy, so CI (`gofuego/.github` `go-ci.yml`) and
`go install` work without the workspace checkout. To develop against an
unreleased Fuego, add a temporary `replace github.com/gofuego/fuego => ../fuego`
and remove it before committing. (v0.4.4 predates Fuego's LICENSE; bump the pin
to the first licensed tag when one is cut.)
