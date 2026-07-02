---
title: Scope Detection
layout: doc
summary: How the CLI decides what to build from the path you give it — the four resolution rules, project vs. isolated mode, and the sibling toggle.
nav_section: "Concepts"
nav_weight: 3
---

Every command (`build`, `serve`, `validate`, `list`) takes an optional path and resolves it the same way. No argument means the current directory — the common case is `cd`-ing into a project and running the tool there.

## The four rules

Checked in order:

| You give it | It builds | Mode |
|---|---|---|
| a directory **containing** `.claude/` | `<arg>/.claude` | **project** — siblings folded in |
| a directory **named** `.claude` | that directory | isolated |
| any **other directory** | that directory as the content root | isolated |
| a path that isn't a directory | — | error |

The third rule is for a dedicated repo that holds the `.claude` layout under a different name — an `ai/` repo with `agents/`, `skills/`, and `commands/` at its root renders exactly like a `.claude` would.

## Project mode and siblings

In project mode, three files from the directory *above* the `.claude` are folded in as pages: the root `CLAUDE.md`, `CLAUDE.local.md`, and `.mcp.json`. The root `CLAUDE.md` also takes precedence as the site's [home page](docs/concepts/what-gets-rendered/#the-home-page).

The content directory stays the real `.claude` — siblings are injected during the build, not copied — so `serve`'s file watching works inside the directory you're actually editing.

## Overriding the default

Two mutually exclusive flags force the sibling behavior regardless of what detection decided:

```bash
fuego-dotclaude build ~/.claude --siblings     # fold in ~/CLAUDE.md etc. anyway
fuego-dotclaude build my-project --no-siblings # project, but .claude only
```

## Examples

```bash
fuego-dotclaude build                   # cwd contains .claude/ → project mode
fuego-dotclaude build ~/code/app       # app contains .claude/ → project mode
fuego-dotclaude build ~/.claude        # isolated: just the user-scope .claude
fuego-dotclaude build ~/code/ai-repo   # isolated: repo IS the layout
```
