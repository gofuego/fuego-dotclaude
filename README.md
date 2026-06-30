# fuego-dotclaude

Render a `.claude/` directory as a navigable, cross-referenced documentation
site. fuego-dotclaude is a [Fuego](https://github.com/gofuego/fuego) format pack
(plus a thin CLI) that turns the agents, skills, commands, output styles, memory,
MCP/settings config, and installed plugins of a Claude Code workspace into a
static site — with whole-word cross-links between artifacts, backlinks, faceted
taxonomies, and a catalog.

## Install

```sh
go install github.com/gofuego/fuego-dotclaude@latest
```

## Usage

The CLI detects what you mean from the path (default: the current directory):

```sh
# In your project (the common case): cd in and run it.
cd my-project
fuego-dotclaude build            # builds ./.claude + folds in root CLAUDE.md/.mcp.json
fuego-dotclaude serve            # live-reloading preview

# Or point at something explicitly.
fuego-dotclaude build path/to/project   # a repo containing .claude/
fuego-dotclaude build ~/.claude         # a .claude directly -> isolated
fuego-dotclaude build path/to/ai        # a dedicated repo that IS the .claude,
                                        # under a different name -> isolated
```

Detection rules (default argument is the current directory):

- **a directory containing `.claude/`** → `<arg>/.claude`, with the root-level
  `CLAUDE.md`, `CLAUDE.local.md`, and `.mcp.json` folded in
- **a `.claude` directory** → that directory, in isolation
- **any other directory** → rendered as the content root, in isolation — for a
  dedicated repo that holds the `.claude` layout under a different name (e.g. an
  `ai/` repo)
- a path that isn't a directory → an error
- `--siblings` / `--no-siblings` override the sibling default

### Commands

| Command | What it does |
|---|---|
| `build [path]` | Build the static site (`-o`, `--base-url`, `--incremental`). |
| `serve [path]` | Build and serve with live reload (`--port`). |
| `validate [path]` | Report coherence problems; `--strict` exits non-zero for CI. |
| `list [path]` | Print every discovered artifact (type, name, route). |

## What it renders

- **Agents, skills (+ bundled files), slash commands (namespaced `/ns:command`),
  output styles, memory** (`CLAUDE.md` / `CLAUDE.local.md`).
- **`.mcp.json`** as a server-per-card page; **`settings.json` /
  `settings.local.json`** as curated sections plus the long-tail key/value table
  and a raw block.
- **Home** is your top-level `CLAUDE.md` followed by a generated catalog; with no
  `CLAUDE.md`, a generated dashboard.
- **Cross-references:** any whole-word mention of an artifact's name links to it;
  each page lists what references it; ambiguous names get a disambiguation page;
  `/references/` is the global graph.
- **Taxonomies:** group by `tools`, `model`, and `source` (project vs each
  plugin).
- **Plugins:** each installed plugin's internal agents/skills/commands/MCP render
  as first-class pages, namespaced by plugin, with the manifest and any
  marketplace metadata.

## Using the pack directly

The pack is the product; the CLI is a zero-config wrapper. Embed it in any Fuego
build with one line:

```go
eng := engine.New()
eng.Use(dotclaude.Pack())
eng.Build(ctx, engine.BuildOptions{ContentDir: ".claude", OutputDir: "build"})
```

To fold in root-level siblings, also register `dotclaude.SiblingHook(parentDir)`.

## Development

```sh
go build ./...
go test ./... -race
go run . build testdata/sample/.claude -o /tmp/out
```

Fuego is pinned to a tagged release (`github.com/gofuego/fuego v0.4.5`) and
resolved from the module proxy, so CI and `go install` work without the workspace
checkout. To develop against an unreleased Fuego, add a temporary
`replace github.com/gofuego/fuego => ../fuego` and drop it before committing.

## License

MIT — see [LICENSE](LICENSE).
