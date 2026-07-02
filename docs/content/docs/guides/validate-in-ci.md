---
title: Validate in CI
layout: doc
summary: The validate command's diagnostics, what each one means, and how to gate a repository on a clean .claude tree with --strict.
nav_section: "Guides"
nav_weight: 3
---

`fuego-dotclaude validate` runs advisory coherence checks over a `.claude` directory without building HTML. Diagnostics print one per line, attributed to the file to fix; the exit code stays zero unless you pass `--strict`.

```bash
fuego-dotclaude validate            # report problems, exit 0
fuego-dotclaude validate --strict   # exit non-zero if anything is found
```

## The diagnostics

| Diagnostic | Meaning |
|---|---|
| `missing required frontmatter: name` | An agent or skill has no `name:` — Claude Code can't address it. |
| `missing required frontmatter: description` | An agent, skill, or command has no `description:` — nothing tells Claude when to use it. |
| `skill name "x" does not match its directory "y"` | A skill's `name:` differs from its folder name; Claude Code resolves skills by directory. |
| `references unknown subagent "x"` | An artifact declares a `subagent_type` that matches no agent in the tree. |
| `references unknown MCP server "x"` | An artifact references an MCP server name that no `.mcp.json` defines. |
| `malformed JSON: …` | A JSON config file failed to parse (the site renders it as a raw view). |

Everything is **advisory by design**: the site always renders, because a half-written skill shouldn't block viewing the rest of the workspace. `--strict` only changes the exit code.

## In GitHub Actions

```yaml
name: Validate .claude
on:
  pull_request:
  push:
    branches-ignore: [main]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.25"
      - run: go run github.com/gofuego/fuego-dotclaude@latest validate --strict
```

Run from the repo root: [scope detection](docs/concepts/scope-detection/) finds `./.claude` and folds in the root `CLAUDE.md`/`.mcp.json`, so the siblings are validated too.

## Listing what was found

`fuego-dotclaude list` prints every discovered artifact as a `TYPE / NAME / ROUTE` table — useful for a quick inventory, or for diffing what a plugin install actually added.
